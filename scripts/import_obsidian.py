#!/usr/bin/env python3
"""
Import workouts from Obsidian Training Journal markdown format.

Format expected:
  ## Тренировка N — Title | DD.MM.YYYY
  ### optional sub-header

  1. Exercise name — sets
  2. Exercise name — sets

Separated by --- lines.

Usage:
  python3 scripts/import_obsidian.py [--file path] [--client Dymko] [--dry-run] [--from-date YYYY-MM-DD]
"""

import re, sys, argparse, psycopg2
from datetime import date
from pathlib import Path

# shared parsing from import_journal
sys.path.insert(0, str(Path(__file__).parent))
from import_journal import split_groups, parse_group, find_user, DB_URL

JOURNAL_DIR = Path.home() / "Desktop/Second Brain/Journal/Training"

# ── markdown cleaning ─────────────────────────────────────────────────────────

def clean_sets(text):
    text = re.sub(r'\*\*(.+?)\*\*', r'\1', text)         # **bold** → text
    text = re.sub(r'\*\([^)]*\)\*', '', text)              # *(note)* → ''
    text = re.sub(
        r'[\U00010000-\U0010FFFF\U0001F300-\U0001FAFF☀-⛿✀-➿]',
        '', text
    )
    text = re.sub(r'\*', '', text)
    text = re.sub(r'\([^)]*\)', '', text)                  # (parenthetical) → ''
    text = re.sub(r'\s{2,}', ' ', text)
    return text.strip().rstrip(',').strip()


def normalize_title(title):
    title = re.sub(r'Back Full\s*«Fit»', 'Back Full. В Fit', title)
    return title.strip()


# ── obsidian parser ───────────────────────────────────────────────────────────

HEADER_RE = re.compile(
    r'^## Тренировка\s+(\d+)\s+[—–]\s+(.+?)\s+\|\s+(\d{1,2}\.\d{2}\.\d{4})',
    re.MULTILINE
)


def parse_obsidian(content):
    workouts = []
    headers = list(HEADER_RE.finditer(content))

    for i, m in enumerate(headers):
        start = m.start()
        end = headers[i + 1].start() if i + 1 < len(headers) else len(content)
        block = content[start:end]

        num = int(m.group(1))
        title = normalize_title(m.group(2).strip())
        d, mo, y = m.group(3).split('.')
        workout_date = date(int(y), int(mo), int(d))

        exercises = []
        for line in block.splitlines():
            line = line.strip()
            ex = re.match(r'^(\d+)\.\s+(.+?)\s+[—–]\s+(.+)$', line)
            if not ex:
                continue
            name = ex.group(2).strip()
            sets_str = clean_sets(ex.group(3))

            parsed_sets = []
            for group in split_groups(sets_str):
                parsed_sets.extend(parse_group(group.strip()))

            exercises.append({
                'name': name,
                'sets': parsed_sets,
                'original': line,
            })

        workouts.append({
            'num': num,
            'title': title,
            'date': workout_date,
            'exercises': exercises,
        })

    return sorted(workouts, key=lambda w: w['date'])


# ── database insert ───────────────────────────────────────────────────────────

def insert_workouts(workouts, client_id, trainer_id, dry_run=False, from_date=None):
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    inserted = skipped = 0
    try:
        for wo in workouts:
            if from_date and wo['date'] < from_date:
                continue

            cur.execute(
                "SELECT id FROM workouts WHERE user_id=%s AND workout_date=%s AND title=%s",
                (client_id, wo['date'], wo['title'])
            )
            if cur.fetchone():
                print(f"  SKIP: {wo['date']} — {wo['title']}")
                skipped += 1
                continue

            if dry_run:
                print(f"  DRY #{wo['num']:2d}  {wo['date']} — {wo['title']}  ({len(wo['exercises'])} упр.)")
                for ex in wo['exercises']:
                    sets_str = ', '.join(
                        f"{'б/в' if w is None else w}×{r}" for w, r in ex['sets']
                    ) or '—'
                    print(f"           {ex['name']}: {sets_str}")
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

            print(f"  OK  #{wo['num']:2d}  {wo['date']} — {wo['title']}  ({len(wo['exercises'])} упр.)")
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


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--file', help='Path to .md journal file (default: all in JOURNAL_DIR)')
    parser.add_argument('--client', default='Dymko')
    parser.add_argument('--trainer', default='admin')
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--from-date', help='Import only workouts on or after YYYY-MM-DD')
    args = parser.parse_args()

    if args.file:
        files = [Path(args.file).expanduser()]
    else:
        files = sorted(JOURNAL_DIR.glob('20*.md'))

    if not files:
        print(f"Нет файлов в {JOURNAL_DIR}")
        sys.exit(1)

    all_workouts = []
    for f in files:
        content = f.read_text(encoding='utf-8')
        wos = parse_obsidian(content)
        print(f"  {f.name}: {len(wos)} тренировок")
        all_workouts.extend(wos)

    all_workouts.sort(key=lambda w: w['date'])
    print(f"\nИтого найдено: {len(all_workouts)}\n")

    from_date = None
    if args.from_date:
        from_date = date.fromisoformat(args.from_date)

    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    client_id, clogin, cname = find_user(cur, args.client)
    trainer_id, tlogin, tname = find_user(cur, args.trainer)
    cur.close()
    conn.close()

    if not client_id:
        print(f"Клиент '{args.client}' не найден в БД")
        sys.exit(1)

    print(f"Клиент:  {cname} ({clogin})")
    print(f"Тренер:  {tname} ({tlogin})")
    print(f"Режим:   {'dry-run' if args.dry_run else 'INSERT'}\n")

    insert_workouts(all_workouts, client_id, trainer_id,
                    dry_run=args.dry_run, from_date=from_date)


if __name__ == '__main__':
    main()
