package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/detector"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/drawer"
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
	streamWidth  = 1 * 4 * 32 // the horizontal resolution is rounded up to the nearest multiple of 32 pixels
	streamHeight = 1 * 3 * 32 // the vertical resolution is rounded up to the nearest multiple of 16 pixels
	streamFPS    = 15
)

// LastState of detector
type LastState struct {
	sync.Mutex

	Img         image.RGBA     // previous analyzed image
	DotPoint    image.Point    // current dot position from servo field controller
	MotionPoint detector.Point // previous detected motion point
}

func errorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
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
			"--hflip", // set horizontal flip
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

	// generate random dot movements
	servoFieldXY.SetRandomMovements(0.01, time.Second*2)

	// start raspivid stream
	raspividImageCh, err := startRaspividStream()
	if err != nil {
		errorAndExit(err)
	}

	// image with detected motion highlighting and current dot position
	debugImageCh := make(chan []byte)

	var lastState LastState

	// read input jpeg stream, move laser dot and send debug image to output stream
	go func() {
		for {
			jpegBytes := <-raspividImageCh
			img, err := drawer.ImageRGBAFromJpegBytes(jpegBytes)
			if err != nil {
				fmt.Println(err)
				return
			}

			lastState.Lock()

			debugImg, motionPoint := detector.DetectMotion(img, lastState.Img)
			lastState.Img = img

			// move laser dot
			notZeroPoint := motionPoint.X != 0 || motionPoint.Y != 0
			pointMoved := motionPoint.X != lastState.MotionPoint.X || motionPoint.Y != lastState.MotionPoint.Y
			if notZeroPoint && pointMoved {
				lastState.MotionPoint = motionPoint

				motionX := float64(motionPoint.X) / float64(streamWidth)
				motionY := float64(motionPoint.Y) / float64(streamHeight)

				// run away from the motion
				servoFieldXY.RunAway(motionX, motionY)

				// track to the motion
				//servoFieldXY.LineTo(motionX, motionY)
			}

			// draw debug infomation
			imgDrawer := drawer.New(debugImg)

			// draw current dot position
			imgDrawer.DrawCrosshead(lastState.DotPoint.X, lastState.DotPoint.Y, 20, 2)

			// draw detected motion
			imgDrawer.DrawRect(
				lastState.MotionPoint.Rect.X0,
				lastState.MotionPoint.Rect.Y0,
				lastState.MotionPoint.Rect.X1,
				lastState.MotionPoint.Rect.Y1,
			)

			lastState.Unlock()

			debugImageCh <- imgDrawer.JpegBytes(100)
		}
	}()

	// save current laser dot position to draw on debug image
	go func() {
		for {
			p := <-servoFieldXY.CurrentPercentPointCh
			lastState.Lock()
			lastState.DotPoint = image.Point{
				X: int(float64(streamWidth) * p.X),
				Y: int(float64(streamHeight) * p.Y),
			}
			lastState.Unlock()
		}
	}()

	streamServer := &mjpeg.Server{
		Addr:      ":8081",
		StreamURL: "/stream",
		Source:    debugImageCh,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		fmt.Println("Interrupted by user...")
		os.Exit(0)
	}()

	// start
	streamServer.ListenAndServe()
}
