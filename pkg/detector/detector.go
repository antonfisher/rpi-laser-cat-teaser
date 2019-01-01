package detector

import (
	"fmt"
	"image"
	"time"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/debug"
	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/editor"
)

// Point is a point on XY field
type Point struct {
	X int
	Y int
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
		var x, y int
		for {
			select {
			case <-cancelCh:
				break
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
					xDetected, yDetected := findCenter(diffArray)
					if xDetected > 0 || yDetected > 0 {
						x = xDetected
						y = yDetected
					}
					imageEditor.DrawCrosshead(x, y, 20, 2)
					motionPointsCh <- Point{X: x, Y: y}
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
func findCenter(a [][]int) (int, int) {
	var cx, cy, stepsPerformed int
	var w = len(a)
	var h = len(a[0])

	aCopy := make([][]int, w)
	for i := range aCopy {
		aCopy[i] = make([]int, h)
	}

	for {
		stepsPerformed++

		for x, row := range a {
			for y, v := range row {
				aCopy[x][y] = v
			}
		}

		var stepMax int
		for x, row := range a {
			for y := range row {
				neighbours := [][]int{
					[]int{x - 1, y - 1},
					[]int{x - 1, y},
					[]int{x - 1, y + 1},
					[]int{x, y + 1},
					[]int{x, y - 1},
					[]int{x + 1, y - 1},
					[]int{x + 1, y},
					[]int{x + 1, y + 1},
				}
				var sum int
				for _, n := range neighbours {
					xn := n[0]
					yn := n[1]
					if xn > 0 && yn > 0 && xn < w && yn < h {
						sum += aCopy[xn][yn]
					}
				}
				if sum < 8 {
					a[x][y] = 0
				}
				if sum > stepMax {
					stepMax = sum
					cx = x
					cy = y
				}
			}
		}
		if stepMax < 8 {
			break
		}
	}

	fmt.Printf("steps performed: %v\t", stepsPerformed)

	return cx, cy
}
