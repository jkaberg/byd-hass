# BYD-HASS

A Go application that exports BYD vehicle data to Home Assistant and ABRP.

## What It Does

BYD-HASS polls your BYD vehicle's local Diplus API and forwards the data to:
- **Home Assistant** via MQTT (with auto-discovery)
- **A Better Route Planner (ABRP)** for optimized route planning

The application runs as a static binary on Android devices through Termux.

## Requirements

- [Diplus app](http://lanye.pw/di/)
- [Termux](https://termux.com/)
- [Termux:Boot](https://github.com/termux/termux-boot)
- [Termux:API](https://github.com/termux/termux-api)
- MQTT broker for Home Assistant integration

**Make sure you've configured the above properly before running `byd-hass`!**

## Installation

Run this command in Termux:

```bash
curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/main/install.sh | bash
```

The installer will:
1. Install required dependencies
2. Download the latest byd-hass binary
3. Guide you through configuration
4. Set up automatic startup

## Configuration

### Command Line Options
```bash
byd-hass [options]

Options:
  -mqtt-url string          MQTT URL (ws://, wss://, mqtt://, mqtts://)
  -abrp-api-key string      ABRP API key (optional)
  -abrp-vehicle-key string  ABRP vehicle key (optional)
  -device-id string         Device identifier (auto-generated)
  -verbose                  Enable detailed logging
  -discovery-prefix string  Home Assistant prefix (default: homeassistant)
```

### Environment Variables
All options can be set with `BYD_HASS_` prefix:
- `BYD_HASS_MQTT_URL`
- `BYD_HASS_ABRP_API_KEY`
- `BYD_HASS_ABRP_VEHICLE_KEY`
- `BYD_HASS_DEVICE_ID`
- `BYD_HASS_VERBOSE`

### Supported MQTT Protocols
- `ws://` - MQTT over WebSocket
- `wss://` - MQTT over WebSocket Secure
- `mqtt://` - Standard MQTT
- `mqtts://` - MQTT over SSL/TLS

## How It Works

- **Diplus API**: Polled every 15 seconds for latest vehicle data
- **MQTT Transmission**: Sends to Home Assistant every 60 seconds (only if data changed)
- **ABRP Transmission**: Sends to ABRP every 10 seconds (only if data changed)
- **Caching**: Avoids redundant transmissions by detecting value changes

## Home Assistant Integration

The application automatically creates sensors for:
- Battery state of charge and charging status
- Speed, mileage, and location
- Door and window status
- Interior and exterior temperatures
- Tire pressures
- Various vehicle systems status

All sensors appear as a single device with proper device classes and units.

## ABRP Integration

Provides telemetry data including:
- Battery state of charge and power consumption
- GPS location and speed
- Charging status and power estimation
- Temperature data for range calculations

Data helps ABRP provide accurate route planning and energy consumption estimates.

## Service Management

### View Logs
```bash
adb shell tail -f /storage/emulated/0/bydhass/byd-hass.log
```

### Restart Service
Re-run the installer to restart with new configuration.

### Stop Service
The installer will prompt to stop any running instance.

## Building from Source

```bash
./build.sh  # Creates byd-hass-arm64 for Android
```

Cross-compiles for Android ARM64 architecture.

## Notes

- This is unofficial software - use at your own risk
