package params

import (
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/servo"
)

// LDFLAGS build properties
var (
	// Name - application name
	Name string

	// Version, to set use flag:
	// go build -ldflags "-X github.com/antonfisher/rpi-laser-cat-teaser/cmd/main.Version=..."
	Version string

	// Commit, to set use flag:
	// go build -ldflags "-X github.com/antonfisher/rpi-laser-cat-teaser/cmd/main.Commit=..."
	Commit string
)

// default application config
var (
	// servo X
	ServoXPin                 = servo.RpiPwmPin12
	ServoXMinAnglePulseLength = 74 // tested camera angle min (tested servo min: 49)  [right]
	ServoXMaxAnglePulseLength = 97 // tested camera angle max (tested servo max: 114) [left ]

	// servo Y
	ServoYPin                 = servo.RpiPwmPin13
	ServoYMinAnglePulseLength = 56 // tested camera angle min (tested servo min: 28)  [down]
	ServoYMaxAnglePulseLength = 75 // tested camera angle max (tested servo max: 104) [up  ]

	// rpi camera (raspivid stream)
	// keep 4 x 3 dimension, otherwise raspivid will crop the image
	CameraMinWidth  = 1 * 4 * 32 // the horizontal resolution is rounded up to the nearest multiple of 32 pixels
	CameraMinHeight = 1 * 3 * 32 // the vertical resolution is rounded up to the nearest multiple of 16 pixels
	CameraScale     = 1
	CameraFPS       = 24

	// motion detector
	DetectorThreshold       = 7500 // color difference sensitivity
	DetectorBlindSpotRadius = 10   // blind radius to prevent self-detection

	// run-away algorithm
	RunAwayRadius             = 0.5   // as percent of view area width
	AlwaysStayOnRunAwayRadius = false // run after motion if it's futher then run-away radius

	// random dot movements
	RandomMovementsAmplitude = 0.02 // as percent of view area width
	RandomMovementsInterval  = 2    // * time.Second

	// debug image stream
	StreamPort = "8081"
)
