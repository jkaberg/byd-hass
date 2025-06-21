# byd-hass
Export your BYD car data to Home Assistant

## Disclaimer
This is made available with best intentions, however you're solely responsible for whatever happends (good or bad). The author of this repository takes no reponsibility. 

## What is this?
This is an set of scripts to export information/data from your BYD car to Home Assistant. These data are made available through the app Diplus, for more context/information [see this Github issue](https://github.com/jkaberg/byd-react-app-reverse/issues/2)

## Installation

- First you need to be able to sideload apps, there are various methods on how-to do this depeding on BYD Dilink OS version (see youtube or similar for your car)
- Sideload [Diplus](http://lanye.pw/di/), [Termux](https://github.com/termux/termux-app), [Termux:Boot](https://github.com/termux/termux-boot/) and [Termux:API](https://github.com/termux/termux-api) (make sure you give permissions and configure these apps appropriately)
- Launch Termux and run `curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/refs/heads/main/install.sh | bash`
- Create the file `hass_config` in the `$HOME/scripts` directory and add the following content modifying to your HASS installation
```
HA_BASE_URL="https://HASS-URL"
HA_TOKEN="LONG-LIVED-ACCESS-TOKEN"
```
- Optionally also configure and transmit [ABRP telemetry](https://documenter.getpostman.com/view/7396339/SWTK5a8w#fdb20525-51da-4195-8138-54deabe907d5) (see [docs](https://documenter.getpostman.com/view/7396339/SWTK5a8w#intro) for obtaining API keys etc), in `$HOME/scripts` create the file `abrp_config` with:
```
ABRP_USER_TOKEN=XXXXX
ABRP_API_KEY=XXXX
```

## Features
- Readonly integration with Diplus
- Pushes data to Home Assistant
- Caches data and transmits only on changes (saves bandwith)
- Customizeble (in terms of which sensor data is consumed and transmitted)
- Optional [ABRP](https://documenter.getpostman.com/view/7396339/SWTK5a8w#fdb20525-51da-4195-8138-54deabe907d5) telemetry script to feed ABRP with up to date SOC etc.


## Sensors
- State of charge (Diplus)
- Mileage (Diplus)
- Lock state (Diplus)
- Speed (Diplus)
- Charge gun state (Diplus)
- Position (Termux:API)

## TODO

- [x] Verify that the solution always runs when the car is powered on
- [ ] Run the solution when the car is not running (How does Diplus do it?)
- [ ] Support more sensors
- [ ] Move to MQTT instead of HTTP push
