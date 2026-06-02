#!/bin/bash
# Usage: ./scripts/db.sh "SELECT ..."
ssh -o BatchMode=yes root@89.127.216.143 \
  "docker exec workout-tracker-db-1 psql -U workout -d workout_tracker -c \"$*\""
