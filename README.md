# rpi-laser-cat-teaser

[![Build Status](https://travis-ci.org/antonfisher/rpi-laser-cat-teaser.svg?branch=master)](https://travis-ci.org/antonfisher/rpi-laser-cat-teaser)
[![Go Report Card](https://goreportcard.com/badge/github.com/antonfisher/rpi-laser-cat-teaser)](https://goreportcard.com/report/github.com/antonfisher/rpi-laser-cat-teaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

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
