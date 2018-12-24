package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/servo"
)

var (
	servoPinX = servo.RpiPwmPin12
	servoPinY = servo.RpiPwmPin13
)

func main() {
	fmt.Printf("Run...\n")

	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	rpio.StartPwm()
	defer rpio.StopPwm()

	servoX, err := servo.NewServo(servoPinX)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	servoY, err := servo.NewServo(servoPinY)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	servoX.SetPercent(0.5)
	servoY.SetPercent(0.5)

	// circle
	var r = 0.025
	for n := 0; n < 25; n++ {
		for i := float64(0); i < 2*math.Pi; i += math.Pi / 180 {
			servoX.SetPercent(r*math.Sin(i) + 0.5)
			servoY.SetPercent(r*math.Cos(i) + 0.5)
			time.Sleep(time.Second / 50)
		}
	}

	servoX.SetPercent(0.5)
	servoY.SetPercent(0.5)

	time.Sleep(time.Second * 2)
}
