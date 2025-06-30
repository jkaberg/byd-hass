# BYD-HASS

BYD-HASS is a small Go program that turns data from the BYD "Diplus" API into MQTT messages that Home Assistant can understand, and (optionally) telemetry for A Better Route Planner (ABRP).  It is built as a single static binary so it can run on an Android device under Termux while the car is parked.

## How it works

1. Every 15 seconds the program calls `http://localhost:8988/api/getDiPars` provided by the Diplus app.
2. Values are cached in memory.  Nothing is sent unless a value has changed since the last time it was transmitted.
3. Changed values are published:
   • to MQTT every 60 seconds so that Home Assistant can create sensors automatically (MQTT Discovery).
   • to ABRP every 10 seconds if API and vehicle keys are supplied.

## Quick start (Termux)

```bash
curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/main/install.sh | bash
```

The installer downloads the binary, asks for basic settings, and configures Termux:Boot so the program starts automatically.

Requirements:
- [Diplus app](http://lanye.pw/di/) running and reachable on `localhost:8988`
- [Termux](https://termux.com/) plus the [Termux:Boot](https://github.com/termux/termux-boot) add-on so the program can start automatically
- [Termux:API](https://github.com/termux/termux-api) (for location)
- An MQTT broker – normally the one already used by Home Assistant

---

### Installer script

The same `install.sh` script can be run again later to update the binary or to stop the service if you need to change the configuration.

## Configuration

Settings can be supplied as command-line flags or environment variables (prefix `BYD_HASS_`).

| Flag | Environment variable | Purpose |
| ---- | -------------------- | ------- |
| `-mqtt-url`            | `BYD_HASS_MQTT_URL`          | MQTT connection string (e.g. `ws://user:pass@broker:9001/mqtt`) |
| `-abrp-api-key`        | `BYD_HASS_ABRP_API_KEY`      | ABRP API key (optional) |
| `-abrp-vehicle-key`    | `BYD_HASS_ABRP_VEHICLE_KEY`  | Vehicle identifier used by ABRP (optional) |
| `-device-id`           | `BYD_HASS_DEVICE_ID`         | Unique name for this car (default is auto-generated) |
| `-verbose`             | `BYD_HASS_VERBOSE`           | Enable extra logging |
| `-discovery-prefix`    | ―                            | MQTT discovery prefix (default `homeassistant`) |

## Home Assistant sensors

When connected to MQTT, Home Assistant automatically discovers a single device with many entities such as battery %, speed, mileage, lock state, and more.

![Example sensors in Home Assistant](docs/pictures/mqtt-2025-06-30.png)

## Building from source

```bash
./build.sh   # produces a static arm64 binary for Termux
```

The build script cross-compiles for Android (GOOS=linux GOARCH=arm64 CGO_ENABLED=0) and strips debug symbols for a small footprint.

## Notes

This project is not affiliated with BYD, the Diplus authors, Home Assistant, or ABRP.  Use at your own risk.
