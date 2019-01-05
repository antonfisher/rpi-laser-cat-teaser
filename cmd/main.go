package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/stianeikeland/go-rpio"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/detector"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/mjpeg"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/raspivid"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/servo"
)

// application config
var (
	// servo X
	servoPinX                        = servo.RpiPwmPin12
	servoXMinAnglePulseLength uint32 = 74 // tested camera angle min (tested servo min: 49)  [right]
	servoXMaxAnglePulseLength uint32 = 97 // tested camera angle max (tested servo max: 114) [left ]

	// servo Y
	servoPinY                        = servo.RpiPwmPin13
	servoYMinAnglePulseLength uint32 = 56 // tested camera angle min (tested servo min: 28)  [down]
	servoYMaxAnglePulseLength uint32 = 75 // tested camera angle max (tested servo max: 104) [up  ]

	// raspivid stream
	// keep 4 x 3 dimension, otherwise raspivid will crop the image
	streamWidth  = 4 * 32 // the horizontal resolution is rounded up to the nearest multiple of 32 pixels
	streamHeight = 3 * 32 // the vertical resolution is rounded up to the nearest multiple of 16 pixels
	streamFPS    = 15
)

func errorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	// prepare RPi GPIO hardware
	err := rpio.Open()
	if err != nil {
		errorAndExit(err)
	}
	defer rpio.Close()

	rpio.StartPwm()
	defer rpio.StopPwm()

	// create servos XY field
	servoFieldXY, err := createServoFieldXY()
	if err != nil {
		errorAndExit(err)
	}

	// start raspivid stream
	raspividImageCh, err := startRaspividStream()
	if err != nil {
		errorAndExit(err)
	}

	// detect motion in stream, move laser dot
	cancelCh := make(chan bool)
	detectorImageCh, motionPointsCh := detector.DetectMotion(raspividImageCh, cancelCh)

	// move laser dot
	go func() {
		var previousPoint detector.Point
		for {
			point := <-motionPointsCh
			if point.X != previousPoint.X || point.Y != previousPoint.Y {
				previousPoint = point
				fmt.Printf("move to: %v %v\n", point.X, point.Y)
				servoFieldXY.LineTo(float64(point.X)/float64(streamWidth), float64(point.Y)/float64(streamHeight))
			}
		}
	}()

	streamServer := &mjpeg.Server{
		Addr:      ":8081",
		StreamURL: "/stream",
		Source:    detectorImageCh,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		//raspividImageStream.Stop()
		//streamServer.Stop()
		fmt.Println("Interrupted by user...")
		os.Exit(0)
	}()

	// start
	streamServer.ListenAndServe()
}

func createServoFieldXY() (*servo.FieldXY, error) {
	servoX, err := servo.NewServo(servoPinX, servoXMinAnglePulseLength, servoXMaxAnglePulseLength)
	if err != nil {
		return nil, err
	}

	servoY, err := servo.NewServo(servoPinY, servoYMinAnglePulseLength, servoYMaxAnglePulseLength)
	if err != nil {
		return nil, err
	}

	return servo.NewFieldXY(servoX, servoY, true, true), nil
}

func startRaspividStream() (chan []byte, error) {
	raspividImageStream := &raspivid.ImageStream{
		FPS:    streamFPS,
		Width:  streamWidth,
		Height: streamHeight,
		Options: []string{
			"--vflip", // set vertical flip
			"--hflip", // set horisontal flip
			//"--saturation", "-100", // set image saturation (-100 to 100), -100 for grayscale
			//"--annotate", "12", // add timestamp (enable/set annotate flags or text)
		},
	}

	raspividImageCh, err := raspividImageStream.Start()
	if err != nil {
		return nil, err
	}

	return raspividImageCh, nil
}
