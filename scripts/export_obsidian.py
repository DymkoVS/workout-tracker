#!/usr/bin/env python3
"""
Export new completed workouts to Obsidian journal.

Two modes:
  --from-json FILE  Read workouts from pre-fetched JSON (server-side data)
  (default)         Query DB directly via SSH (requires server to be reachable)

Usage:
  python3 scripts/export_obsidian.py                        # SSH mode
  python3 scripts/export_obsidian.py --from-json /tmp/w.json
  python3 scripts/export_obsidian.py --dry-run
"""

import subprocess
import re
import sys
import json
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


# ── DB access (SSH mode) ──────────────────────────────────────────────────────

def run_query(sql):
    cmd = f'docker exec {DB_CONTAINER} psql -U {DB_USER} -d {DB_NAME} -t -A -F"\t" -c "{sql}"'
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
            rows.append(line.split("\t"))
    return rows


def fetch_workouts_via_ssh():
    rows = run_query(
        "SELECT w.id, w.title, w.workout_date::text, COALESCE(g.name, ''), "
        "COALESCE(we.name, ''), COALESCE(we.order_num::text, ''), "
        "COALESCE(TRIM(TRAILING '0' FROM TRIM(TRAILING '.' FROM s.weight::text)), ''), "
        "COALESCE(s.reps::text, '0') "
        "FROM workouts w "
        "LEFT JOIN gyms g ON g.id = w.gym_id "
        "JOIN users u ON u.id = w.user_id "
        "LEFT JOIN workout_exercises we ON we.workout_id = w.id "
        "LEFT JOIN sets s ON s.workout_exercise_id = we.id "
        f"WHERE u.login = '{CLIENT_LOGIN}' AND w.ended_at IS NOT NULL "
        "ORDER BY w.workout_date, w.id, we.order_num, s.set_num"
    )

    workouts = []
    current_wid = None
    exercises_map = {}

    for row in rows:
        wid, title, wdate, gym, ex_name, ex_order, weight, reps_str = row
        if wid != current_wid:
            if current_wid is not None:
                workouts[-1]["exercises"] = [exercises_map[k] for k in sorted(exercises_map)]
            workouts.append({"id": wid, "title": title, "date": wdate, "gym": gym, "exercises": []})
            current_wid = wid
            exercises_map = {}
        if ex_order:
            order = int(ex_order)
            name = ex_name.replace("×", "х")
            if order not in exercises_map:
                exercises_map[order] = {"name": name, "sets": []}
            exercises_map[order]["sets"].append([weight, int(reps_str)])

    if workouts:
        workouts[-1]["exercises"] = [exercises_map[k] for k in sorted(exercises_map)]

    return workouts


# ── Obsidian file helpers ─────────────────────────────────────────────────────

def scan_obsidian_state() -> tuple:
    """Single pass over all journal files — returns (existing_dates, max_workout_num)."""
    existing = set()
    max_num = 0
    heading_re = re.compile(r"## Тренировка (\d+) — .+ \| (\d{2})\.(\d{2})\.(\d{4})")
    for f in JOURNAL_DIR.glob("20*.md"):
        for line in f.read_text(encoding="utf-8").splitlines():
            m = heading_re.match(line)
            if m:
                num, d, mo, y = m.groups()
                max_num = max(max_num, int(num))
                existing.add(date(int(y), int(mo), int(d)))
    return existing, max_num


def get_journal_file(wo_date):
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


# ── Formatting ────────────────────────────────────────────────────────────────

def format_sets(sets):
    parts = []
    i = 0
    while i < len(sets):
        weight, reps = sets[i]
        count = 1
        while i + count < len(sets) and sets[i + count] == [weight, reps]:
            count += 1
        w = weight if weight else "б/в"
        parts.append(f"{w}×{reps}×{count}" if count > 1 else f"{w}×{reps}")
        i += count
    return ", ".join(parts)


def format_block(num, wo, wo_date):
    date_str = f"{wo_date.day:02d}.{wo_date.month:02d}.{wo_date.year}"
    lines = [f"## Тренировка {num} — {wo['title']} | {date_str}", ""]
    for i, ex in enumerate(wo["exercises"], 1):
        lines.append(f"{i}. {ex['name']} — {format_sets(ex['sets'])}")
    lines += ["", "---", ""]
    return "\n".join(lines)


# ── Main ──────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--from-json", metavar="FILE",
                        help="Read workouts from JSON file instead of querying DB via SSH")
    args = parser.parse_args()

    print("workout-tracker → Obsidian sync")

    if args.from_json:
        all_workouts = json.loads(Path(args.from_json).read_text(encoding="utf-8"))
        print(f"Источник: {args.from_json} ({len(all_workouts)} записей)")
    else:
        all_workouts = fetch_workouts_via_ssh()

    existing_dates, max_num = scan_obsidian_state()
    new_workouts = [w for w in all_workouts
                    if date.fromisoformat(w["date"]) not in existing_dates]

    print(f"Obsidian: {len(existing_dates)} | БД: {len(all_workouts)} | Новых: {len(new_workouts)}")

    if not new_workouts:
        print("Нечего добавлять.")
        return

    next_num = max_num + 1
    file_counts = {}

    for wo in new_workouts:
        wo_date = date.fromisoformat(wo["date"])
        block = format_block(next_num, wo, wo_date)
        fpath = get_journal_file(wo_date)

        key = (wo_date.year, wo_date.month)
        if key not in file_counts:
            file_counts[key] = count_entries_in_file(fpath)

        if args.dry_run:
            print(f"\n--- DRY RUN: #{next_num} → {fpath.name} ---")
            print(block)
        else:
            file_counts[key] += 1
            append_to_file(fpath, block)
            update_month_count(fpath, file_counts[key])
            print(f"  ✓ #{next_num} — {wo['title']} | {wo_date}")

        next_num += 1

    if not args.dry_run:
        print("Готово.")


if __name__ == "__main__":
    main()
