#!/data/data/com.termux/files/usr/bin/bash

# Exit if we're running already
pgrep -f "$(basename "$0")" | grep -v "^$$\$" | grep -q . && exit

# ================================
# Configuration
# ================================

SCRIPT_DIR="$(dirname "$0")"
HASS_CONFIG_FILE="$SCRIPT_DIR/hass_config"
CACHE_DIR="$SCRIPT_DIR/ha_cache"
HA_SENSOR_PREFIX="byd_car_"
POLL_INTERVAL=60  # seconds

# ================================
# Startup checks
# ================================

if [[ ! -f "$HASS_CONFIG_FILE" ]]; then
  echo "❌ Config file 'hass_config' not found in script directory. Exiting."
  exit 1
fi

source "$HASS_CONFIG_FILE"

if [[ -z "$HA_BASE_URL" || -z "$HA_TOKEN" ]]; then
  echo "❌ HA_BASE_URL or HA_TOKEN not set in 'hass_config'. Exiting."
  exit 1
fi

# ================================
# Functions
# ================================

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

push_sensor_to_hass() {
  local sensor="$1"
  local value="$2"
  local entity="sensor.${HA_SENSOR_PREFIX}${sensor}"

  curl -s -o /dev/null -X POST "$HA_BASE_URL/api/states/$entity" \
    -H "Authorization: Bearer $HA_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"state\": \"$value\", \"attributes\": {\"unit_of_measurement\": \"none\", \"friendly_name\": \"$sensor\"}}"

  log "📡 Updated $entity to '$value'"
}

# ================================
# Main Loop
# ================================

declare -A last_values

log "🚀 Starting HASS sensor updater..."

while true; do
  for file in "$CACHE_DIR"/*; do
    [[ -f "$file" ]] || continue
    sensor="$(basename "$file")"
    value="$(cat "$file")"

    if [[ "${last_values[$sensor]}" != "$value" ]]; then
      push_sensor_to_hass "$sensor" "$value"
      last_values[$sensor]="$value"
    else
      log "⏩ $sensor unchanged, skipping"
    fi
  done

  sleep "$POLL_INTERVAL"
done
