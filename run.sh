#!/data/data/com.termux/files/usr/bin/bash

# Ensure device stays awake
termux-wake-lock

# Base directories
HOME="/data/data/com.termux/files/home"
SCRIPTS_DIR="$HOME/scripts"

# Print system uptime
echo -e "\n📡 Termux boot started. Uptime: $(uptime)\n"

# Run all scripts in ~/scripts
if [ -d "$SCRIPTS_DIR" ]; then
    for script in "$SCRIPTS_DIR"/*.sh; do
        [ -f "$script" ] || continue
        if [[ "$script" == *"_nohup.sh" ]]; then
            echo "▶ Running $script in background ..."
            nohup bash "$script" > /dev/null 2>&1 &
        else
            echo "▶ Running $script ..."
            bash "$script"
        fi
        echo ""
    done
else
    echo "⚠ No scripts found in $SCRIPTS_DIR"
fi