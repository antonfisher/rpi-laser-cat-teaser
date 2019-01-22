# rpi-laser-cat-teaser

[![Build Status](https://travis-ci.org/antonfisher/rpi-laser-cat-teaser.svg?branch=master)](https://travis-ci.org/antonfisher/rpi-laser-cat-teaser)
[![Go Report Card](https://goreportcard.com/badge/github.com/antonfisher/rpi-laser-cat-teaser)](https://goreportcard.com/report/github.com/antonfisher/rpi-laser-cat-teaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

## Idea

A laser dot runs away from a cat using an RPi camera, a laser, and two servos.

What the program does:
- detects the cat movement using an RPi camera
- calculates required laser dot position to be kept away from the cat
- moves the laser mounted on top of two servos (X and Y axes)
- streams a debug video over HTTP (MJPEG).

## Plan

- [x] assemble servo controller
- [ ] order potentiometer
- [x] control servo software
    - [x] move with servo controller
    - [x] move with RPi PWM (?)
    - [x] calibrate servos
- [ ] laser
    - [ ] movement mechanics
- [x] camera software
    - [x] connect camera to RPi
    - [x] read image
    - [x] movement detection
    - [x] run away algorithm
- [x] pair servo and camera software
- [ ] power
    - [ ] calculation and experiments
    - [ ] order batteries case
- [ ] case design
    - [ ] order case
    - [ ] fit component
- [ ] final assembling

## RPi configuration

```bash
raspi-config
# ---> 5 Interfacing Options
# ---> P1 Camera      Enable/Disable connection to the Raspberry Pi Camera
# ---> Yes
```

## Configuration

Change [main.go](https://github.com/antonfisher/rpi-laser-cat-teaser/blob/master/cmd/main.go#L21).

## Build

```bash
make build # produces build for RPi(ARM)
```

## Run

```bash
bin/rpi-laser-cat-teaser
```
