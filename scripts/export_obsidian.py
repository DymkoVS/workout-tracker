#!/usr/bin/env python3
"""
Export new completed workouts from workout-tracker DB to Obsidian journal.

Usage:
  python3 scripts/export_obsidian.py           # sync new workouts
  python3 scripts/export_obsidian.py --dry-run # show what would be added
"""

import subprocess
import re
import sys
import argparse
from datetime import date
from pathlib import Path

JOURNAL_DIR = Path.home() / "Desktop/Second Brain/Journal/Training"
SSH_HOST = "root@89.127.216.143"
DB_CONTAINER = "workout-tracker-db-1"
DB_USER = "workout"
DB_NAME = "workout_tracker"
CLIENT_LOGIN = "Dymko"

MONTHS_RU = {
    1: "Январь", 2: "Февраль", 3: "Март", 4: "Апрель",
    5: "Май", 6: "Июнь", 7: "Июль", 8: "Август",
    9: "Сентябрь", 10: "Октябрь", 11: "Ноябрь", 12: "Декабрь",
}


def run_query(sql):
    """Run SQL on production DB via SSH. Returns list of rows."""
    cmd = f'docker exec {DB_CONTAINER} psql -U {DB_USER} -d {DB_NAME} -t -A -F"|" -c "{sql}"'
    result = subprocess.run(
        ["ssh", "-o", "ConnectTimeout=15", SSH_HOST, cmd],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(f"SSH/psql failed: {result.stderr.strip()}")
    rows = []
    for line in result.stdout.strip().splitlines():
        line = line.strip()
        if line and not line.startswith("("):
            rows.append(line.split("|"))
    return rows


def get_existing_dates():
    """Parse all journal files and return set of workout dates already recorded."""
    existing = set()
    date_re = re.compile(r"## Тренировка \d+ — .+ \| (\d{2})\.(\d{2})\.(\d{4})")
    for f in JOURNAL_DIR.glob("20*.md"):
        for line in f.read_text(encoding="utf-8").splitlines():
            m = date_re.match(line)
            if m:
                d, mo, y = m.groups()
                existing.add(date(int(y), int(mo), int(d)))
    return existing


def get_max_workout_num():
    """Find the highest тренировка number across all journal files."""
    max_num = 0
    num_re = re.compile(r"## Тренировка (\d+) —")
    for f in JOURNAL_DIR.glob("20*.md"):
        for line in f.read_text(encoding="utf-8").splitlines():
            m = num_re.match(line)
            if m:
                max_num = max(max_num, int(m.group(1)))
    return max_num


def get_db_workouts():
    """Fetch all completed workouts for CLIENT_LOGIN, ordered by date."""
    sql = (
        "SELECT w.id, w.title, w.workout_date, COALESCE(g.name, '') "
        "FROM workouts w "
        "LEFT JOIN gyms g ON g.id = w.gym_id "
        "JOIN users u ON u.id = w.user_id "
        f"WHERE u.login = '{CLIENT_LOGIN}' AND w.ended_at IS NOT NULL "
        "ORDER BY w.workout_date"
    )
    result = []
    for row in run_query(sql):
        wid, title, wdate_str, gym = row[0], row[1], row[2], row[3]
        result.append({"id": wid, "title": title, "date": date.fromisoformat(wdate_str), "gym": gym})
    return result


def get_exercises(workout_id):
    """Fetch exercises + sets for a workout, grouped by exercise order."""
    sql = (
        "SELECT we.name, we.order_num, s.set_num, "
        "COALESCE(TRIM(TRAILING '0' FROM TRIM(TRAILING '.' FROM weight::text)), ''), s.reps "
        "FROM workout_exercises we "
        "JOIN sets s ON s.workout_exercise_id = we.id "
        f"WHERE we.workout_id = '{workout_id}' "
        "ORDER BY we.order_num, s.set_num"
    )
    exercises = {}
    for row in run_query(sql):
        name, order_num, _, weight, reps = row[0], int(row[1]), int(row[2]), row[3], int(row[4])
        # Fix × (multiplication sign) mistakenly used instead of Cyrillic х in names
        name = name.replace("×", "х")
        if order_num not in exercises:
            exercises[order_num] = {"name": name, "sets": []}
        exercises[order_num]["sets"].append((weight, int(reps)))
    return [exercises[k] for k in sorted(exercises.keys())]


def format_sets(sets):
    """Collapse consecutive identical sets: (10, 8), (10, 8) → 10×8×2."""
    parts = []
    i = 0
    while i < len(sets):
        weight, reps = sets[i]
        count = 1
        while i + count < len(sets) and sets[i + count] == (weight, reps):
            count += 1
        w = weight if weight else "б/в"
        parts.append(f"{w}×{reps}×{count}" if count > 1 else f"{w}×{reps}")
        i += count
    return ", ".join(parts)


def format_block(num, wo, exercises):
    """Return Obsidian markdown block for a workout."""
    d = wo["date"]
    date_str = f"{d.day:02d}.{d.month:02d}.{d.year}"
    lines = [f"## Тренировка {num} — {wo['title']} | {date_str}", ""]
    for i, ex in enumerate(exercises, 1):
        lines.append(f"{i}. {ex['name']} — {format_sets(ex['sets'])}")
    lines += ["", "---", ""]
    return "\n".join(lines)


def get_journal_file(wo_date):
    """Return path to the month journal file, creating it if needed."""
    fpath = JOURNAL_DIR / f"{wo_date.year}-{wo_date.month:02d}.md"
    if not fpath.exists():
        month_name = MONTHS_RU[wo_date.month]
        fpath.write_text(
            f"# Тренировки — {month_name} {wo_date.year}\n\n"
            f"Тип сплита: **Legs / Грудь+Дельты+Трицепс / Back Full**  \n"
            f"Тренировок за месяц: **0**  \n\n---\n\n",
            encoding="utf-8",
        )
    return fpath


def count_entries_in_file(fpath):
    content = fpath.read_text(encoding="utf-8")
    return len(re.findall(r"^## Тренировка \d+", content, re.MULTILINE))


def update_month_count(fpath, count):
    content = fpath.read_text(encoding="utf-8")
    new = re.sub(
        r"Тренировок за месяц: \*\*\d+\*\*",
        f"Тренировок за месяц: **{count}**",
        content,
    )
    fpath.write_text(new, encoding="utf-8")


def append_to_file(fpath, text):
    content = fpath.read_text(encoding="utf-8").rstrip("\n")
    fpath.write_text(content + "\n\n" + text, encoding="utf-8")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--dry-run", action="store_true", help="Show what would be added, don't write")
    args = parser.parse_args()

    print("workout-tracker → Obsidian sync")

    existing_dates = get_existing_dates()
    db_workouts = get_db_workouts()
    new_workouts = [w for w in db_workouts if w["date"] not in existing_dates]

    print(f"Obsidian: {len(existing_dates)} | БД: {len(db_workouts)} | Новых: {len(new_workouts)}")

    if not new_workouts:
        print("Нечего добавлять.")
        return

    next_num = get_max_workout_num() + 1
    file_counts = {}

    for wo in new_workouts:
        exercises = get_exercises(wo["id"])
        block = format_block(next_num, wo, exercises)
        fpath = get_journal_file(wo["date"])

        key = (wo["date"].year, wo["date"].month)
        if key not in file_counts:
            file_counts[key] = count_entries_in_file(fpath)

        if args.dry_run:
            print(f"\n--- DRY RUN: #{next_num} → {fpath.name} ---")
            print(block)
        else:
            file_counts[key] += 1
            append_to_file(fpath, block)
            update_month_count(fpath, file_counts[key])
            print(f"  ✓ #{next_num} — {wo['title']} | {wo['date']}")

        next_num += 1

    if not args.dry_run:
        print("Готово.")


if __name__ == "__main__":
    main()
