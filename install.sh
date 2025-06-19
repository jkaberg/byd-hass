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
curl -o "$BOOT_DIR/run.sh" "$BASE_URL/run.sh"
chmod +x "$BOOT_DIR/run.sh"

# Download our scripts
mkdir -p "$SCRIPTS_DIR"
curl -sL -o "$SCRIPTS_DIR/poll_diplus_nohup.sh" "$BASE_URL/scripts/poll_diplus_nohup.sh"
curl -sL -o "$SCRIPTS_DIR/keep_alive_nohup.sh" "$BASE_URL/scripts/keep_alive_nohup.sh"

chmod +x $SCRIPTS_DIR/*

bash $BOOT_DIR/run.sh