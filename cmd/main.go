package main

import (
	"flag"
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
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/params"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/raspivid"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/servo"
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

func startRaspividStream(w, h, fps int, flipH, flipV bool) (chan []byte, error) {
	options := []string{
		//"--saturation", "-100", // set image saturation (-100 to 100), -100 for grayscale
		//"--annotate", "12", // add timestamp (enable/set annotate flags or text)
	}

	if flipH {
		options = append(options, "--hflip") // set horizontal flip
	}
	if flipV {
		options = append(options, "--vflip") // set vertical flip
	}

	raspividImageStream := &raspivid.ImageStream{
		FPS:     fps,
		Width:   w,
		Height:  h,
		Options: options,
	}

	raspividImageCh, err := raspividImageStream.Start()
	if err != nil {
		return nil, err
	}

	return raspividImageCh, nil
}

func main() {
	var (
		fDebug = flag.Bool("debug", false, "print fps to output")

		fCameraFPS   = flag.Int("camera-fps", params.CameraFPS, "camera fps")
		fCameraFlipH = flag.Bool("camera-flip-h", false, "flip camera image horizontally")
		fCameraFlipV = flag.Bool("camera-flip-v", false, "flip camera image vertical")
		fCameraScale = flag.Int(
			"camera-scale",
			params.CameraScale,
			"camera resolution scale (128*scale x 96*scale)",
		)

		fServoXFlip = flag.Bool("servo-x-flip", false, "flip servo x position calculation")
		fServoYFlip = flag.Bool("servo-y-flip", false, "flip servo y position calculation")
		fServoXMin  = flag.Int(
			"servo-x-min",
			params.ServoXMinAnglePulseLength,
			"servo x min angle pulse length ~[20-120]",
		)
		fServoYMin = flag.Int(
			"servo-y-min",
			params.ServoYMinAnglePulseLength,
			"servo y min angle pulse length ~[20-120]",
		)
		fServoXMax = flag.Int(
			"servo-x-max",
			params.ServoXMaxAnglePulseLength,
			"servo x max angle pulse length ~[20-120]",
		)
		fServoYMax = flag.Int(
			"servo-y-max",
			params.ServoYMaxAnglePulseLength,
			"servo y max angle pulse length ~[20-120]",
		)

		fStream     = flag.Bool("stream", false, "stream debug image")
		fStreamPort = flag.String("stream-port", params.StreamPort, "stream port, url: IP:PORT/stream)")

		fLaserRunAwayRadius = flag.Float64(
			"run-away-radius",
			params.RunAwayRadius,
			"laser run away radius as percent of width [0-1]",
		)
		fFollow = flag.Bool(
			"follow",
			params.AlwaysStayOnRunAwayRadius,
			"laser stays on run away radius",
		)

		fDetectorThreshold = flag.Int(
			"detector-threshold",
			params.DetectorThreshold,
			"detector sensitivity threshold",
		)
		fDetectorBlindSpotRadius = flag.Int(
			"detector-blind-spot-radius",
			params.DetectorBlindSpotRadius,
			"detector blind spot radius (to prevent self-detection)",
		)

		fRandomAmplitude = flag.Float64(
			"ramdom-amplitude",
			params.RandomMovementsAmplitude,
			"laser random movements amplitude [0.005-1]",
		)
		fRandomInterval = flag.Int(
			"random-interval",
			params.RandomMovementsInterval,
			"laser random movements interval in seconds (0 to disable)",
		)

		fVersion = flag.Bool("version", false, "print version")
	)

	flag.Parse()

	if *fVersion {
		fmt.Printf("%s@%s-%s\n", params.Name, params.Version, params.Commit)
		os.Exit(0)
	}

	// prepare RPi GPIO hardware
	err := rpio.Open()
	if err != nil {
		errorAndExit(err)
	}
	defer rpio.Close()

	rpio.StartPwm()
	defer rpio.StopPwm()

	// create servo X
	servoX, err := servo.NewServo(params.ServoXPin, uint32(*fServoXMin), uint32(*fServoXMax))
	if err != nil {
		errorAndExit(err)
	}

	// create servos Y
	servoY, err := servo.NewServo(params.ServoYPin, uint32(*fServoYMin), uint32(*fServoYMax))
	if err != nil {
		errorAndExit(err)
	}

	// create servos XY field
	servoFieldXY := servo.NewFieldXY(servoX, servoY, *fServoXFlip, *fServoYFlip)

	// generate random laser dot movements
	if *fRandomInterval > 0 {
		servoFieldXY.SetRandomMovements(*fRandomAmplitude, time.Second*time.Duration(*fRandomInterval))
	}

	cameraWidth := params.CameraMinWidth * *fCameraScale
	cameraHeight := params.CameraMinHeight * *fCameraScale

	// start raspivid stream
	raspividImageCh, err := startRaspividStream(cameraWidth, cameraHeight, *fCameraFPS, *fCameraFlipH, *fCameraFlipV)
	if err != nil {
		errorAndExit(err)
	}

	// channel of images with detected motion highlighting and current dot position
	debugImageCh := make(chan []byte)

	var lastState LastState

	// read input jpeg stream, move laser dot and send debug image to output stream
	go func() {
		var startTime time.Time
		for {
			startTime = time.Now()

			jpegBytes := <-raspividImageCh
			img, err := drawer.ImageRGBAFromJpegBytes(jpegBytes)
			if err != nil {
				fmt.Println(err)
				return
			}

			lastState.Lock()

			// do not detect laser dot itself
			detectorBlindSpot := &detector.Rect{
				X0: lastState.DotPoint.X - *fDetectorBlindSpotRadius,
				Y0: lastState.DotPoint.Y - *fDetectorBlindSpotRadius,
				X1: lastState.DotPoint.X + *fDetectorBlindSpotRadius,
				Y1: lastState.DotPoint.Y + *fDetectorBlindSpotRadius,
			}

			debugImg, motionPoint := detector.DetectMotion(
				img,
				lastState.Img,
				uint32(*fDetectorThreshold),
				detectorBlindSpot,
			)
			lastState.Img = img

			// move laser dot
			notZeroPoint := motionPoint.X != 0 || motionPoint.Y != 0
			pointMoved := motionPoint.X != lastState.MotionPoint.X || motionPoint.Y != lastState.MotionPoint.Y
			if notZeroPoint && pointMoved {
				lastState.MotionPoint = motionPoint

				motionX := float64(motionPoint.X) / float64(cameraWidth)
				motionY := float64(motionPoint.Y) / float64(cameraHeight)

				// run away from the motion
				servoFieldXY.RunAway(motionX, motionY, *fLaserRunAwayRadius, *fFollow)

				//DEBUG: track to the motion
				//servoFieldXY.LineTo(motionX, motionY)
			}

			// draw debug infomation
			imgDrawer := drawer.New(debugImg)

			// draw blind spot
			imgDrawer.DrawRect(
				detectorBlindSpot.X0,
				detectorBlindSpot.Y0,
				detectorBlindSpot.X1,
				detectorBlindSpot.Y1,
				drawer.ColorGreen,
			)

			// draw current dot position
			imgDrawer.DrawCrosshead(lastState.DotPoint.X, lastState.DotPoint.Y, params.DetectorBlindSpotRadius, 2)

			// draw detected motion
			imgDrawer.DrawRect(
				lastState.MotionPoint.Rect.X0,
				lastState.MotionPoint.Rect.Y0,
				lastState.MotionPoint.Rect.X1,
				lastState.MotionPoint.Rect.Y1,
				drawer.ColorRed,
			)

			lastState.Unlock()

			if *fStream {
				debugImageCh <- imgDrawer.JpegBytes(100)
			}

			if *fDebug {
				fmt.Printf("fps: %5.1f\tframe took: %s\n", 1/time.Since(startTime).Seconds(), time.Since(startTime))
			}
		}
	}()

	// save current laser dot position to draw on debug image
	go func() {
		for {
			p := <-servoFieldXY.CurrentPercentPointCh
			lastState.Lock()
			lastState.DotPoint = image.Point{
				X: int(float64(cameraWidth) * p.X),
				Y: int(float64(cameraHeight) * p.Y),
			}
			lastState.Unlock()
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		//TODO handle this in all goroutines
		<-signalCh

		fmt.Println("Interrupted.")
		wg.Done()
		os.Exit(0)
	}()

	if *fStream {
		streamServer := &mjpeg.Server{
			Addr:      fmt.Sprintf(":%s", *fStreamPort),
			StreamURL: "/stream",
			Source:    debugImageCh,
		}

		// start
		streamServer.ListenAndServe()
	} else {
		wg.Wait()
	}
}
