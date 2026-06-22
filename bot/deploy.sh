#!/bin/bash
# Деплой workout-bot на EU VPS
# Использование: bash ~/Developer/workout-tracker/bot/deploy.sh

set -e
SERVER="root@89.127.216.143"
REMOTE="/opt/workout-bot"

echo "→ Копируем bot.py и requirements.txt..."
ssh "$SERVER" "mkdir -p $REMOTE/logs"
scp "$(dirname "$0")/bot.py"           "$SERVER:$REMOTE/bot.py"
scp "$(dirname "$0")/requirements.txt" "$SERVER:$REMOTE/requirements.txt"

echo "→ Устанавливаем зависимости (venv)..."
ssh "$SERVER" "
  cd $REMOTE
  python3 -m venv venv --upgrade-deps
  venv/bin/pip install -q -r requirements.txt
"

echo "→ Устанавливаем systemd-сервис..."
scp "$(dirname "$0")/workout-bot.service" "$SERVER:/etc/systemd/system/workout-bot.service"
ssh "$SERVER" "
  systemctl daemon-reload
  systemctl enable workout-bot
  systemctl restart workout-bot
  sleep 2
  systemctl status workout-bot --no-pager
"

echo "✅ Готово"
