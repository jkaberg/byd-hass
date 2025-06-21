#!/data/data/com.termux/files/usr/bin/bash

# Push all cached sensor values to Home Assistant

SCRIPT_DIR="$(dirname "$0")"
HASS_CONFIG_FILE="$SCRIPT_DIR/hass_config"
CACHE_DIR="/data/data/com.termux/files/home/scripts/ha_cache"
HA_SENSOR_PREFIX="byd_car_"

if [[ ! -f "$HASS_CONFIG_FILE" ]]; then
  echo "❌ Config file 'hass_config' not found in script directory. Exiting."
  exit 1
fi

source "$HASS_CONFIG_FILE"

if [[ -z "$HA_BASE_URL" || -z "$HA_TOKEN" ]]; then
  echo "❌ HA_BASE_URL or HA_TOKEN not set in 'hass_config'. Exiting."
  exit 1
fi

for file in "$CACHE_DIR"/*; do
  [[ -f "$file" ]] || continue
  filename="$(basename "$file")"
  value="$(cat "$file")"

  curl -s -o /dev/null -X POST "$HA_BASE_URL/api/states/sensor.${HA_SENSOR_PREFIX}${filename}" \
    -H "Authorization: Bearer $HA_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"state\": \"$value\", \"attributes\": {\"unit_of_measurement\": \"none\", \"friendly_name\": \"$filename\"}}"
done
