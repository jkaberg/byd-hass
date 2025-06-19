# Keep awake
termux-wake-lock

# Parameter passed in
param=$1

# In some environments, $HOME is unavailable, set it directly here
home="/data/data/com.termux/files/home"
termux_dir="/storage/emulated/0/Termux"
if [ ! -d "$termux_dir" ]; then
    mkdir -p "$termux_dir"
fi

# Show system uptime
status=$(uptime)
echo -e "\nTermux status: $status\n"

# =========================Below is the Termux keep-alive script===============================

# Install android-tools (includes adb needed for keep-alive script)
if ! command -v adb > /dev/null; then
    echo -e "\nInstalling android-tools ..."
    pkg install android-tools -y
    adb --version
    echo "Press any key to activate adb. If the system asks for USB debugging permission, please click 'Always allow' ..."
    read -n 1 -s
    adb devices
    echo -e "android-tools installed successfully!\n"
fi

# Check if the Termux keep-alive script exists; if not, create it
script_name="keep_alive.sh"
#script_path="$termux_dir/$script_name" # check why this doesnt exist.
script_path="$home/$script_name"
if [ ! -f "$script_path" ]; then
    cat << 'EOF' > "$script_path"
#!/bin/bash
echo $(date +"%Y-%m-%d %H:%M:%S") > $(dirname "$0")"/keep_alive.log"

while true; do
    if ! pgrep -x "com.termux" > /dev/null; then
        am start -n com.termux/.HomeActivity
    fi
    sleep 10
done
EOF
fi

# Start Termux keep-alive script
adb_server="localhost:5555"
adb connect $adb_server
process_count=$(adb -s $adb_server shell "pgrep -f $script_name" | wc -l)
if [ $process_count -gt 1 ]; then
    log_file="$termux_dir/keep_alive.log"
    start_time=$(cat $log_file)
    echo "Termux keep-alive script ($script_name) started at $start_time and is still running."
elif [ -f "$script_path" ]; then
    echo "Starting keep-alive script $script_path ..."
    adb -s $adb_server shell "nohup sh $script_path > /dev/null 2>&1 &"
fi
if [ "$param" != "install" ]; then
    sleep 2
    adb -s $adb_server shell "input keyevent KEYCODE_HOME" # Return to home screen
fi
adb disconnect $adb_server

# =========================Above is the Termux keep-alive script===============================

# Run all scripts in the boot directory
script_dir="$home/scripts"
for script in "$script_dir"/*.sh; do
    # Scripts ending in _nohup.sh will run in the background
    if [[ "$script" == *"_nohup.sh" ]]; then
        echo "Running $script in background ..."
        nohup bash "$script" > /dev/null 2>&1 &
    else
        echo "Running $script ..."
        bash "$script"
    fi
    echo ""
done
