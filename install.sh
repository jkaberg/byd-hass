#!/data/data/com.termux/files/usr/bin/bash

# BYD-HASS Bootstrapping Installation Script
# This script sets up a self-healing, two-tiered system.
# 1. A Termux:Boot script acts as the orchestrator.
# 2. The orchestrator starts an external keep-alive script via ADB.
# 3. The external script ensures the Termux app itself stays running.
# 4. The orchestrator then starts the main byd-hass binary.

set -e

# --- Configuration ---
# Termux-internal paths
INSTALL_DIR="$HOME/.byd-hass"
BINARY_NAME="byd-hass"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
CONFIG_PATH="$INSTALL_DIR/config.env"
LOG_FILE="$INSTALL_DIR/byd-hass.log"
INTERNAL_LOG_FILE="$INSTALL_DIR/starter.log"

# Termux:Boot script (The Orchestrator)
BOOT_DIR="$HOME/.termux/boot"
BOOT_SCRIPT_NAME="byd-hass-starter.sh"
BOOT_SCRIPT_PATH="$BOOT_DIR/$BOOT_SCRIPT_NAME"

# ADB-accessible shared storage paths (The External Guardian)
SHARED_DIR="/storage/emulated/0/bydhass"
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

# --- Script Start ---
echo -e "${GREEN}🚗 BYD-HASS Bootstrapping Installer${NC}"

# 1. Setup Termux Environment
echo -e "\n${BLUE}1. Setting up Termux environment...${NC}"
echo "Installing dependencies (adb, curl, jq, termux-api)..."
pkg install -y android-tools curl jq termux-api >/dev/null 2>&1
echo "✅ Environment ready."

# 2. Explain Manual Steps
echo -e "\n${BLUE}2. Important Manual Steps Required:${NC}"
echo -e "${YELLOW}   - You must install the Diplus, Termux:Boot and Termux:API apps from Github (see README.md),${NC}"
echo -e "${YELLOW}   - and make sure you configured the apps according to the app instructions.${NC}"
echo -e "${YELLOW}   - You must enable 'Wireless debugging' in Android Developer Options.${NC}"
read -p "Press [Enter] to continue once you have completed these steps..."

# 3a. Connect ADB to self
echo -e "\n${BLUE}3a. Connecting ADB to localhost (make sure to Accept and remember the connection)...${NC}"
adb connect "$ADB_SERVER"
echo "✅ ADB connected."

# 3b. Enable background start for Termux, Termux:Boot and Termux:API
echo -e "\n${BLUE}3b. Opening 'Deactive background start' app, uncheck Termux, Termux:Boot and Termux:API and hit OK...${NC}"
adb -s "$ADB_SERVER" shell "am start -n com.byd.appstartmanagement/.frame.AppStartManagement" >/dev/null 2>&1
read -p "Press [Enter] to continue once you have completed these steps..."

# 4. Create Directories
echo -e "\n${BLUE}4. Creating necessary directories...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$BOOT_DIR"
adb -s "$ADB_SERVER" shell "mkdir -p '$SHARED_DIR'"
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

# 6. Stop Previous Instances
echo -e "\n${BLUE}6. Stopping any previous instances...${NC}"
adb -s "$ADB_SERVER" shell "pkill -f '$ADB_KEEPALIVE_SCRIPT_NAME'" || true
pkill -f "$BOOT_SCRIPT_NAME" || true
pkill -f "$BINARY_PATH" || true
echo "✅ Old processes terminated."

# Move new binary into place
mv "$TEMP_BINARY_PATH" "$BINARY_PATH"

# 7. Get User Configuration
echo -e "\n${BLUE}7. Please provide your configuration:${NC}"
read -p "   - MQTT WebSocket URL (e.g., ws://user:pass@host:port): " MQTT_URL
read -p "   - ABRP API Key (optional): " ABRP_API_KEY
if [ -n "$ABRP_API_KEY" ]; then
  read -p "   - ABRP Vehicle Key: " ABRP_VEHICLE_KEY
else
  ABRP_VEHICLE_KEY=""
fi
read -p "   - Enable verbose logging? (y/N): " VERBOSE_INPUT
VERBOSE=$([ "${VERBOSE_INPUT,,}" == "y" ] && echo "true" || echo "false")

# 8. Create Environment File
echo -e "\n${BLUE}8. Creating environment configuration file...${NC}"
cat > "$CONFIG_PATH" << EOF
# Configuration for byd-hass service
export BYD_HASS_MQTT_URL='$MQTT_URL'
export BYD_HASS_ABRP_API_KEY='$ABRP_API_KEY'
export BYD_HASS_ABRP_VEHICLE_KEY='$ABRP_VEHICLE_KEY'
export BYD_HASS_VERBOSE='$VERBOSE'
EOF
echo "✅ Config file created at $CONFIG_PATH"

# 9. Create External Guardian Script
echo -e "\n${BLUE}9. Creating external keep-alive script...${NC}"
adb -s "$ADB_SERVER" shell "cat > '$ADB_KEEPALIVE_SCRIPT_PATH'" << KEEP_ALIVE_EOF
#!/system/bin/sh
echo "[\$(date)] External keep-alive service started." >> "$ADB_LOG_FILE"
while true; do
    if ! pgrep -x "com.termux" > /dev/null; then
        echo "[\$(date)] Termux not running. Starting it..." >> "$ADB_LOG_FILE"
        am start -n com.termux/.HomeActivity
        sleep 2
        input keyevent KEYCODE_HOME
    fi
    sleep 10
done
KEEP_ALIVE_EOF
adb shell "chmod +x '$ADB_KEEPALIVE_SCRIPT_PATH'"
echo "✅ External keep-alive script created."

# 10. Create Termux:Boot Orchestrator Script
echo -e "\n${BLUE}10. Creating Termux:Boot orchestrator script...${NC}"
cat > "$BOOT_SCRIPT_PATH" << BOOT_EOF
#!/data/data/com.termux/files/usr/bin/sh

# This script is the main orchestrator, started by Termux:Boot.
# It ensures the external guardian is running, then starts the main app.

exec >> "$INTERNAL_LOG_FILE" 2>&1

echo "---"
echo "[\$(date)] Orchestrator started."

# 1. Ensure the external ADB-based guardian is running.
if ! adb devices | grep -q "$ADB_SERVER"; then
    adb connect "$ADB_SERVER"
fi

if ! pgrep -f "$ADB_KEEPALIVE_SCRIPT_PATH" > /dev/null; then
    echo "[\$(date)] External guardian not found. Starting it..."
    adb shell "nohup sh '$ADB_KEEPALIVE_SCRIPT_PATH' > /dev/null 2>&1 &"
else
    echo "[\$(date)] External guardian is already running."
fi

# 2. Run the main byd-hass application in its own keep-alive loop.
. "$CONFIG_PATH"
while true; do
    echo "[\$(date)] Acquiring wake lock and starting byd-hass service..."
    termux-wake-lock
    $BINARY_PATH
    echo "[\$(date)] Service stopped with exit code \$?. Restarting in 30 seconds..."
    sleep 30
done >> "$LOG_FILE"
BOOT_EOF
chmod +x "$BOOT_SCRIPT_PATH"
echo "✅ Termux:Boot orchestrator script created."

# 11. Start the main orchestrator script
echo -e "\n${BLUE}11. Starting the main service orchestrator...${NC}"
nohup sh "$BOOT_SCRIPT_PATH" > /dev/null 2>&1 &
echo "Orchestrator has been launched. It will start the other components."

echo -e "\n${GREEN}🎉 Installation complete! BYD-HASS is now managed by a self-healing service.${NC}"
echo -e "${YELLOW}The service will restart automatically if the app is killed or the device reboots.${NC}"
echo -e "${YELLOW}To see the main app logs, run: tail -f $LOG_FILE${NC}"
echo -e "${YELLOW}To see the orchestrator logs, run: tail -f $INTERNAL_LOG_FILE${NC}"
echo -e "${YELLOW}To see the external guardian logs, run: tail -f $ADB_LOG_FILE${NC}"
echo -e "${YELLOW}To stop everything, re-run this install script.${NC}"

adb disconnect "$ADB_SERVER" >/dev/null 2>&1 || true
exit 0
