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

## RPi configuration

```bash
raspi-config
# ---> 5 Interfacing Options
# ---> P1 Camera      Enable/Disable connection to the Raspberry Pi Camera
# ---> Yes
```

## Build

```bash
make build # produces build for RPi(ARM)
```

## Run

```bash
bin/rpi-laser-cat-teaser
```

Options:

```bash
$ ./bin/rpi-laser-cat-teaser --help
Usage of ./bin/rpi-laser-cat-teaser:
  -camera-flip-h
    	flip camera image horizontally
  -camera-flip-v
    	flip camera image vertical
  -camera-fps int
    	camera fps (default 24)
  -camera-scale int
    	camera resolution scale (128*scale x 96*scale) (default 1)
  -debug
    	print fps to output
  -detector-blind-spot-radius int
    	detector blind spot radius (to prevent self-detection) (default 10)
  -detector-threshold int
    	detector sensitivity threshold (default 7500)
  -follow
    	laser stays on run away radius
  -ramdom-amplitude float
    	laser random movements amplitude [0.005-1] (default 0.02)
  -random-interval int
    	laser random movements interval in seconds (0 to disable) (default 2)
  -run-away-radius float
    	laser run away radius as percent of width [0-1] (default 0.5)
  -servo-x-flip
    	flip servo x position calculation
  -servo-x-max int
    	servo x max angle pulse length ~[20-120] (default 97)
  -servo-x-min int
    	servo x min angle pulse length ~[20-120] (default 74)
  -servo-y-flip
    	flip servo y position calculation
  -servo-y-max int
    	servo y max angle pulse length ~[20-120] (default 75)
  -servo-y-min int
    	servo y min angle pulse length ~[20-120] (default 56)
  -stream
    	stream debug image
  -stream-port string
    	stream port, url: IP:PORT/stream) (default "8081")
  -version
    	print version
```

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
