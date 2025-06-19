# byd-hass
Export your BYD car data to Home Assistant

## Disclaimer
This is made available with best intentions, however you're solely responsible for whatever happends (good or bad). The author of this repository takes no reponsibility. 


## Installation

- First you need to be able to sideload apps, there are various methods on how-to do this depeding on Dilink OS version (see youtube or similar for your car)
- Sideload [Diplus](http://lanye.pw/di/), [Termux](https://github.com/termux/termux-app) and [Termux:Boot](https://github.com/termux/termux-boot/) (make sure you start and configure these apps appropriately)
- Launch Termux and run `curl -sSL https://raw.githubusercontent.com/jkaberg/byd-hass/refs/heads/main/install.sh | bash`
- Modify the file `poll_diplus_nohup.sh` and the *Home Assistant config* section

Hopefully if I remembered everything, you should now be all set and good to go!