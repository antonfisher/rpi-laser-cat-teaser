package servo

import (
	"fmt"

	"github.com/stianeikeland/go-rpio"
)

// Servo controller
//
// Before usage open rpio and start PWM:
// 	rpio.Open()
// 	defer rpio.Close()
// 	rpio.StartPwm()
// 	defer rpio.StopPwm()
//

// RpiPwmPin - Raspberry PWM GPIO pin number
type RpiPwmPin uint8

var (
	// RpiPwmPin12 - channel 1 (pwm0) for pin 12
	RpiPwmPin12 RpiPwmPin = 12

	// RpiPwmPin13 - channel 2 (pwm1) for pin 13
	RpiPwmPin13 RpiPwmPin = 13
)

// defaults
var (
	// DefaultCycle - default PMW cycle length
	DefaultCycle uint32 = 128000

	// DefaultMinAnglePulseLength [1-999]
	// Default pulse length to set servo to its min angle, ex. 27 is 27/1000 = 2.7%
	DefaultMinAnglePulseLength uint32 = 27

	// DefaultMaxAnglePulseLength [1-999]
	// Default pulse length to set servo to its max angle, ex. 114 is 114/1000 = 11.4%
	DefaultMaxAnglePulseLength uint32 = 114
)

// Servo controls a servo connected to the PWM pin
type Servo struct {
	//Pin is
	Pin RpiPwmPin

	// Frequency/period are specific to controlling a specific servo.
	// A typical servo motor expects to be updated every 20 ms with
	// a pulse between 1 ms and 2 ms, or in other words, between
	// a 5 and 10% duty cycle on a 50 Hz waveform.
	// With a 1.5 ms pulse, the servo motor will be at the natural
	// 90 degree position.
	// With a 1 ms pulse, the servo will be at the 0 degree position,
	// and with a 2 ms pulse, the servo will be at 180 degrees.
	// You can obtain the full range of motion by updating the servo
	// with an value in between.

	// DefaultMinAnglePulseLength [1-999]
	// Default pulse length to set servo to its min angle, ex. 27 is 27/1000 = 2.7%
	MinAnglePulseLength uint32

	// DefaultMaxAnglePulseLength [1-999]
	// Default pulse length to set servo to its max angle, ex. 114 is 114/1000 = 11.4%
	MaxAnglePulseLength uint32

	rpioPin  rpio.Pin
	pwmCycle uint32
}

var l uint32

// SetPercent - set servo angle in percent [0.0-1.0]
func (s *Servo) SetPercent(val float64) {
	// calculate available duty range
	rangeFrom := s.pwmCycle * s.MinAnglePulseLength / 1000
	rangeTo := s.pwmCycle * s.MaxAnglePulseLength / 1000

	duty := rangeFrom + ((rangeTo-rangeFrom)*uint32(val*10000))/10000
	s.rpioPin.DutyCycle(duty, s.pwmCycle)
}

// NewServo - create new servo controller
func NewServo(pin RpiPwmPin, minAnglePulseLength, maxAnglePulseLength uint32) (*Servo, error) {
	if pin != RpiPwmPin12 && pin != RpiPwmPin13 {
		return nil, fmt.Errorf("Pin '%v' cannot be used for servo, use 12 or 13", pin)
	}

	servoPin := rpio.Pin(pin)
	servoPin.Mode(rpio.Pwm)

	fmt.Printf("new servo: %v %v %v\n", pin, minAnglePulseLength, maxAnglePulseLength)

	servo := &Servo{
		Pin:                 pin,
		MinAnglePulseLength: minAnglePulseLength,
		MaxAnglePulseLength: maxAnglePulseLength,
		rpioPin:             rpio.Pin(pin),
		pwmCycle:            DefaultCycle,
	}

	servo.rpioPin.Freq(50 * int(servo.pwmCycle))

	return servo, nil
}
