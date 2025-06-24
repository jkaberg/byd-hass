#!/data/data/com.termux/files/usr/bin/bash

HOME="/data/data/com.termux/files/home"
BOOT_DIR="$HOME/.termux/boot"
SCRIPTS_DIR="$HOME/scripts"

BASE_URL="https://raw.githubusercontent.com/jkaberg/byd-hass/refs/heads/main/"

# Install required packages
pkg install -y jq

### Setup the environment ###

# Boot dir is part of Termux:Boot init
mkdir -p "$BOOT_DIR"
curl -sL -o "$BOOT_DIR/run.sh" "$BASE_URL/run.sh"
chmod +x "$BOOT_DIR/run.sh"

# Download our scripts
mkdir -p "$SCRIPTS_DIR"

scripts=("poll_diplus_nohup.sh" "update_abrp_nohup.sh" "update_hass_nohup.sh")

for script in "${scripts[@]}"; do
  # Kill any running instances of the script
  pkill -f "$SCRIPTS_DIR/$script"

  # Download the latest version
  curl -sL -o "$SCRIPTS_DIR/$script" "$BASE_URL/scripts/$script"
  chmod +x "$SCRIPTS_DIR/$script"
done

bash "$BOOT_DIR/run.sh"
