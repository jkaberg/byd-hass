#!/data/data/com.termux/files/usr/bin/bash

# Ensure device stays awake
termux-wake-lock

# Base directories
HOME="/data/data/com.termux/files/home"
SCRIPTS_DIR="$HOME/scripts"
TERMUX_DIR="/storage/emulated/0/Termux" # path is absolute in terms to adb shell
KEEP_ALIVE_SCRIPT="keep_alive.sh"
ADB_SERVER="localhost:5555"

# Print system uptime
echo -e "\n📡 Termux boot started. Uptime: $(uptime)\n"

# Setup .bashrc
if [ ! -f "$HOME/.bashrc" ]; then
    touch "$HOME/.bashrc"
fi

if ! grep -Fxq "$HOME/.termux/boot/run.sh" "$HOME/.bashrc"; then
  echo "$HOME/.termux/boot/run.sh" >> "$HOME/.bashrc"
fi

# Setup ADB
adb connect $ADB_SERVER > /dev/null

if [ ! -d "$TERMUX_DIR" ]; then
    adb -s $ADB_SERVER shell "mkdir -p $TERMUX_DIR"
fi

# Check if our keep alive script is present, if not add it
script_path="$TERMUX_DIR/$KEEP_ALIVE_SCRIPT"
if [ ! -f "$script_path" ]; then
    cat << 'EOF' > "$script_path"
#!/bin/bash
echo "$(date +"%Y-%m-%d %H:%M:%S")" >> "/storage/emulated/0/Termux/keep_alive.log"

while true; do
    if ! pgrep -x "com.termux" > /dev/null; then
        am start -n com.termux/.HomeActivity
	sleep 2
	input keyevent KEYCODE_HOME
    fi
    sleep 10
done
EOF
fi

# Check if the keep alive script is running, if not start it with adb
process_count=$(adb -s $ADB_SERVER shell "pgrep -f $KEEP_ALIVE_SCRIPT" | wc -l)
if [ $process_count -gt 1 ]; then
    log_file="$TERMUX_DIR/keep_alive.log"
    start_time=$(cat $log_file)
    echo "Keep alive script is running since $start_time"
elif [ -f "$script_path" ]; then
    echo "Starting keep alive script $script_path"
    adb -s $ADB_SERVER shell "nohup sh $script_path > /dev/null 2>&1 &"
fi
echo 
adb disconnect $ADB_SERVER > /dev/null

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
