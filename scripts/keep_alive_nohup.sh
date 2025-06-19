#!/data/data/com.termux/files/usr/bin/bash

# Log file
log_file="/data/data/com.termux/files/home/scripts/keep_alive.log"

echo "$(date +'%F %T') Starting keep-alive loop ..." >> "$log_file"

while true; do
    # Ensure Termux stays in foreground (optional)
    if ! pidof com.termux > /dev/null; then
        am start -n com.termux/.HomeActivity
    fi
    sleep 10
done