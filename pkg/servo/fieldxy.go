package servo

import (
	"time"
)

//FieldXY - two-dimensional field operates two servos, one for X, and one for Y axes
type FieldXY struct {
	ServoX         *Servo
	ServoY         *Servo
	FlipHorizontal bool
	FlipVertical   bool
}

// Point moves servos to a single point on the field
func (f *FieldXY) Point(x, y float64) {
	if f.FlipHorizontal {
		x = 1 - x
	}
	if f.FlipVertical {
		y = 1 - y
	}
	f.ServoX.SetPercent(x)
	f.ServoY.SetPercent(y)
}

// Line draws a line on the field (duration: 1s)
func (f *FieldXY) Line(x0, y0, x1, y1 float64) {
	var dX, dY float64
	stepCount := 100
	stepX := (x1 - x0) / float64(stepCount)
	stepY := (y1 - y0) / float64(stepCount)
	for i := 0; i <= stepCount; i++ {
		f.Point(x0+dX, y0+dY)
		dX += stepX
		dY += stepY
		time.Sleep(time.Second / time.Duration(stepCount-1))
	}
}

//Rect draws a rectangle on the field (duration: 4s)
func (f *FieldXY) Rect(x0, y0, x1, y1 float64) {
	f.Line(x0, y0, x0, y1)
	f.Line(x0, y1, x1, y1)
	f.Line(x1, y1, x1, y0)
	f.Line(x1, y0, x0, y0)
}
