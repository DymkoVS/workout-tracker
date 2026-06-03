#!/usr/bin/env python3
"""
Runs on EU VPS (cron, every hour).
Exports completed Dymko workouts from PostgreSQL to /opt/workout-obsidian/workouts.json.

Deploy:
  cat scripts/server_export.py | ssh root@89.127.216.143 "cat > /opt/workout-obsidian/export.py"
  ssh root@89.127.216.143 "crontab -l | { cat; echo '0 * * * * python3 /opt/workout-obsidian/export.py >> /opt/workout-obsidian/export.log 2>&1'; } | crontab -"
"""

import subprocess
import json
from pathlib import Path

DB_CONTAINER = "workout-tracker-db-1"
DB_USER = "workout"
DB_NAME = "workout_tracker"
CLIENT_LOGIN = "Dymko"
OUTPUT = Path("/opt/workout-obsidian/workouts.json")


def psql(sql):
    r = subprocess.run(
        ["docker", "exec", DB_CONTAINER, "psql",
         "-U", DB_USER, "-d", DB_NAME, "-t", "-A", "-F\t", "-c", sql],
        capture_output=True, text=True,
    )
    if r.returncode != 0:
        raise RuntimeError(r.stderr.strip())
    rows = []
    for line in r.stdout.strip().splitlines():
        line = line.strip()
        if line and not line.startswith("("):
            rows.append(line.split("\t"))
    return rows


def main():
    rows = psql(
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

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    tmp = OUTPUT.with_suffix(".tmp")
    tmp.write_text(json.dumps(workouts, ensure_ascii=False, indent=2), encoding="utf-8")
    tmp.rename(OUTPUT)  # atomic on same filesystem
    print(f"Exported {len(workouts)} workouts → {OUTPUT}")


if __name__ == "__main__":
    main()
