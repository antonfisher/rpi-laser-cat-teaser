package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"time"

	"github.com/stianeikeland/go-rpio"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/debug"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/detector"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/editor"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/mjpeg"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/raspivid"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/servo"
)

// application config
var (
	servoPinX                        = servo.RpiPwmPin12
	servoXMinAnglePulseLength uint32 = 74 // tested camera angle min (tested servo min: 49)  [right]
	servoXMaxAnglePulseLength uint32 = 97 // tested camera angle max (tested servo max: 114) [left ]

	servoPinY                        = servo.RpiPwmPin13
	servoYMinAnglePulseLength uint32 = 56 // tested camera angle min (tested servo min: 28)  [down]
	servoYMaxAnglePulseLength uint32 = 75 // tested camera angle max (tested servo max: 104) [up  ]
)

func main() {
	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	rpio.StartPwm()
	defer rpio.StopPwm()

	servoFieldXY, err := createServoFieldXY()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// go func() {
	// 	for {
	// 		servoFieldXY.Rect(0, 0, 1, 1)
	// 	}
	// }()

	raspividImageStream := &raspivid.ImageStream{
		//FPS: 10,
		Options: []string{
			"--vflip", // set vertical flip
			"--hflip", // set horisontal flip
			//"--saturation", "-100", // set image saturation (-100 to 100), -100 for grayscale
			//"--annotate", "12", // add timestamp (enable/set annotate flags or text)
		},
	}

	raspividImageCh, err := raspividImageStream.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	middleware := make(chan []byte)
	go func() {
		var prevImage image.Image
		var i int
		var diffArray [][]int
		var x, y int
		for {
			startTime := time.Now()
			img := <-raspividImageCh
			imageEditor, err := editor.NewEditorFromJpegBytes(img)
			if err != nil {
				fmt.Println(err)
				continue
			}
			imageBeforeEdit := imageEditor.Clone()
			if prevImage != nil {
				diffArray = imageEditor.DiffGreen(prevImage, uint32(7500))
				xDetected, yDetected := detector.FindCenter(diffArray)
				if xDetected > 0 || yDetected > 0 {
					x = xDetected
					y = yDetected
				}
				imageEditor.DrawCrosshead(x, y, 20, 2)

				//servo
				servoFieldXY.Point(float64(x)/128, float64(y)/96)
			}
			middleware <- imageEditor.JpegBytes(90)
			prevImage = imageBeforeEdit
			i++
			debug.LogExecutionTime(fmt.Sprintf("image: %v", i), startTime)
		}
	}()

	streamServer := &mjpeg.Server{
		Addr:      ":8081",
		StreamURL: "/stream",
		Source:    middleware,
		//Source: raspividImageCh,
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

	return &servo.FieldXY{
		ServoX:         servoX,
		ServoY:         servoY,
		FlipHorizontal: true,
		FlipVertical:   true,
	}, nil
}
