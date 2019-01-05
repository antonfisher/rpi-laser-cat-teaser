package detector

import (
	"fmt"
	"image"
	"time"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/debug"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/editor"
)

// Rect is a rectangle represented by two points
type Rect struct {
	X0 int
	Y0 int
	X1 int
	Y1 int
}

// Point is a point on XY field
type Point struct {
	X int
	Y int

	// indicates base rectangle for found motion point
	Rect Rect
}

// DetectMotion - takes a channel with images stream and returns:
// - a channel of images with motion indication
// - a channel of XY Points of detected motion
func DetectMotion(inputStreamCh chan []byte, cancelCh chan bool) (chan []byte, chan Point) {
	outputStreamCh := make(chan []byte)
	motionPointsCh := make(chan Point)

	go func() {
		var prevImage image.Image
		var diffArray [][]int
		var lastDetectedPoint Point
	Loop:
		for {
			select {
			case <-cancelCh:
				break Loop
			case img := <-inputStreamCh:
				startTime := time.Now()
				imageEditor, err := editor.NewEditorFromJpegBytes(img)
				if err != nil {
					fmt.Println(err)
					continue
				}
				imageBeforeEdit := imageEditor.Clone()
				if prevImage != nil {
					diffArray = imageEditor.DiffGreen(prevImage, uint32(7500))
					detectedPoint := findCenter(diffArray)
					if detectedPoint.X > 0 || detectedPoint.Y > 0 {
						lastDetectedPoint = detectedPoint
						//fmt.Printf("detected point: %+v\n", detectedPoint)
						motionPointsCh <- detectedPoint
					}
					imageEditor.DrawCrosshead(lastDetectedPoint.X, lastDetectedPoint.Y, 20, 2)
					imageEditor.DrawRect(
						lastDetectedPoint.Rect.X0,
						lastDetectedPoint.Rect.Y0,
						lastDetectedPoint.Rect.X1,
						lastDetectedPoint.Rect.Y1,
					)
				}
				outputStreamCh <- imageEditor.JpegBytes(90)
				prevImage = imageBeforeEdit
				debug.LogExecutionTime("motion detection", startTime)
			}
		}
	}()

	return outputStreamCh, motionPointsCh
}

//findCenter of binary presented shape
func findCenter(a [][]int) Point {
	var x0, y0, x1, y1 int

	for x, row := range a {
		for y := range row {
			if a[x][y] == 1 {
				if x0 == 0 {
					x0 = x
				}
				if y0 == 0 {
					y0 = y
				}
				if x1 == 0 || x > x1 {
					x1 = x
				}
				if y1 == 0 || y > y1 {
					y1 = y
				}
			}
		}
	}

	return Point{
		X: x0 + (x1-x0)/2,
		Y: y0 + (y1-y0)/2,
		Rect: Rect{
			X0: x0,
			Y0: y0,
			X1: x1,
			Y1: y1,
		},
	}
}
