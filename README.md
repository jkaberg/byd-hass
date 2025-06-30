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

## Estimated data usage (Wi-Fi/Cellular)

> The figures below are **ball-park estimates** intended to help you plan for mobile data usage when running `byd-hass` on an always-on Android device.  Actual usage will vary with driving style, connection quality, MQTT broker behaviour, etc.

### How the numbers were derived

1. **Message sizes** – The program currently sends two types of outbound traffic:
   • **MQTT state payload** (`byd_car/<device>/state`).  A full JSON state containing ~20 numeric/boolean fields plus topic and protocol overhead is ≈ **300 bytes** per publish.
   • **ABRP telemetry call** (HTTPS `POST`).  The documented ABRP payload is smaller than the MQTT state but the TLS, HTTP and header overheads are higher.  In practice one update is ≈ **500 bytes** on the wire.
   (MQTT PING packets are only 2 bytes and are ignored here.)
2. **Send intervals** –
   • **MQTT**: every **60 s** *but only while at least one value has changed*.  When the car is parked nothing changes, so the broker typically only sees a retain/heartbeat publish once an hour (Termux network restarts, SOC drift, etc.).  During driving almost every minute triggers an update because speed, mileage, etc. change.
   • **ABRP**: fixed **10 s** interval while driving and completely **disabled while parked**.
3. **Downtime assumption** – Cars spend most of the time parked.  For a "typical commuter" profile we assume **1 h of driving per day** and **23 h parked**.  A pessimistic worst-case and an optimistic best-case are also shown.

### Monthly totals (30-day month)

| Scenario | Driving / day | MQTT | ABRP | Total |
| -------- | ------------- | ---- | ---- | ----- |
| **Typical** (default) | 1 h | 60 msg × 300 B × 30 d = **0.5 MB** | 360 msg × 500 B × 30 d = **5.4 MB** | **≈ 6 MB** |
| Light usage | 30 min | 0.25 MB | 2.7 MB | **≈ 3 MB** |
| Heavy usage | 4 h | 2 MB | 21.6 MB | **≈ 24 MB** |

Even in the heavy-usage scenario the program stays well below 30 MB ⁄ month, which is a tiny fraction of a typical cellular data plan.

*Tip: if you do not need ABRP telemetry you can disable it (omit `-abrp-api-key`) and reduce the data usage by ~90 %.*
