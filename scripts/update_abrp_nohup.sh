#!/data/data/com.termux/files/usr/bin/bash

# Exit if we're running already
pgrep -f "$(basename "$0")" | grep -v "^$$\$" | grep -q . && exit

# ================================
# Configuration
# ================================

SCRIPT_DIR="$(dirname "$0")"
ABRP_CONFIG_FILE="$SCRIPT_DIR/abrp_config"
CACHE_DIR="$SCRIPT_DIR/ha_cache"
POLL_INTERVAL=60  # seconds
ABRP_PACKAGE_NAME="com.iternio.abrpapp"

# Sensors to send to ABRP
ABRP_SENSORS=(
  "byd_car_battery_soc"
  "byd_car_charge_gun_state"
  "byd_car_speed"
  "byd_car_gps_latitude"
  "byd_car_gps_longitude"
)

# ================================
# Load config
# ================================

if [[ ! -f "$ABRP_CONFIG_FILE" ]]; then
  echo "❌ Config file 'abrp_config' not found in script directory. Exiting."
  exit 1
fi

source "$ABRP_CONFIG_FILE"

if [[ -z "$ABRP_API_KEY" || -z "$ABRP_USER_TOKEN" ]]; then
  echo "❌ ABRP_API_KEY or ABRP_USER_TOKEN not set in 'abrp_config'. Exiting."
  exit 1
fi

# ================================
# Functions
# ================================

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

get_sensor_value() {
  local sensor="$1"
  local path="$CACHE_DIR/$sensor"
  [[ -f "$path" ]] && cat "$path"
}

send_to_abrp() {
  local soc="$1"
  local lat="$2"
  local lon="$3"
  local speed="$4"
  local charging="$5"
  local timestamp
  timestamp=$(date +%s)

  local payload
  payload=$(jq -n \
    --arg utc "$timestamp" \
    --arg soc "$soc" \
    --arg lat "$lat" \
    --arg lon "$lon" \
    --arg speed "$speed" \
    --arg charging "$charging" \
    '{
      utc: ($utc|tonumber),
      soc: ($soc|tonumber),
      lat: ($lat|tonumber),
      lon: ($lon|tonumber),
      speed: ($speed|tonumber),
      is_charging: ($charging|tonumber)
    }')

  curl -s -X POST \
    "https://api.iternio.com/1/tlm/send?api_key=$ABRP_API_KEY&token=$ABRP_USER_TOKEN&tlm=$(jq -sRr @uri <<< "$payload")" \
    -o /dev/null

  log "✅ Sent to ABRP: soc=$soc, lat=$lat, lon=$lon, speed=$speed, charging=$charging"
}

is_abrp_running() {
  # Check if the ABRP app process is running
  pidof "$ABRP_PACKAGE_NAME" > /dev/null 2>&1 || pgrep -f "$ABRP_PACKAGE_NAME" > /dev/null 2>&1
}

# ================================
# Main Loop
# ================================

declare -A last_values

log "🚀 Starting ABRP reporter..."

while true; do
  # Read current values
  soc=$(get_sensor_value "byd_car_battery_soc")
  speed=$(get_sensor_value "byd_car_speed")
  lat=$(get_sensor_value "byd_car_gps_latitude")
  lon=$(get_sensor_value "byd_car_gps_longitude")
  charging=$(get_sensor_value "byd_car_charge_gun_state")

  # Skip if soc is missing or invalid
  if ! [[ "$soc" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
    log "⚠️ Invalid or missing SoC: '$soc', skipping..."
    sleep "$POLL_INTERVAL"
    continue
  fi

  if ! is_abrp_running; then
    log "🛑 ABRP app not running. Skipping transmission."
    sleep "$POLL_INTERVAL"
    continue
  fi

  # Only send if values changed since last time
  changed=false
  for key in soc speed lat lon charging; do
    current="${!key}"
    if [[ "${last_values[$key]}" != "$current" ]]; then
      changed=true
      last_values[$key]="$current"
    fi
  done

  if [[ "$changed" == true ]]; then
    send_to_abrp "$soc" "$lat" "$lon" "$speed" "$charging"
  else
    log "⏩ No changes, skipping ABRP update."
  fi

  sleep "$POLL_INTERVAL"
done
