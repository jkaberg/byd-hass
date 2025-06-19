#!/data/data/com.termux/files/usr/bin/bash


# Exit if we're running already
pgrep -f "$(basename "$0")" | grep -v "^$$\$" | grep -q . && exit

# ================================
# Configuration
# ================================

# To extend what sensors are consumed and pushed to HASS see the url bellow whats available.
# API SPEC: https://apifox.com/apidoc/shared/c3ce5ff5-754f-438c-aef2-055d85aa0391/277818345e0

API_BASE_URL="http://localhost:8988/api/getDiPars"

# This is an key, value mapping seperated by |
# The chinese text is an reference, see API SPEC.
TEXT_TEMPLATE="soc:{电量百分比}|mileage:{里程}|lock:{远程锁车状态}|"

###  Home Assistant config ###
HASS_CONFIG_FILE="$(dirname "$0")/hass_config"

if [[ ! -f "$HASS_CONFIG_FILE" ]]; then
  echo "❌ Config file 'hass_config' not found in script directory. Exiting."
  exit 1
fi

source "$HASS_CONFIG_FILE"

if [[ -z "$HA_BASE_URL" || -z "$HA_TOKEN" ]]; then
  echo "❌ HA_BASE_URL or HA_TOKEN not set in 'hass_config'. Exiting."
  exit 1
fi
###  Home Assistant config ###

# Sensor prefix used in HASS
HA_SENSOR_PREFIX="byd_car_"

# Sensor mapping, where first delimiter is corresponding to the TEXT_TEMPLATE key and HA_SENSOR_PREFIX
# Format: json_key:ha_sensor:unit
# Example "output" in HASS: sensor.byd_car_battery_soc
HA_SENSORS=(
  "soc:battery_soc:%"
  "mileage:car_mileage:km"
  "lock:lock:none"
)

CACHE_DIR="/data/data/com.termux/files/home/scripts/ha_cache"

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

post_to_home_assistant() {
  local sensor_name="$1"
  local value="$2"
  local unit="$3"
  local full_sensor_name="${HA_SENSOR_PREFIX}${sensor_name}"

  curl -s -o /dev/null -X POST "$HA_BASE_URL/api/states/sensor.${full_sensor_name}" \
    -H "Authorization: Bearer $HA_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"state\": \"$value\", \"attributes\": {\"unit_of_measurement\": \"$unit\", \"friendly_name\": \"$sensor_name\"}}"
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
      unit="${rest#*:}"

      if [[ "$key" == "$map_key" ]]; then
        full_name="${HA_SENSOR_PREFIX}${ha_sensor}"
        old_value=$(get_cached_value "$full_name")

        if [[ "$value" != "$old_value" ]]; then
          log "🔁 Updating sensor.${full_name}: $old_value → $value $unit"
          post_to_home_assistant "$ha_sensor" "$value" "$unit"
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

