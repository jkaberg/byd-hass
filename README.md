# BYD-HASS

Export your BYD car data to Home Assistant and ABRP using a modern Go application.

## Overview

BYD-HASS is a standalone Go application that runs on an Android device (like the car's head unit) inside Termux. It polls the local Diplus API for vehicle data, and transmits it to Home Assistant (via MQTT) and A Better Route Planner (ABRP).

**Key Features:**
-   Efficient polling: 15s for Diplus, 10s for ABRP, 60s for MQTT.
-   Home Assistant MQTT auto-discovery for seamless integration.
-   Transmits data only on change to reduce network traffic.
-   Directly supports ABRP telemetry.
-   Interactive installer for easy setup.

## Features

*   Polls the BYD Diplus API for vehicle data at 15-second intervals.
*   Transmits data to Home Assistant via MQTT (every 60s) and/or A Better Route Planner (ABRP) (every 10s).
*   Intelligent caching to only transmit data when values have changed.
*   Automatic Home Assistant MQTT discovery for all supported sensors and a device tracker.
*   Graceful handling of API and network errors.
*   Configurable via command-line flags and environment variables.
*   Provides GPS location data for ABRP and Home Assistant using the Termux:API.

## Dependencies

This application is designed to run on an Android device using [Termux](https://termux.com/). For full functionality, including GPS location tracking, you must have the **Termux:API** add-on application installed and the `termux-api` package installed within Termux.

```bash
# Install the termux-api package from within Termux
pkg install termux-api
```

Without the Termux:API, the application will still function but will not be able to provide location data to ABRP or Home Assistant.

## Installation

### Prerequisites
-   A BYD vehicle with the [Diplus app](http://lanye.pw/di/) installed and active.
-   An Android device (typically the car's head unit) with [Termux](https://github.com/termux/termux-app) installed.
-   An MQTT broker with WebSocket support enabled.

### Install
Run the following command in your Termux session. The installer is interactive and will guide you through the configuration.

```bash
curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/main/install.sh | bash
```

The installer will:
1.  Install necessary dependencies (`adb`, `curl`, `jq`).
2.  Download the latest version of `byd-hass`.
3.  Prompt you for configuration details (MQTT, ABRP).
4.  Set up a keep-alive service to ensure the application is always running.

## Configuration

The application is configured during the installation. If you need to reconfigure, you can re-run the installer.

It can also be configured using command-line arguments or environment variables if you choose to run it manually.

### Command Line Options
```bash
byd-hass [options]

Options:
  -mqtt-url string          MQTT WebSocket URL (ws://user:pass@host:port/path)
  -abrp-api-key string      ABRP API key (optional)
  -abrp-vehicle-key string  ABRP vehicle key (optional)
  -device-id string         Device identifier (auto-generated if not set)
  -verbose                  Enable verbose logging
  -discovery-prefix string  HA discovery prefix (default: homeassistant)
```

### Environment Variables
All options can be set via environment variables with a `BYD_HASS_` prefix (e.g., `BYD_HASS_MQTT_URL`).

## Home Assistant Integration

The application automatically discovers and configures a wide array of sensors in Home Assistant. Entities are grouped by category for clarity.

**Available Sensor Categories:**
- **Core:** Speed, Mileage, Gear Position, Power Status
- **Battery & Charging:** State of Charge (SoC), Power Consumption, Temperatures, Charging Status
- **Environment:** Cabin & Outside Temperature
- **Safety & Security:** Seatbelt Status, Lock Status
- **Doors & Windows:** Status for all doors, windows, hood, and trunk
- **Tire Pressure:** Pressure for each tire
- **Lights:** Status for all exterior and interior lights
- **Climate (HVAC):** AC Status, Fan Speed, Blower Mode
- **And many more...** including steering, control, and radar data.

All sensors are created with appropriate device classes and units of measurement.

## Service Management

The installation script sets up a persistent service.

-   **To Stop The Service**: The installer will stop any running instance when it starts. You can re-run the installer and exit it to stop the service.
-   **To View Logs**:
    ```bash
    adb shell tail -f /storage/emulated/0/bydhass/byd-hass.log
    ```
-   **To Restart**: Re-running the installer will restart the service with the new configuration.

The service runs in the background and is managed by a keep-alive script to ensure it restarts automatically if it stops.

## Building from Source

### Build
```bash
./build.sh  # Creates byd-hass-arm64 for Android
```

## Disclaimer

This is unofficial software. Use at your own risk. The author takes no responsibility for any consequences.
