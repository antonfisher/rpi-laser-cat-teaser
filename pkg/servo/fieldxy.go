package servo

import (
	"sync"
	"time"
)

//PercentPoint - a point percent values of XY
type PercentPoint struct {
	X float64
	Y float64
}

//FieldXY is a two-dimensional field that controls two servos (one for X, and one for Y axes)
type FieldXY struct {
	ServoX         *Servo
	ServoY         *Servo
	FlipHorizontal bool
	FlipVertical   bool

	sync.Mutex
	currentX                float64
	currentY                float64
	cancelCurrentMovementCh chan bool
}

// SetPoint moves servos to a single point on the field
func (f *FieldXY) SetPoint(x, y float64) {
	f.Lock()
	f.currentX = x
	f.currentY = y
	f.Unlock()

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
	if f.cancelCurrentMovementCh != nil {
		//fmt.Println("send cancel ->")
		f.cancelCurrentMovementCh <- true
	}
	f.Lock()
	f.cancelCurrentMovementCh = make(chan bool)
	f.Unlock()

	go func(x0, y0, x1, y1 float64) {
		stepCount := 100
		stepX := (x1 - x0) / float64(stepCount)
		stepY := (y1 - y0) / float64(stepCount)
		step := 0
		ticker := time.NewTicker(time.Second / 2 / time.Duration(stepCount-1))
		var dX, dY float64
		for {
			select {
			case <-f.cancelCurrentMovementCh:
				//fmt.Println("cancel movement...")
				return
			case <-ticker.C:
				f.SetPoint(x0+dX, y0+dY)
				dX += stepX
				dY += stepY
				step++
				if step == stepCount {
					ticker.Stop()
					f.Lock()
					f.cancelCurrentMovementCh = nil
					f.Unlock()
					return
				}
			}
		}
	}(x0, y0, x1, y1)
}

// // Rect draws a rectangle on the field (duration: 4s)
// func (f *FieldXY) Rect(x0, y0, x1, y1 float64) {
// 	f.Line(x0, y0, x0, y1)
// 	f.Line(x0, y1, x1, y1)
// 	f.Line(x1, y1, x1, y0)
// 	f.Line(x1, y0, x0, y0)
// }

// LineTo - smooth movement to the point from current position
func (f *FieldXY) LineTo(x, y float64) {
	go func(x, y float64) {
		f.Lock()
		currentX := f.currentX
		currentY := f.currentY
		f.Unlock()

		//fmt.Printf("line from %v\t%v to \t%v\t%v\n", currentX, currentY, x, y)
		f.Line(currentX, currentY, x, y)
	}(x, y)
}

// NewFieldXY creates new FieldXY
func NewFieldXY(servoX, servoY *Servo, flipHorizontal, flipVertical bool) *FieldXY {
	return &FieldXY{
		ServoX:                  servoX,
		ServoY:                  servoY,
		FlipHorizontal:          flipHorizontal,
		FlipVertical:            flipVertical,
		cancelCurrentMovementCh: nil,
	}
}
