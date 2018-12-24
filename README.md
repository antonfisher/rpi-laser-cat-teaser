# rpi-laser-cat-teaser

## Plan

- [x] assemble servo controller
- [ ] order potentiometer
- [ ] control servo software
    - [x] move with servo controller
    - [x] move with RPi PWM (?)
    - [ ] calibrate servos
- [ ] laser
    - [ ] movement mechanics
- [ ] camera software
    - [ ] connect camera to RPi
    - [ ] read image
    - [ ] movement detection
    - [x] run away algorithm
- [ ] pair servo and camera software
- [ ] power
    - [ ] calculation and experiments
    - [ ] order batteries case
- [ ] case design
    - [ ] order case
    - [ ] fit component
- [ ] final assembling

## RPi configuration

```bash
apt install -y git vim htop

raspi-config
# ---> 5 Interfacing Options
# ---> P1 Camera      Enable/Disable connection to the Raspberry Pi Camera
# ---> Yes
```
