package drawer

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

// Colors
var (
	ColorYellow = color.RGBA{255, 255, 0, 255}
	ColorGreen  = color.RGBA{0, 255, 0, 255}
	ColorBlue   = color.RGBA{0, 0, 255, 255}
	ColorRed    = color.RGBA{255, 0, 0, 255}
)

// Drawer to draw shapes on image
type Drawer struct {
	img image.RGBA
}

// Width of editing image
func (d *Drawer) Width() int {
	return d.img.Bounds().Size().X
}

// Height of editing image
func (d *Drawer) Height() int {
	return d.img.Bounds().Size().Y
}

// Img to get current image
func (d *Drawer) Img() image.RGBA {
	return d.img
}

// JpegBytes returns edited image as jpeg bytes
func (d *Drawer) JpegBytes(quality int) []byte {
	w := new(bytes.Buffer)

	jpeg.Encode(w, &d.img, &jpeg.Options{
		Quality: quality,
	})

	return w.Bytes()
}

// DrawCrosshead on image
func (d *Drawer) DrawCrosshead(x, y, crossheadSize, crossheadStrokeWidth int) {
	imgSize := d.img.Bounds().Size()

	for i := 0; i < crossheadSize; i++ {
		if i > crossheadSize/3 && i < crossheadSize*2/3 {
			continue
		}
		for w := -crossheadStrokeWidth / 2; w < crossheadStrokeWidth/2; w++ {
			xd := x - crossheadSize/2 + i
			if xd >= 0 && xd < imgSize.X && y >= 0 && y < imgSize.Y {
				d.img.Set(xd, y+w, ColorBlue)
			}
			yd := y - crossheadSize/2 + i
			if x >= 0 && x < imgSize.X && yd >= 0 && yd < imgSize.Y {
				d.img.Set(x+w, yd, ColorBlue)
			}
		}
	}
}

// DrawRect on image
func (d *Drawer) DrawRect(x0, y0, x1, y1 int, c color.RGBA) {
	imgSize := d.img.Bounds().Size()

	for x := x0; x <= x1 && x <= imgSize.X; x++ {
		d.img.Set(x, y0, c)
		d.img.Set(x, y1, c)
	}
	for y := y0; y <= y1 && y <= imgSize.Y; y++ {
		d.img.Set(x0, y, c)
		d.img.Set(x1, y, c)
	}
}

// Clone current drawer
func (d *Drawer) Clone() Drawer {
	return Drawer{img: d.CloneImg()}
}

// CloneImg returns a clone of currently edited image
func (d *Drawer) CloneImg() image.RGBA {
	size := d.img.Bounds().Size()

	// create empty image with the same size (to be able to draw)
	newImg := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	// copy source image to the new one
	draw.Draw(newImg, d.img.Bounds(), &d.img, image.ZP, draw.Src)

	return *newImg
}

// ImageRGBAFromJpegBytes creates image.RGBA from jpeg bytes
func ImageRGBAFromJpegBytes(imgBytes []byte) (img image.RGBA, err error) {
	imgReader := bytes.NewReader(imgBytes)

	sourceImg, err := jpeg.Decode(imgReader)
	if err != nil {
		return img, fmt.Errorf("[Drawer] cannot decode jpeg, error: %v", err)
	}

	size := sourceImg.Bounds().Size()

	// create empty image with the same size (to be able to draw)
	newImg := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	// copy source image to the new one
	draw.Draw(newImg, sourceImg.Bounds(), sourceImg, image.ZP, draw.Src)

	return *newImg, err
}

// NewFromJpegBytes creates new Drawer from jpeg image bytes
func NewFromJpegBytes(jpegBytes []byte) (d Drawer, err error) {
	newImg, err := ImageRGBAFromJpegBytes(jpegBytes)
	if err != nil {
		return d, err
	}

	return Drawer{img: newImg}, nil
}

// New creates new Drawer from image.RGBA
func New(img image.RGBA) Drawer {
	return Drawer{img: img}
}
