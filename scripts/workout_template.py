#!/usr/bin/env python3
"""
workout_template.py — шаблон тренировки из журнала Obsidian.

Читает: ~/Desktop/Second Brain/Journal/Training/20*.md

Использование:
  python3 scripts/workout_template.py legs      # Legs Full
  python3 scripts/workout_template.py upper     # Грудь / Дельты / Трицепс (+ Upper Body)
  python3 scripts/workout_template.py back      # Back Full (TopGym)
  python3 scripts/workout_template.py backfit   # Back Full «Fit» (FitGround)
  python3 scripts/workout_template.py zadn      # Задняя цепь
  python3 scripts/workout_template.py list      # все тренировки
"""

import re
import sys
from pathlib import Path
from datetime import date

JOURNAL_DIR = Path.home() / "Desktop/Second Brain/Journal/Training"

SPLIT_NAMES = {
    "legs":    "🦵  Legs Full",
    "upper":   "💪  Грудь / Дельты / Трицепс",
    "back":    "🔙  Back Full  (TopGym)",
    "backfit": "🔙  Back Full «Fit»  (FitGround)",
    "zadn":    "⛓️   Задняя цепь",
}

# Ключевое упражнение для отслеживания прогресса по каждому типу
KEY_EXERCISE = {
    "legs":    "жим платформы",
    "upper":   "жим штанги лёжа",
    "back":    "тяга т-грифа",
    "backfit": "тяга т-грифа",
    "zadn":    "мост",
}

LINE  = "═" * 64
LINE2 = "─" * 64


# ── утилиты ──────────────────────────────────────────────────────────────────

def clean_sets(text: str) -> str:
    """Убирает markdown-форматирование и emoji из строки подходов."""
    text = re.sub(r'\*\*(.+?)\*\*', r'\1', text)        # **bold** → text
    text = re.sub(r'\*\([^)]*\)\*', '', text)             # *(заметка)* → ''
    text = re.sub(r'[\U0001F000-\U0001FFFF]', '', text)   # emoji (🎯 и др.)
    text = re.sub(r'\*', '', text)                         # оставшиеся *
    return re.sub(r'\s{2,}', ' ', text).strip().rstrip(',').strip()


def matches_split(title: str, split_key: str) -> bool:
    """Проверяет, соответствует ли заголовок тренировки типу сплита."""
    t = title.lower()
    if split_key == "legs":
        return bool(re.search(r'legs\s+full', t))
    if split_key == "upper":
        # и старый «Upper Body», и новый «Грудь / Дельты / Трицепс»
        return bool(re.search(r'грудь', t)) or bool(re.search(r'upper\s*body', t))
    if split_key == "back":
        # Back Full без «Fit» → TopGym
        return bool(re.search(r'back\s+full', t)) and '«' not in title
    if split_key == "backfit":
        # Back Full «Fit» → FitGround
        return bool(re.search(r'back\s+full', t)) and '«' in title
    if split_key == "zadn":
        return bool(re.search(r'задняя\s+цепь', t))
    return False


def parse_date_str(s: str):
    m = re.match(r'(\d{1,2})\.(\d{2})\.(\d{4})', s.strip())
    return date(int(m.group(3)), int(m.group(2)), int(m.group(1))) if m else None


# ── парсинг журналов ─────────────────────────────────────────────────────────

def parse_journals():
    """
    Читает все файлы 20*.md из JOURNAL_DIR.
    Возвращает список тренировок, отсортированных по дате.

    Формат блока в markdown:
        ## Тренировка N — Title | DD.MM.YYYY
        ### опциональный подзаголовок

        1. Упражнение — подходы
        2. Упражнение — подходы
    """
    workouts = []
    for filepath in sorted(JOURNAL_DIR.glob("20*.md")):
        content = filepath.read_text(encoding='utf-8')

        # Разбиваем по горизонтальному разделителю ---
        blocks = re.split(r'\n\s*---\s*\n', content)

        for block in blocks:
            block = block.strip()
            if not block:
                continue

            # Заголовок тренировки
            m = re.match(
                r'^## Тренировка\s+(\d+)\s+[—–]\s+(.+?)\s+\|\s+(\d{1,2}\.\d{2}\.\d{4})',
                block
            )
            if not m:
                continue

            num          = int(m.group(1))
            title        = m.group(2).strip()
            workout_date = parse_date_str(m.group(3))
            if not workout_date:
                continue

            # Строки упражнений: "N. Название — подходы"
            exercises = []
            for line in block.splitlines():
                line = line.strip()
                ex = re.match(r'^(\d+)\.\s+(.+?)\s+[—–]\s+(.+)$', line)
                if ex:
                    exercises.append({
                        'num':  int(ex.group(1)),
                        'name': ex.group(2).strip(),
                        'sets': clean_sets(ex.group(3)),
                    })

            if exercises:
                workouts.append({
                    'num':      num,
                    'title':    title,
                    'date':     workout_date,
                    'exercises': exercises,
                })

    return sorted(workouts, key=lambda w: (w['date'], w['num']))


# ── вывод ─────────────────────────────────────────────────────────────────────

def print_template(workout, split_key, matching):
    """Выводит шаблон последней тренировки + историю ключевого упражнения."""
    date_str = workout['date'].strftime('%d.%m.%Y')
    total = len(matching)

    print()
    print(LINE)
    print(f"  {SPLIT_NAMES[split_key]}")
    print(f"  📅 {date_str}  —  Тренировка #{workout['num']}   ({total} всего этого типа)")
    print(LINE)
    print()

    # Упражнения с выравниванием
    max_name_len = max(len(ex['name']) for ex in workout['exercises'])
    for ex in workout['exercises']:
        padding = max_name_len - len(ex['name']) + 3
        print(f"  {ex['num']}. {ex['name']}" + " " * padding + ex['sets'])

    # История ключевого упражнения
    key = KEY_EXERCISE.get(split_key, '')
    if key:
        history_rows = []
        for h in matching:   # все тренировки этого типа
            for ex in h['exercises']:
                if key in ex['name'].lower():
                    marker = "  ← последняя" if h['num'] == workout['num'] else ""
                    history_rows.append(
                        f"   {h['date'].strftime('%d.%m')}  #{h['num']:2d}   {ex['sets']}{marker}"
                    )
                    break

        if history_rows:
            print()
            print(LINE2)
            print(f"  📈 {key}:")
            # Показываем последние 6 записей
            for row in history_rows[-6:]:
                print(row)

    print()
    print(LINE)
    print()


def cmd_list(all_workouts):
    print(f"\n  Журнал тренировок  ({len(all_workouts)} всего)\n")
    print(f"  {'Дата':<13} {'#':>3}   Тип")
    print("  " + "─" * 50)
    for w in all_workouts:
        print(f"  {w['date'].strftime('%d.%m.%Y'):<13} #{w['num']:2d}   {w['title']}")
    print()


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    cmd = sys.argv[1].lower().strip() if len(sys.argv) > 1 else ''

    if not cmd:
        print(__doc__)
        sys.exit(0)

    all_workouts = parse_journals()

    if cmd == 'list':
        cmd_list(all_workouts)
        return

    if cmd not in SPLIT_NAMES:
        print(f"\n  Неизвестный тип: '{cmd}'")
        print(f"  Доступные: {', '.join(list(SPLIT_NAMES) + ['list'])}\n")
        sys.exit(1)

    matching = [w for w in all_workouts if matches_split(w['title'], cmd)]

    if not matching:
        print(f"\n  Тренировок типа '{cmd}' не найдено.\n")
        sys.exit(1)

    print_template(matching[-1], cmd, matching)


if __name__ == '__main__':
    main()
