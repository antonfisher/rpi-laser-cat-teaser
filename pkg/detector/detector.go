package detector

import (
	"image"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/drawer"
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

// DetectMotion takes a channel with image.RGBA stream and
// returns a channel of XY Points of detected motion
func DetectMotion(img, previousImg image.RGBA, threshold uint32, blindSpot *Rect) (
	debugImg image.RGBA,
	motionPoint Point,
) {
	imgDrawer := drawer.New(img)
	debugImg = imgDrawer.CloneImg()

	// images size
	w := imgDrawer.Width()
	h := imgDrawer.Height()

	//TODO use struct with one array underlines slices:
	// https://golang.org/doc/effective_go.html#two_dimensional_slices
	diffArray := make([][]int, w)
	for i := range diffArray {
		diffArray[i] = make([]int, h)
	}

	// calculate difference between images based on green channel
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			//TODO what color channel to use? Detector still can be self-triggers by laser dot
			_, g1, _, _ := debugImg.At(x, y).RGBA()
			_, g2, _, _ := previousImg.At(x, y).RGBA()
			gDiff := absUInt32Diff(g1, g2)
			insideBlindSpot := blindSpot.X0 < x && x < blindSpot.X1 && blindSpot.Y0 < y && y < blindSpot.Y1
			if gDiff > threshold && !insideBlindSpot {
				debugImg.Set(x, y, drawer.ColorYellow)
				diffArray[x][y] = 1
				// } else {
				// 	d.img.Set(x, y, &color.RGBA{0, uint8(g2), 0, 255})
			}
		}
	}

	motionPoint = findCenterPoint(diffArray)

	return
}

func absUInt32Diff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

//findCenterPoint of binary presented shape
func findCenterPoint(a [][]int) Point {
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
