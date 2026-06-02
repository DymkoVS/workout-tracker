#!/bin/bash
# Fetch workouts.json from server, merge new workouts into Obsidian journal.
# Called by launchd every hour (com.vladimir.workout-obsidian-sync).

SERVER="root@89.127.216.143"
REMOTE_JSON="/opt/workout-obsidian/workouts.json"
LOCAL_JSON="/tmp/workout_obsidian_data.json"
SCRIPT="$HOME/Developer/workout-tracker/scripts/export_obsidian.py"
PYTHON="/Library/Frameworks/Python.framework/Versions/3.12/bin/python3"
LOG="$HOME/Developer/workout-tracker/logs/obsidian_sync.log"

log() { echo "$(date '+%Y-%m-%d %H:%M:%S') $*" >> "$LOG"; }

log "─── Синхронизация ───"

# Забрать JSON с сервера
if ! rsync -az \
    -e "ssh -o ConnectTimeout=10 -o BatchMode=yes" \
    "$SERVER:$REMOTE_JSON" "$LOCAL_JSON" >> "$LOG" 2>&1; then
    log "⚠️  Сервер недоступен, пропускаю"
    exit 0
fi

# Смержить в Obsidian
"$PYTHON" "$SCRIPT" --from-json "$LOCAL_JSON" >> "$LOG" 2>&1
STATUS=$?

if [ $STATUS -eq 0 ]; then
    log "✅ Готово"
else
    log "❌ Ошибка (код $STATUS)"
fi
