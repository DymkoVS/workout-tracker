#!/usr/bin/env python3
"""
Import workout journal (text format) into workout-tracker PostgreSQL database.

Usage:
  python3 scripts/import_journal.py --file path/to/journal.txt --client Владимир [--dry-run]
"""

import re
import sys
import argparse
import psycopg2
from datetime import date

DB_URL = "postgresql://workout:workout_secret@localhost:5432/workout_tracker"

FILE_PATH = (
    "/Users/vladimir/Library/Mobile Documents/"
    "com~apple~CloudDocs/Тренировки/"
    "Журнал_тренировок_февраль_апрель_2026.txt"
)

MONTHS = {
    'января': 1, 'февраля': 2, 'марта': 3, 'апреля': 4,
    'мая': 5, 'июня': 6, 'июля': 7, 'августа': 8,
    'сентября': 9, 'октября': 10, 'ноября': 11, 'декабря': 12,
}

# ─── text normalization ────────────────────────────────────────────────────────

def strip_emojis(text):
    return re.sub(
        r'[\U00010000-\U0010FFFF\U0001F300-\U0001FAFF☀-⛿✀-➿]',
        '', text
    )

def normalize(text):
    text = strip_emojis(text)
    text = text.replace('×', '×').replace('✕', '×')
    # "по по10" → "по 10" (double-по typo)
    text = re.sub(r'по\s+по\s*(\d)', r'по \1', text)
    # "по по N" (with space before N)
    text = re.sub(r'по\s+по\s+', 'по ', text)
    # "по 1ой/по 2ой" attachment selector — not a set notation, strip it
    text = re.sub(r'по\s+\d+(?:ой|ей|ий)\s+', '', text, flags=re.IGNORECASE)
    return text

# ─── date parsing ─────────────────────────────────────────────────────────────

def parse_date(text):
    text = text.strip().rstrip('.')
    m = re.match(r'(\d{1,2})\s+(\S+)\s+(\d{4})', text)
    if m:
        month = MONTHS.get(m.group(2).lower())
        if month:
            return date(int(m.group(3)), month, int(m.group(1)))
    return None

# ─── set spec parsing ─────────────────────────────────────────────────────────

def split_groups(text):
    """Split 'W×R, W×R×N' into groups by comma.
    A comma is treated as a decimal separator ONLY if the very next character
    (no spaces) is a digit — e.g. '7,5' stays together, '15, 10' splits.
    """
    groups, current = [], ""
    for i, c in enumerate(text):
        if c == ',':
            nxt = text[i + 1] if i + 1 < len(text) else ''
            if nxt.isdigit():       # decimal comma: "7,5"
                current += c
            else:
                if current.strip():
                    groups.append(current.strip())
                current = ""
        else:
            current += c
    if current.strip():
        groups.append(current.strip())
    return groups


def parse_reps_str(s):
    """'12', '12-15', '8-10' → take the first (lower) number."""
    m = re.match(r'(\d+)', s.strip())
    return int(m.group(1)) if m else None


def expand(rest, weight):
    """
    Parse 'reps[×sets]' or 'reps1/reps2/reps3' → [(weight, reps), ...]
    """
    rest = rest.strip()
    rest = re.sub(r'^×\s*', '', rest)   # strip leading × (e.g. "пустая ×15" → "15")
    rest = rest.strip().rstrip('.')
    # strip trailing failure/technique notes
    rest = re.sub(
        r'\s*(в\s+\S+\s*отказ|отказ|\d+[а-яё]+\s+в\s+отказ|всё\s+в\s+отказ)\s*$',
        '', rest, flags=re.IGNORECASE
    ).strip()
    # strip trailing prose after ". " (e.g. "10×2. Прекрасно и в отказ!")
    rest = re.sub(r'\s*\.\s+.*$', '', rest).strip()
    # strip remaining trailing Cyrillic words (e.g. "10×3 оба в жёсткий")
    rest = re.sub(r'\s+[а-яёА-ЯЁ].+$', '', rest).strip()
    rest = rest.rstrip('.')
    if not rest:
        return []

    # slash notation: "10/15/12" or "15/12×2"
    if '/' in rest:
        result = []
        for part in rest.split('/'):
            part = part.strip()
            mx = re.match(r'^(\d+(?:-\d+)?)\s*×\s*(\d+)$', part)
            if mx:
                r = parse_reps_str(mx.group(1))
                n = int(mx.group(2))
                if r:
                    result.extend([(weight, r)] * n)
            elif re.match(r'^\d', part):
                r = parse_reps_str(part)
                if r:
                    result.append((weight, r))
        return result

    # "reps×sets"
    mx = re.match(r'^(\d+(?:-\d+)?)\s*×\s*(\d+)\s*$', rest)
    if mx:
        r, n = parse_reps_str(mx.group(1)), int(mx.group(2))
        return [(weight, r)] * n if r else []

    # just "reps"
    mx = re.match(r'^(\d+(?:-\d+)?)\s*$', rest)
    if mx:
        r = parse_reps_str(mx.group(1))
        return [(weight, r)] if r else []

    return []


def parse_group(group):
    """
    Parse one comma-group into list of (weight_or_None, reps).

    Handled patterns:
      N×R, N×R×S, +N×R, б/в×R×S, по N×R, по N+руб×R,
      N к×R×S, пустая×R, кривая×R, по рублю×R,
      дроп сет по N×R/по N×R
    """
    g = group.strip()
    # strip trailing technique notes
    g = re.sub(
        r'\s*(оба\s+в\s+\S+|обе\s+в\s+\S+|в\s+\S+\s+отказ|в отказ|отказ|дроп\s*сет)\s*$',
        '', g, flags=re.IGNORECASE
    ).strip()
    g = re.sub(r'\s*\.\s+.*$', '', g).strip()
    if not g:
        return []

    # drop set: "дроп сет по N×R/по N×R"
    dm = re.match(r'^дроп\s*сет\s+(.+)$', g, re.IGNORECASE)
    if dm:
        result = []
        for part in dm.group(1).split('/'):
            part = part.strip()
            pm = re.match(r'^по\s+(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$', part)
            if pm:
                w = float(pm.group(1).replace(',', '.'))
                result.extend(expand(pm.group(2), w))
            else:
                result.extend(_simple(part))
        return result

    # "по рублю [×R[×S]]"
    if re.search(r'по\s+рублю', g, re.IGNORECASE):
        rest = re.sub(r'по\s+рублю', '', g, flags=re.IGNORECASE).lstrip('×').strip()
        return expand(rest, None)

    # "пустая [×R]" or "кривая пустая [×R]"
    if re.search(r'(кривая\s+)?пустая', g, re.IGNORECASE):
        rest = re.sub(r'(кривая\s+)?пустая', '', g, flags=re.IGNORECASE).lstrip('×').strip()
        return expand(rest, None)

    # "б/в×R[×S]"
    bm = re.match(r'^б/в\s*×?\s*(.+)$', g)
    if bm:
        return expand(bm.group(1), None)

    # "+N плюх[а/и] [×R[×S]]" — plate count notation, weight unknown → None
    pm = re.match(r'^\+\d+\s+плюх(?:а|и)?\s*(.*)', g, re.IGNORECASE)
    if pm:
        return expand(pm.group(1), None)

    # "+N ×R[×S]"  (added machine weight)
    pm = re.match(r'^\+(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$', g)
    if pm:
        return expand(pm.group(2), float(pm.group(1).replace(',', '.')))

    # "по N+руб×R[×S]"  e.g. "по 2,5+руб×15"
    pm = re.match(r'^по\s+(\d+(?:[,\.]\d+)?)\+(?:руб|рублю)\s*×\s*(.+)$', g)
    if pm:
        return expand(pm.group(2), float(pm.group(1).replace(',', '.')))

    # "по N×R[×S]"  (dumbbell per-side weight)
    pm = re.match(r'^по\s+(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$', g)
    if pm:
        return expand(pm.group(2), float(pm.group(1).replace(',', '.')))

    # "N к×R[×S]"  e.g. "2 к×12×3"
    km = re.match(r'^(\d+(?:[,\.]\d+)?)\s+к\s*×\s*(.+)$', g)
    if km:
        return expand(km.group(2), float(km.group(1).replace(',', '.')))

    return _simple(g)


def _simple(g):
    """Standard 'N×R[×S]'."""
    g = g.strip()
    mx = re.match(r'^(\d+(?:[,\.]\d+)?)\s*×\s*(.+)$', g)
    if mx:
        return expand(mx.group(2), float(mx.group(1).replace(',', '.')))
    # bare "×R[×S]"
    mx = re.match(r'^×\s*(.+)$', g)
    if mx:
        return expand(mx.group(1), None)
    return []


# ─── exercise line parsing ────────────────────────────────────────────────────

# Pattern marking the START of the set spec part of an exercise line
SET_START_RE = re.compile(
    r'('
    r'\d+[,.]?\d*\s*×'              # N×
    r'|\d+\s+к\s*×'                  # N к×
    r'|\+\d'                          # +N
    r'|б/в'                           # bodyweight
    r'|по\s+(?:\d|рублю|по\s*\d)'    # по N / по рублю / по по N
    r'|пустая'                        # empty bar
    r'|кривая\s+пустая'              # curved empty bar
    r'|дроп\s*сет'                   # drop set
    r')'
)


def parse_exercise_line(raw_line):
    """
    Parse '1. Exercise name 27×15, 32×12×2' →
    (name: str, sets: [(weight, reps), ...], original_text: str)
    """
    line = normalize(raw_line)
    line = re.sub(r'^\d+\.\s*', '', line).strip()  # remove "1. " prefix

    m = SET_START_RE.search(line)
    if not m:
        return line.strip().rstrip('.'), [], raw_line.strip()

    name = line[:m.start()].strip().rstrip('.')
    # strip inline prose note from exercise name (e.g. ". Здесь не хуярим, надо включить.")
    name = re.sub(r'\.\s+[А-ЯЁа-яё].+$', '', name).strip()
    sets_text = line[m.start():]

    # If there's an inline date reference "NN.MM. " — use only data after it
    inline_date = re.search(r'\.\s+\d{2}\.\d{2}\.\s+', sets_text)
    if inline_date:
        sets_text = sets_text[inline_date.end():]

    # If there's an inline Russian note sentence "Сегодня жим в блоке... NN×R"
    # (trainer switching to a different exercise mid-record), keep only the latter part
    inline_note = re.search(r'\.\s+[А-ЯЁа-яё][^×]{5,}\.\s+', sets_text)
    if inline_note:
        after = sets_text[inline_note.end():]
        if re.search(r'\d.*×', after):
            sets_text = after

    all_sets = []
    for group in split_groups(sets_text):
        all_sets.extend(parse_group(group.strip()))

    return name, all_sets, raw_line.strip()


# ─── workout block parsing ────────────────────────────────────────────────────

def _parse_block_old(lines):
    """Old format: 'Тренировка N. Title.' on line 0, date on line 1."""
    title_m = re.match(r'Тренировка\s+\d+\.\s*(.+?)\.?\s*$', lines[0])
    if not title_m:
        return None
    title = title_m.group(1).strip()
    workout_date = None
    ex_start = 2
    for idx in range(1, len(lines)):
        d = parse_date(lines[idx])
        if d:
            workout_date = d
            ex_start = idx + 1
            break
    if not workout_date:
        return None
    return title, workout_date, ex_start


def _parse_block_new(lines):
    """New format: 'Тренировка DD.MM.YYYY. Title.' on line 0."""
    m = re.match(r'Тренировка\s+(\d{2}\.\d{2}\.\d{4})\.\s*(.+?)\.?\s*$', lines[0])
    if not m:
        return None
    try:
        d, mo, y = m.group(1).split('.')
        workout_date = date(int(y), int(mo), int(d))
    except ValueError:
        return None
    title = m.group(2).strip()
    return title, workout_date, 1


def parse_workouts(content):
    workouts = []

    # Choose splitting strategy based on presence of --- separators
    if re.search(r'\n\s*---+\s*\n', content):
        blocks = re.split(r'\n\s*---+\s*\n', content)
        parse_header = _parse_block_old
    else:
        blocks = re.split(r'\n{2,}', content)
        parse_header = _parse_block_new

    for block in blocks:
        block = block.strip()
        if not block:
            continue
        lines = [l.strip() for l in block.splitlines() if l.strip()]
        if len(lines) < 2:
            continue

        result = parse_header(lines)
        if not result:
            continue
        title, workout_date, ex_start = result

        exercises = []
        for line in lines[ex_start:]:
            if re.match(r'^\d+\.', line):
                name, sets, original = parse_exercise_line(line)
                if name:
                    exercises.append({'name': name, 'sets': sets, 'original': original})

        workouts.append({'title': title, 'date': workout_date, 'exercises': exercises})

    return workouts


# ─── database operations ──────────────────────────────────────────────────────

def find_user(cur, name_or_login):
    cur.execute(
        "SELECT id, login, full_name FROM users WHERE login ILIKE %s OR full_name ILIKE %s",
        (name_or_login, name_or_login)
    )
    row = cur.fetchone()
    if row:
        return row[0], row[1], row[2]
    return None, None, None


def insert_workouts(workouts, client_id, trainer_id, dry_run=False):
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    try:
        inserted = 0
        skipped = 0
        for wo in workouts:
            # Skip if workout already exists for this user on this date
            cur.execute(
                "SELECT id FROM workouts WHERE user_id=%s AND workout_date=%s AND title=%s",
                (client_id, wo['date'], wo['title'])
            )
            if cur.fetchone():
                print(f"  SKIP (exists): {wo['date']} — {wo['title']}")
                skipped += 1
                continue

            if dry_run:
                print(f"  DRY: {wo['date']} — {wo['title']}  ({len(wo['exercises'])} упражнений)")
                for ex in wo['exercises']:
                    sets_str = ', '.join(
                        f"{w if w is not None else 'б/в'}×{r}"
                        for w, r in ex['sets']
                    ) or '—'
                    print(f"       {ex['name']}: {sets_str}")
                inserted += 1
                continue

            cur.execute(
                """INSERT INTO workouts (user_id, trainer_id, title, workout_date, notes)
                   VALUES (%s, %s, %s, %s, '') RETURNING id""",
                (client_id, trainer_id, wo['title'], wo['date'])
            )
            workout_id = cur.fetchone()[0]

            for i, ex in enumerate(wo['exercises']):
                cur.execute(
                    """INSERT INTO workout_exercises (workout_id, name, order_num, notes)
                       VALUES (%s, %s, %s, %s) RETURNING id""",
                    (workout_id, ex['name'], i + 1, ex['original'])
                )
                ex_id = cur.fetchone()[0]

                for j, (weight, reps) in enumerate(ex['sets']):
                    cur.execute(
                        """INSERT INTO sets (workout_exercise_id, set_num, weight, reps)
                           VALUES (%s, %s, %s, %s)""",
                        (ex_id, j + 1, weight, reps)
                    )

            print(f"  OK:  {wo['date']} — {wo['title']}  ({len(wo['exercises'])} упражнений)")
            inserted += 1

        if not dry_run:
            conn.commit()
        print(f"\nВсего: вставлено {inserted}, пропущено {skipped}")
    except Exception as e:
        conn.rollback()
        print(f"Ошибка: {e}")
        raise
    finally:
        cur.close()
        conn.close()


# ─── main ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--file', default=FILE_PATH)
    parser.add_argument('--client', default='Владимир')
    parser.add_argument('--trainer', default='admin')
    parser.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()

    with open(args.file, encoding='utf-8-sig') as f:
        content = f.read()

    workouts = parse_workouts(content)
    print(f"Распознано тренировок: {len(workouts)}\n")

    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    client_id, clogin, cname = find_user(cur, args.client)
    trainer_id, tlogin, tname = find_user(cur, args.trainer)
    cur.close()
    conn.close()

    if not client_id:
        print(f"Клиент '{args.client}' не найден в БД")
        sys.exit(1)

    print(f"Клиент:  {cname} ({clogin})  [{client_id}]")
    print(f"Тренер:  {tname} ({tlogin})  [{trainer_id}]")
    print(f"Режим:   {'dry-run' if args.dry_run else 'INSERT'}\n")

    insert_workouts(workouts, client_id, trainer_id, dry_run=args.dry_run)


if __name__ == '__main__':
    main()
