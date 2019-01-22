package servo

import (
	"fmt"
	"math"
	"sync"
	"time"
)

//const keepAwayR = 0.45
const keepAwayR = 0.5

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

	CurrentPercentPointCh chan PercentPoint

	sync.Mutex
	currentX         float64
	currentY         float64
	cancelNoiseCh    chan struct{}
	cancelMovementCh chan struct{}
}

// SetPoint moves servos to a single point on the field
func (f *FieldXY) SetPoint(x, y float64) {
	f.Lock()
	f.currentX = x
	f.currentY = y
	f.Unlock()

	select {
	case f.CurrentPercentPointCh <- PercentPoint{
		X: x,
		Y: y,
	}:
	default:
	}

	if f.FlipHorizontal {
		x = 1 - x
	}
	if f.FlipVertical {
		y = 1 - y
	}
	f.ServoX.SetPercent(x)
	f.ServoY.SetPercent(y)
}

// Line draws a line on the field
func (f *FieldXY) Line(x0, y0, x1, y1 float64) {
	// cancel previous movement if still running
	if f.cancelMovementCh != nil {
		select {
		case f.cancelMovementCh <- struct{}{}:
		default:
		}
	}

	go func(x0, y0, x1, y1 float64) {
		f.Lock()
		f.cancelMovementCh = make(chan struct{})
		f.Unlock()
		defer func() {
			f.Lock()
			f.cancelMovementCh = nil
			f.Unlock()
		}()

		stepCount := int(distance(x0, y0, x1, y1) * 100)
		if stepCount < 0 {
			stepCount *= -1
		}
		if stepCount < 5 {
			stepCount = 5
		}
		fmt.Printf("steps: %v\n", stepCount)
		stepX := (x1 - x0) / float64(stepCount)
		stepY := (y1 - y0) / float64(stepCount)
		step := 0
		ticker := time.NewTicker(time.Second / 4 / time.Duration(stepCount-1))
		defer ticker.Stop()
		var dX, dY float64
		for {
			select {
			case <-f.cancelMovementCh:
				//ticker.Stop()
				fmt.Println("movement canceled")
				return
			case <-ticker.C:
				f.SetPoint(x0+dX, y0+dY)
				dX += stepX
				dY += stepY
				step++
				if step == stepCount {
					//ticker.Stop()
					return
				}
			}
		}
	}(x0, y0, x1, y1)
}

// LineTo - smooth movement to the point from current position
func (f *FieldXY) LineTo(x, y float64) {
	f.Lock()
	currentX := f.currentX
	currentY := f.currentY
	f.Unlock()

	f.Line(currentX, currentY, x, y)
}

// RunAway from the point
func (f *FieldXY) RunAway(x, y float64) {
	f.Lock()
	dotX := f.currentX * 4
	dotY := f.currentY * 3
	f.Unlock()

	const W = 1.0 * 4
	const H = 1.0 * 3

	x = x * 4
	y = y * 3

	keepAwayR := keepAwayR * 4

	// closest point from the current laser position to the "keep away" circle
	kaX := x + (keepAwayR*(dotX-x))/math.Sqrt(math.Pow(dotX-x, 2)+math.Pow(dotY-y, 2))
	kaY := y + (keepAwayR*(dotY-y))/math.Sqrt(math.Pow(dotX-x, 2)+math.Pow(dotY-y, 2))

	if distance(x, y, dotX, dotY) < distance(x, y, kaX, kaY) {
		dotX = kaX
		dotY = kaY
	}

	// pushed out of canvas
	if dotX < 0 || dotX > W || dotY < 0 || dotY > H {
		intersections := [][]float64{}
		// pushed to the top
		if y-keepAwayR < 0 {
			dx := math.Sqrt(math.Pow(keepAwayR, 2) - math.Pow(y, 2))
			if 0 <= x+dx && x+dx <= W {
				intersections = append(intersections, []float64{x + dx, 0})
			}
			if 0 <= x-dx && x-dx <= W {
				intersections = append(intersections, []float64{x - dx, 0})
			}
		}
		// pushed to the bottom
		if y+keepAwayR > H {
			dx := math.Sqrt(math.Pow(keepAwayR, 2) - math.Pow(H-y, 2))
			if 0 <= x+dx && x+dx <= W {
				intersections = append(intersections, []float64{x + dx, H})
			}
			if 0 <= x-dx && x-dx <= W {
				intersections = append(intersections, []float64{x - dx, H})
			}
		}
		// pushed to the left
		if x-keepAwayR < 0 {
			dy := math.Sqrt(math.Pow(keepAwayR, 2) - math.Pow(x, 2))
			if 0 <= y+dy && y+dy <= H {
				intersections = append(intersections, []float64{0, y + dy})
			}
			if 0 <= y-dy && y-dy <= H {
				intersections = append(intersections, []float64{0, y - dy})
			}
		}
		// pushed to the right
		if x+keepAwayR > W {
			dy := math.Sqrt(math.Pow(keepAwayR, 2) - math.Pow(W-x, 2))
			if 0 <= y+dy && y+dy <= H {
				intersections = append(intersections, []float64{W, y + dy})
			}
			if 0 <= y-dy && y-dy <= H {
				intersections = append(intersections, []float64{W, y - dy})
			}
		}

		if len(intersections) > 0 {
			minDistance := keepAwayR
			closestPoint := intersections[0]
			for _, v := range intersections {
				d := distance(dotX, dotY, v[0], v[1])
				if d < minDistance {
					minDistance = d
					closestPoint = v
				}
			}
			dotX = closestPoint[0]
			dotY = closestPoint[1]
		}
	}

	//f.SetPoint(dotX/4, dotY/3)
	f.LineTo(dotX/4, dotY/3)
}

func distance(x0, y0, x1, y1 float64) float64 {
	return math.Sqrt(math.Pow(x0-x1, 2) + math.Pow(y0-y1, 2))
}

// SetRandomMovements - move dot a little bit
// step=0 to disable random movements
func (f *FieldXY) SetRandomMovements(step float64, interval time.Duration) {
	// cancel previous noise motions if running
	select {
	case f.cancelNoiseCh <- struct{}{}:
	default:
	}

	if step > 0 {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for {
				select {
				case <-f.cancelNoiseCh:
					fmt.Println("noise canceled")
					ticker.Stop()
					return
				case <-ticker.C:
					f.MoveRandom(step)
				}
			}
		}()
	}
}

// MoveRandom - move
func (f *FieldXY) MoveRandom(step float64) {
	f.Lock()
	x := f.currentX
	y := f.currentY
	f.Unlock()

	if x+step > 1 {
		x -= step
	} else if x-step < 0 {
		x += step
	} else if time.Now().Nanosecond()%2 == 0 {
		x += step
	} else {
		x -= step
	}

	if y+step > 1 {
		y -= step
	} else if y-step < 0 {
		y += step
	} else if time.Now().Nanosecond()%2 == 0 {
		y += step
	} else {
		y -= step
	}

	f.SetPoint(x, y)
}

// NewFieldXY creates new FieldXY
func NewFieldXY(servoX, servoY *Servo, flipHorizontal, flipVertical bool) *FieldXY {
	return &FieldXY{
		ServoX:                servoX,
		ServoY:                servoY,
		FlipHorizontal:        flipHorizontal,
		FlipVertical:          flipVertical,
		CurrentPercentPointCh: make(chan PercentPoint),
		cancelNoiseCh:         make(chan struct{}),
	}
}
