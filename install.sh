#!/data/data/com.termux/files/usr/bin/bash

HOME="/data/data/com.termux/files/home"
BOOT_DIR="$HOME/.termux/boot"
SCRIPTS_DIR="$HOME/scripts"

BASE_URL="https://raw.githubusercontent.com/jkaberg/byd-hass/refs/heads/main/"

# Install required packages
pkg install -y jq

### Setup the environment ###

# Boot dir is an part of Termux:Boot init 
mkdir -p "$BOOT_DIR"
curl -sL -o "$BOOT_DIR/run.sh" "$BASE_URL/run.sh"
chmod +x "$BOOT_DIR/run.sh"

# Download our scripts
mkdir -p "$SCRIPTS_DIR"

scripts=("poll_diplus_nohup.sh" "keep_alive_nohup.sh")

for script in "${scripts[@]}"; do
  curl -sL -o "$SCRIPTS_DIR/$script" "$BASE_URL/scripts/$script"
  chmod +x "$SCRIPTS_DIR/$script"
done

bash $BOOT_DIR/run.sh