#!/data/data/com.termux/files/usr/bin/bash

# BYD-HASS Bootstrapping Installation Script
# This script sets up a self-healing, two-tiered system.
# 1. A Termux:Boot script acts as the orchestrator.
# 2. The orchestrator starts an external keep-alive script via ADB.
# 3. The external script ensures the Termux app itself stays running.
# 4. The orchestrator then starts the main byd-hass binary.

#set -e

# Disable strict error handling for interactive and ADB setup steps
#set +e

# Re-attach stdin to the user's terminal when the script is executed through a pipe (e.g. curl | bash)
if [ ! -t 0 ] && [ -t 1 ] && [ -e /dev/tty ]; then
  exec < /dev/tty
fi

# --- Configuration ---
# Shared storage paths (primary location for binary & config)
BINARY_NAME="byd-hass"
SHARED_DIR="/storage/emulated/0/bydhass"
BINARY_PATH="$SHARED_DIR/$BINARY_NAME"
EXEC_PATH="/data/local/tmp/$BINARY_NAME"  # Execution location inside Android shell (exec allowed)
CONFIG_PATH="$SHARED_DIR/config.env"
LOG_FILE="$SHARED_DIR/byd-hass.log"

# Termux-local paths (used only for the Termux:Boot starter logs)
INSTALL_DIR="$HOME/.byd-hass"
INTERNAL_LOG_FILE="$INSTALL_DIR/starter.log"

# Termux:Boot script (The starter that keeps the external guardian alive)
BOOT_DIR="$HOME/.termux/boot"
BOOT_SCRIPT_NAME="byd-hass-starter.sh"
BOOT_GPS_SCRIPT_NAME="byd-hass-gpsdata.sh"
BOOT_SCRIPT_PATH="$BOOT_DIR/$BOOT_SCRIPT_NAME"
BOOT_GPS_SCRIPT_PATH="$BOOT_DIR/$BOOT_GPS_SCRIPT_NAME"

# External guardian (runs under Android's 'sh')
ADB_KEEPALIVE_SCRIPT_NAME="keep-alive.sh"
ADB_KEEPALIVE_SCRIPT_PATH="$SHARED_DIR/$ADB_KEEPALIVE_SCRIPT_NAME"
ADB_LOG_FILE="$SHARED_DIR/keep-alive.log"

# ADB connection to self
ADB_SERVER="localhost:5555"

# GitHub repo details
REPO="jkaberg/byd-hass"
ASSET_NAME="byd-hass-arm64"
RELEASES_API="https://api.github.com/repos/$REPO/releases/latest"

# Local temporary download path
TEMP_BINARY_PATH="/data/data/com.termux/files/usr/tmp/$BINARY_NAME"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper function: ensure ADB connection, then run the given command in the device shell
adbs() {
  # Establish connection if it is not already present
  if ! adb devices | grep -q "$ADB_SERVER"; then
    echo "Connecting to ADB $ADB_SERVER ..."
    adb connect "$ADB_SERVER" >/dev/null 2>&1 || return 1
  fi
  # Execute the requested command in the device shell
  adb -s "$ADB_SERVER" shell "$@"
}

# Comprehensive cleanup function
cleanup_all_processes() {
  echo -e "${YELLOW}Performing cleanup of all BYD-HASS processes...${NC}"
  
  # Kill Android-side processes via ADB
  echo "Stopping Android-side processes..."
  adbs "pkill -f $ADB_KEEPALIVE_SCRIPT_NAME" 2>/dev/null || true
  adbs "pkill -f $BINARY_NAME" 2>/dev/null || true
  adbs "pkill -f byd-hass" 2>/dev/null || true
  # Force kill any remaining processes by exact binary path
  adbs "pkill -f $EXEC_PATH" 2>/dev/null || true
  adbs "pkill -f $BINARY_PATH" 2>/dev/null || true
  
  # Kill Termux-side processes
  echo "Stopping Termux-side processes..."
  pkill -f "$BOOT_SCRIPT_NAME" 2>/dev/null || true
  pkill -f "$BOOT_GPS_SCRIPT_NAME" 2>/dev/null || true
  pkill -f "byd-hass-starter.sh" 2>/dev/null || true
  pkill -f "byd-hass-gpsdata.sh" 2>/dev/null || true
  pkill -f "$BINARY_NAME" 2>/dev/null || true
  pkill -f "byd-hass" 2>/dev/null || true
  
  # Wait a moment for processes to terminate
  sleep 2
  
  # Double-check and force kill any stubborn processes
  echo "Performing final cleanup..."
  adbs "ps | grep -E '(keep-alive|byd-hass|gpsdata)' | grep -v grep | awk '{print \$2}' | xargs -r kill -9" 2>/dev/null || true
  ps aux | grep -E '(byd-hass|keep-alive|gpsdata)' | grep -v grep | awk '{print $2}' | xargs -r kill -9 2>/dev/null || true
  
  echo "✅ All processes terminated."
}

# --- Script Start ---
echo -e "${GREEN}🚗 BYD-HASS Bootstrapping Installer${NC}"

# Handle cleanup-only mode
if [ "$1" = "cleanup" ]; then
  echo -e "\n${BLUE}Cleanup mode - stopping all BYD-HASS processes...${NC}"
  cleanup_all_processes
  echo -e "\n${GREEN}✅ Cleanup complete. All BYD-HASS processes have been terminated.${NC}"
  exit 0
fi

# 1. Setup Termux Environment
echo -e "\n${BLUE}1. Setting up Termux environment...${NC}"
echo "Installing dependencies (adb, curl, jq, termux-api)..."
pkg install -y android-tools curl jq termux-api bc >/dev/null 2>&1
echo "✅ Environment ready."

# 2. Explain Manual Steps
echo -e "\n${BLUE}2. Important Manual Steps Required:${NC}"
echo -e "${YELLOW}   - You must install the Diplus and Termux:Boot apps from Github (see README.md),${NC}"
echo -e "${YELLOW}   - and make sure you configured the apps according to the app instructions.${NC}"
echo -e "${YELLOW}   - You must enable 'Wireless debugging' in Android Developer Options.${NC}"
read -p "Press [Enter] to continue once you have completed these steps..."

# 3a. Connect ADB to self
echo -e "\n${BLUE}3a. Connecting ADB to localhost (make sure to Accept and remember the connection)...${NC}"
if adbs true; then
  echo "✅ ADB connected."
else
  echo -e "${RED}❌ Failed to connect to ADB.${NC}"
  exit 1
fi

# 3b. Enable background start for Termux, Termux:Boot and Termux:API
echo -e "\n${BLUE}3b. Opening 'Deactive background start' app, uncheck Diplus, Termux, Termux:Boot and Termux:API and hit OK...${NC}"
adb -s "$ADB_SERVER" shell "am start -n com.byd.appstartmanagement/.frame.AppStartManagement" >/dev/null 2>&1 || true
read -p "Press [Enter] to continue once you have completed these steps..." || true

# 4. Create Directories
echo -e "\n${BLUE}4. Creating necessary directories...${NC}"
# Local directories for logs
mkdir -p "$INSTALL_DIR" 2>/dev/null || true
# Starter script directory
mkdir -p "$BOOT_DIR" 2>/dev/null || true
# Shared storage on Android side (binary + config + logs)
adbs "mkdir -p $SHARED_DIR"
echo "✅ Directories created."

# 5. Download Latest Binary
echo -e "\n${BLUE}5. Downloading latest binary from GitHub...${NC}"
RELEASE_INFO=$(curl -s "$RELEASES_API")
DOWNLOAD_URL=$(echo "$RELEASE_INFO" | jq -r --arg ASSET_NAME "$ASSET_NAME" '.assets[] | select(.name == $ASSET_NAME) | .browser_download_url')
if [ -z "$DOWNLOAD_URL" ] || [ "$DOWNLOAD_URL" == "null" ]; then
    echo -e "${RED}❌ Could not find asset '$ASSET_NAME' in the latest release.${NC}"
    exit 1
fi
LATEST_VERSION=$(echo "$RELEASE_INFO" | jq -r .tag_name)
echo "Downloading '$ASSET_NAME' v$LATEST_VERSION..."
curl -sL -o "$TEMP_BINARY_PATH" "$DOWNLOAD_URL"
chmod +x "$TEMP_BINARY_PATH"
echo "✅ Download complete."

# 6. Stop Previous Instances (Comprehensive Cleanup)
echo -e "\n${BLUE}6. Stopping any previous instances...${NC}"
cleanup_all_processes

# Move new binary into shared storage location
mv "$TEMP_BINARY_PATH" "$BINARY_PATH"

# Copy binary into exec-friendly location and make it runnable for the shell user
echo -e "\n${BLUE}6b. Copying binary to exec-friendly path (/data/local/tmp)...${NC}"
adbs "cp $BINARY_PATH $EXEC_PATH && chmod 755 $EXEC_PATH"
echo "✅ Binary copied to $EXEC_PATH and made executable."

# 7. Configuration
CONFIG_CHANGED=false
if [ -f "$CONFIG_PATH" ]; then
  echo -e "\n${BLUE}7. Existing configuration detected at $CONFIG_PATH.${NC}"
  read -p "   - Do you want to update the configuration? (y/N): " UPDATE_CONF || true
  if [ "${UPDATE_CONF,,}" == "y" ]; then
    CONFIG_CHANGED=true
  else
    echo "✅ Keeping existing configuration."
  fi
else
  CONFIG_CHANGED=true
fi

if [ "$CONFIG_CHANGED" = true ]; then
  echo -e "\n${BLUE}7. Please provide your configuration:${NC}"
  read -p "   - MQTT WebSocket URL (e.g., ws://user:pass@host:port): " MQTT_URL || true
  read -p "   - ABRP API Key (optional): " ABRP_API_KEY || true
  if [ -n "$ABRP_API_KEY" ]; then
    read -p "   - ABRP User Token: " ABRP_TOKEN || true
  else
    ABRP_TOKEN=""
  fi
  read -p "   - Enable verbose logging? (y/N): " VERBOSE_INPUT || true
  VERBOSE=$([ "${VERBOSE_INPUT,,}" == "y" ] && echo "true" || echo "false")
  # Ask whether the ABRP Android app must be running (only relevant if an API key was provided)
  if [ -n "$ABRP_API_KEY" ]; then
    read -p "   - Require the ABRP Android app to be running? (Y/n): " REQUIRE_ABRP_INPUT || true
    REQUIRE_ABRP_APP=$([ "${REQUIRE_ABRP_INPUT,,}" == "n" ] && echo "false" || echo "true")
  else
    # No ABRP telemetry configured, disable requirement explicitly
    REQUIRE_ABRP_APP="false"
  fi

  # 8. Create or update Environment File
  echo -e "\n${BLUE}8. Saving environment configuration...${NC}"
  cat > "$CONFIG_PATH" << EOF
# Configuration for byd-hass service
export BYD_HASS_MQTT_URL='$MQTT_URL'
export BYD_HASS_ABRP_API_KEY='$ABRP_API_KEY'
export BYD_HASS_ABRP_TOKEN='$ABRP_TOKEN'
export BYD_HASS_VERBOSE='$VERBOSE'
export BYD_HASS_REQUIRE_ABRP_APP='$REQUIRE_ABRP_APP'
EOF
  echo "✅ Config file saved at $CONFIG_PATH"
else
  echo "Using existing configuration file."
fi

# 9. Create External Guardian Script
echo -e "\n${BLUE}9. Creating external keep-alive script...${NC}"
adbs "cat > $ADB_KEEPALIVE_SCRIPT_PATH" << KEEP_ALIVE_EOF
#!/system/bin/sh
echo "[\$(date)] BYD-HASS keep-alive started." >> "$ADB_LOG_FILE"

# Path to binary & config (mounted shared storage)
BIN_EXEC="$EXEC_PATH"            # Exec-friendly path inside Android shell
BIN_SRC="$BINARY_PATH"           # Persistent copy in shared storage
CONFIG_PATH="$CONFIG_PATH"
LOG_FILE="$LOG_FILE"
ADB_LOG_FILE="$ADB_LOG_FILE"

# Track current day for daily log rotation
CUR_DAY="\$(date +%Y%m%d)"

# Ensure executable exists (copy from shared storage if /data/local/tmp was cleared)
if [ ! -x "\$BIN_EXEC" ]; then
    cp "\$BIN_SRC" "\$BIN_EXEC" && chmod 755 "\$BIN_EXEC"
fi

while true; do
  # Export configuration variables if present
    if [ -f "\$CONFIG_PATH" ]; then
        . "\$CONFIG_PATH"
    fi

    # Rotate logs daily, keep only yesterday's copy
    NEW_DAY="\$(date +%Y%m%d)"
    if [ "\$NEW_DAY" != "\$CUR_DAY" ]; then
        # Remove previous .old
        rm -f "\$LOG_FILE.old" "\$ADB_LOG_FILE.old"
        # Rotate current to .old if exists
        [ -f "\$LOG_FILE" ] && mv "\$LOG_FILE" "\$LOG_FILE.old"
        [ -f "\$ADB_LOG_FILE" ] && mv "\$ADB_LOG_FILE" "\$ADB_LOG_FILE.old"
        CUR_DAY="\$NEW_DAY"
    fi

    # Ensure executable exists (copy from shared storage if /data/local/tmp was cleared)
    if [ ! -x "\$BIN_EXEC" ]; then
        cp "\$BIN_SRC" "\$BIN_EXEC" && chmod 755 "\$BIN_EXEC"
    fi

    if ! pgrep -f "\$BIN_EXEC" > /dev/null; then
        echo "[\$(date)] BYD-HASS not running. Starting it..." >> "$ADB_LOG_FILE"
        nohup "\$BIN_EXEC" >> \$LOG_FILE 2>&1 &
    fi
    sleep 10
done
KEEP_ALIVE_EOF
adbs "chmod +x $ADB_KEEPALIVE_SCRIPT_PATH"
echo "✅ External keep-alive script created."

# 10. Create Termux:Boot Orchestrator Script
echo -e "\n${BLUE}10. Creating Termux:Boot orchestrator script...${NC}"
cat > "$BOOT_SCRIPT_PATH" << BOOT_EOF
#!/data/data/com.termux/files/usr/bin/sh
termux-wake-lock

# Ensure ADB is connected before executing commands
if ! adb devices | grep -q "$ADB_SERVER"; then
    adb connect "$ADB_SERVER"  > /dev/null 2>&1
fi

# This script is the starter, launched by Termux:Boot. It only ensures
# that the external keep-alive guardian is running on the Android side.

# --- Simple Log Rotation for starter log ---
if [ -f "$INTERNAL_LOG_FILE" ]; then
    mv -f "$INTERNAL_LOG_FILE" "$INTERNAL_LOG_FILE.old"
fi

# Redirect all subsequent output of this orchestrator session to the fresh log
exec >> "$INTERNAL_LOG_FILE" 2>&1

echo "---"
echo "[\$(date)] Starter script running."

while true; do
    # Ensure ADB connection is alive; reconnect if necessary
    if ! adb devices | grep -q "$ADB_SERVER"; then
        adb connect "$ADB_SERVER" >/dev/null 2>&1
    fi

    # Count keep-alive processes by checking if the script file is being executed
    # Use a more reliable method that won't match the check itself
    PROCESS_COUNT=\$(adb -s "$ADB_SERVER" shell "ps -ef 2>/dev/null | grep '/storage/emulated/0/bydhass/keep-alive.sh' | grep -v grep | wc -l" 2>/dev/null || echo "0")
    
    if [ "\${PROCESS_COUNT:-0}" -eq 0 ]; then
        echo "[\$(date)] Keep-alive not running. Starting it..."
        adb -s "$ADB_SERVER" shell "nohup sh $ADB_KEEPALIVE_SCRIPT_PATH > /dev/null 2>&1 &"
    elif [ "\${PROCESS_COUNT:-0}" -gt 1 ]; then
        echo "[\$(date)] Multiple keep-alive processes detected (\$PROCESS_COUNT). Cleaning up..."
        adb -s "$ADB_SERVER" shell "pkill -f $ADB_KEEPALIVE_SCRIPT_NAME"
        sleep 2
        echo "[\$(date)] Starting single keep-alive process..."
        adb -s "$ADB_SERVER" shell "nohup sh $ADB_KEEPALIVE_SCRIPT_PATH > /dev/null 2>&1 &"
    fi
    sleep 60
done
BOOT_EOF
chmod +x "$BOOT_SCRIPT_PATH"
echo "✅ Termux:Boot orchestrator script created."

# 10a. Create Termux:Boot GPS Script
echo -e "\n${BLUE}10a. Creating Termux:Boot GPS script...${NC}"
cat > "$BOOT_GPS_SCRIPT_PATH" << 'BOOT_GPS_EOF'
#!/data/data/com.termux/files/usr/bin/bash

# --- CONFIG -----------------------------------------------------

INTERVAL=12  # seconds

# Previous values (empty on first run)
OLD_LAT=""
OLD_LON=""

# Trim function: keep 6 decimal digits
trim() {
        printf "%.6f" "$1"
}

trim_acc() {
        printf "%.2f" "$1"
}

to_decimal() {
        printf "%f" "$1"
}

# --- LOOP -------------------------------------------------------
while true; do

        # Request GPS fix
        LOC=$(termux-location)

        # If termux-location failed, wait and retry
        if [ -z "$LOC" ]; then
                sleep $INTERVAL
                continue
        fi

        # Extract raw values
        RAW_LAT=$(echo "$LOC" | jq -r .latitude)
        RAW_LON=$(echo "$LOC" | jq -r .longitude)
        SPD=$(echo "$LOC" | jq -r .speed)
        RAW_ACC=$(echo "$LOC" | jq -r .accuracy)

        # Trim coordinates
        LAT=$(trim "$RAW_LAT")
        LON=$(trim "$RAW_LON")
        ACC=$(trim_acc "$RAW_ACC")

        # -------------------------
        # CHANGE DETECTION SECTION
        # -------------------------

        if [ -n "$OLD_LAT" ] && [ -n "$OLD_LON" ]; then

                # Compute difference
                DIFF_LAT=$(awk -v a="$LAT" -v b="$OLD_LAT" 'BEGIN{print (a-b)}')
                DIFF_LON=$(awk -v a="$LON" -v b="$OLD_LON" 'BEGIN{print (a-b)}')

                # Absolute values
                ABS_LAT=$(awk -v x="$DIFF_LAT" 'BEGIN {print (x<0?-x:x)}')
                ABS_LON=$(awk -v x="$DIFF_LON" 'BEGIN {print (x<0?-x:x)}')

                # Only publish if moved >0.00001° (~1 m)
                ABS_LAT_DEC=$(to_decimal "$ABS_LAT")
                ABS_LON_DEC=$(to_decimal "$ABS_LON")

                if (( $(echo "$ABS_LAT_DEC < 0.00001" | bc -l) )) && \
                   (( $(echo "$ABS_LON_DEC < 0.00001" | bc -l) )); then
                        # No significant change → skip publish
                        sleep $INTERVAL
                        continue
                fi
        fi

        # Update previous values
        OLD_LAT="$LAT"
        OLD_LON="$LON"

        # -------------------------
        # JSON PAYLOAD
        # -------------------------

        JSON_PAYLOAD=$(jq -n -c \
                --arg lat "$LAT" \
                --arg lon "$LON" \
                --arg spd "$SPD" \
                --arg acc "$ACC" \
                '{
                        latitude: ($lat|tonumber),
                        longitude: ($lon|tonumber),
                        speed: ($spd|tonumber),
                        accuracy: ($acc|tonumber)
                }')

        echo "$JSON_PAYLOAD" > /storage/emulated/0/bydhass/gps
        sleep $INTERVAL
done
BOOT_GPS_EOF
chmod +x "$BOOT_GPS_SCRIPT_PATH"
echo "✅ Termux:Boot GPS script created."

# 10b. Ensure .bashrc autostart entries
BASHRC_PATH="$HOME/.bashrc"
AUTOSTART_CMD="$BOOT_SCRIPT_PATH &"
AUTOSTART_GPS_CMD="$BOOT_GPS_SCRIPT_PATH &"

echo -e "\n${BLUE}10b. Ensuring .bashrc autostart entries...${NC}"
# Create .bashrc if it does not exist
if [ ! -f "$BASHRC_PATH" ]; then
  touch "$BASHRC_PATH"
  echo "Created $BASHRC_PATH"
fi
# Add orchestrator autostart if not already present
if grep -Fxq "$AUTOSTART_CMD" "$BASHRC_PATH"; then
  echo "✅ Orchestrator autostart entry already present."
else
  echo "$AUTOSTART_CMD" >> "$BASHRC_PATH"
  echo "✅ Added orchestrator autostart entry."
fi
# Add GPS autostart if not already present
if grep -Fxq "$AUTOSTART_GPS_CMD" "$BASHRC_PATH"; then
  echo "✅ GPS autostart entry already present."
else
  echo "$AUTOSTART_GPS_CMD" >> "$BASHRC_PATH"
  echo "✅ Added GPS autostart entry."
fi

# 11. Start the services
echo -e "\n${BLUE}11. Starting the services...${NC}"
nohup sh "$BOOT_SCRIPT_PATH" > /dev/null 2>&1 &
ORCHESTRATOR_PID=$!
nohup sh "$BOOT_GPS_SCRIPT_PATH" > /dev/null 2>&1 &
GPS_PID=$!

# Wait a moment and verify the services started
sleep 2
if kill -0 "$ORCHESTRATOR_PID" 2>/dev/null; then
    echo "✅ Orchestrator started successfully (PID: $ORCHESTRATOR_PID)"
    echo "   → The orchestrator will start the keep-alive script via ADB"
    echo "   → The keep-alive script will start the BYD-HASS binary"
else
    echo "⚠️  Orchestrator may have failed to start. Check logs: tail -f $INTERNAL_LOG_FILE"
fi
if kill -0 "$GPS_PID" 2>/dev/null; then
    echo "✅ GPS script started successfully (PID: $GPS_PID)"
else
    echo "⚠️  GPS script may have failed to start."
fi

echo -e "\n${GREEN}🎉 Installation complete! BYD-HASS is now managed by a self-healing service.${NC}"
echo -e "${YELLOW}The service will restart automatically if the app is killed or the device reboots.${NC}"
echo -e "${YELLOW}To see the main app logs, run: tail -f $LOG_FILE${NC}"
echo -e "${YELLOW}To see the orchestrator logs, run: tail -f $INTERNAL_LOG_FILE${NC}"
echo -e "${YELLOW}To see the external guardian logs, run: tail -f $ADB_LOG_FILE${NC}"
echo -e "${YELLOW}To stop everything, run: ./install.sh cleanup${NC}"
echo -e "${YELLOW}To reinstall/update, re-run this install script.${NC}"

adb disconnect "$ADB_SERVER" >/dev/null 2>&1 || true
exit 0
