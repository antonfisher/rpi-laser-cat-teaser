package detector

import (
	"image"
	"time"

	"github.com/antonfisher/rpi-laser-cat-teaser/pkg/debug"
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
func DetectMotion(img, previousImg image.RGBA, blindSpot *Rect) (debugImg image.RGBA, motionPoint Point) {
	defer debug.LogExecutionTime("motion detection", time.Now())

	imgDrawer := drawer.New(img)
	debugImgDrawer := imgDrawer.Clone()                              //TODO is clone needed here?
	diffArray := debugImgDrawer.DiffGreen(previousImg, uint32(7500)) //TODO move to function arguments

	if blindSpot != nil && len(diffArray) > 0 {
		for x := blindSpot.X0; x <= blindSpot.X1; x++ {
			for y := blindSpot.Y0; y <= blindSpot.Y1; y++ {
				if x >= 0 && x < len(diffArray) && y >= 0 && y < len(diffArray[0]) {
					diffArray[x][y] = 0
				}
			}
		}
	}

	debugImg = debugImgDrawer.Img()
	motionPoint = findCenterPoint(diffArray)

	return
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
