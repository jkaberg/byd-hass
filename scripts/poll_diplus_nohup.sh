#!/data/data/com.termux/files/usr/bin/bash

# Exit if we're running already
pgrep -f "$(basename "$0")" | grep -v "^$$\$" | grep -q . && exit

# ================================
# Configuration
# ================================

LAST_UPDATE_SENSOR="last_update"

API_BASE_URL="http://localhost:8988/api/getDiPars"
TEXT_TEMPLATE="soc:{电量百分比}|mileage:{里程}|lock:{远程锁车状态}|charge_gun_state:{充电枪插枪状态}|speed:{车速}"

# Charge gun state: 1 = disconnected | 2 = connected

SCRIPT_DIR="$(dirname "$0")"
HASS_CONFIG_FILE="$SCRIPT_DIR/hass_config"

if [[ ! -f "$HASS_CONFIG_FILE" ]]; then
  echo "❌ Config file 'hass_config' not found in script directory. Exiting."
  exit 1
fi

source "$HASS_CONFIG_FILE"

if [[ -z "$HA_BASE_URL" || -z "$HA_TOKEN" ]]; then
  echo "❌ HA_BASE_URL or HA_TOKEN not set in 'hass_config'. Exiting."
  exit 1
fi

HA_SENSOR_PREFIX="byd_car_"
CACHE_DIR="/data/data/com.termux/files/home/scripts/ha_cache"

HA_SENSORS=(
  "soc:battery_soc:%"
  "mileage:car_mileage:km"
  "lock:lock:none"
  "charge_gun_state:charge_gun_state:none"
  "speed:speed:km/h"
)

# ================================
# Functions
# ================================

log() {
  [[ "$VERBOSE" == true ]] && echo "$@"
}

urlencode() {
  local raw="$1"
  jq -nr --arg v "$raw" '$v|@uri'
}

fetch_data() {
  local url="${API_BASE_URL}?text=$(urlencode "$TEXT_TEMPLATE")"
  curl -s --fail "$url"
}

ensure_cache_dir() {
  mkdir -p "$CACHE_DIR"
}

get_cached_value() {
  local sensor_name="$1"
  local file="${CACHE_DIR}/${sensor_name}"
  [[ -f "$file" ]] && cat "$file"
}

set_cached_value() {
  local sensor_name="$1"
  local value="$2"
  echo "$value" > "${CACHE_DIR}/${sensor_name}"

  # Update last_update sensor if we're updating a real sensor (not last_update itself)
  if [[ "$sensor_name" != "${HA_SENSOR_PREFIX}${LAST_UPDATE_SENSOR}" ]]; then
    local ts=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "$ts" > "${CACHE_DIR}/${HA_SENSOR_PREFIX}${LAST_UPDATE_SENSOR}"
    log "🕒 Updated sensor.${HA_SENSOR_PREFIX}${LAST_UPDATE_SENSOR}: $ts"
  fi
}

process_response() {
  local json="$1"
  local success val_string

  success=$(echo "$json" | jq -r '.success')
  if [[ "$success" != "true" ]]; then
    log "❌ Request failed: $json"
    return 1
  fi

  val_string=$(echo "$json" | jq -r '.val')
  log "✅ Result: $val_string"

  IFS='|' read -ra pairs <<< "$val_string"
  for pair in "${pairs[@]}"; do
    key="${pair%%:*}"
    value="${pair#*:}"

    for map in "${HA_SENSORS[@]}"; do
      map_key="${map%%:*}"
      rest="${map#*:}"
      ha_sensor="${rest%%:*}"

      if [[ "$key" == "$map_key" ]]; then
        full_name="${HA_SENSOR_PREFIX}${ha_sensor}"
        old_value=$(get_cached_value "$full_name")

        if [[ "$value" != "$old_value" ]]; then
          log "🔁 Updating cache sensor.${full_name}: $old_value → $value"
          set_cached_value "$full_name" "$value"
        else
          log "⏩ Skipping unchanged sensor.${full_name}: $value"
        fi
      fi
    done
  done
}

# ================================
# Main Loop
# ================================

VERBOSE=false
[[ "$1" == "--verbose" ]] && VERBOSE=true

ensure_cache_dir
trap "echo '⏹️ Exiting...'; exit 0" SIGINT

while true; do
  response=$(fetch_data)
  if [[ $? -eq 0 ]]; then
    process_response "$response"
  else
    log "❌ Failed to fetch data"
  fi

  log "⏳ Waiting 60 seconds..."
  sleep 60
done
