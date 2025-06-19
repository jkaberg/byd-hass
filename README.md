# byd-hass
Export your BYD car data to Home Assistant

## Disclaimer
This is made available with best intentions, however you're solely responsible for whatever happends (good or bad). The author of this repository takes no reponsibility. 

## What is this?
This is an set of scripts to export information/data from your BYD car to Home Assistant. These data are made available through the app Diplus, for more context/information [see this Github issue](https://github.com/jkaberg/byd-react-app-reverse/issues/2)

## Installation

- First you need to be able to sideload apps, there are various methods on how-to do this depeding on Dilink OS version (see youtube or similar for your car)
- Sideload [Diplus](http://lanye.pw/di/), [Termux](https://github.com/termux/termux-app) and [Termux:Boot](https://github.com/termux/termux-boot/) (make sure you start and configure these apps appropriately)
- Launch Termux and run `curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/refs/heads/main/install.sh | bash`
- Create the file `hass_config` in the `$HOME/scripts` directory and add the following content modifying to your HASS installation
```
HA_BASE_URL="https://HASS-URL"
HA_TOKEN="LONG-LIVED-ACCESS-TOKEN"
```

Hopefully if I remembered everything, you should now be all set and good to go!

## Features
- Readonly integration with Diplus
- Caches data and transmits only on changes (saves bandwith)
- Customizeble (in terms of which sensor data is consumed and transmitted)

## TODO

- [ ] Verify that the project always runs at boot (eg when the car is powered on)
- [ ] Run the solution when the car is not running (How does Diplus do it?)
- [ ] Support more sensors