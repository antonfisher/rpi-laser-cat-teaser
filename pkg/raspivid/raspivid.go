package raspivid

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	// keep 4 x 3 dimension, otherwise raspivid will crop the image
	defaultWidth  = 4 * 32 // the horizontal resolution is rounded up to the nearest multiple of 32 pixels
	defaultHeight = 3 * 32 // the vertical resolution is rounded up to the nearest multiple of 16 pixels
	defaultFPS    = 15
)

// ImageStream runs `rapsivid` program and returns a stream of pictures (as []byte)
type ImageStream struct {
	Width   int
	Height  int
	FPS     int
	Options []string // any additional options to pass to `rapsivid`
}

func (s *ImageStream) makeOptions() []string {
	if s.FPS == 0 {
		s.FPS = defaultFPS
	}
	if s.Width == 0 {
		s.Width = defaultWidth
	}
	if s.Height == 0 {
		s.Height = defaultHeight
	}

	options := []string{
		"-o", "-", // to write to stdout
		"--codec", "MJPEG", // MJPEG codec for Motion JPEG
		"--width", fmt.Sprint(s.Width), // set image width <size>
		"--height", fmt.Sprint(s.Height), // set image height <size>
		"--framerate", fmt.Sprint(s.FPS), // specify the frames per second to record (FPS)
		"--nopreview",    // do not display a preview window
		"--timeout", "0", // time (in ms) to capture for. If not specified, set to 5s. Zero to disable
		"--flush", // flush buffers in order to decrease latency
	}

	if len(s.Options) > 0 {
		options = append(options, s.Options...)
	}

	return options
}

func (s *ImageStream) parseRaspividOutput(output io.ReadCloser, ch chan []byte) {
	// JPEG SOI marker-|----------|
	var marker = []byte{0xFF, 0xD8, 0xFF, 0xDB, 0x00, 0x84, 0x00}
	var markerLength = len(marker)
	var imagesBuffer = new(bytes.Buffer)

	for {
		// read raspivid output by chunks
		var readBuffer = make([]byte, 4096) //TODO try other values
		n, err := output.Read(readBuffer)   // check n like there: https://golang.org/doc/effective_go.html#defer
		if err != nil {
			fmt.Printf("[raspivid ImageStream] read output error: %s\n", err)
			close(ch)
			break
		}

		for i := 0; i < n; i++ {
			imagesBuffer.WriteByte(readBuffer[i])
			// look for the marker at the end of buffer (ignore the first found marker)
			if bytes.HasSuffix(imagesBuffer.Bytes(), marker) && imagesBuffer.Len() > markerLength {
				// cut off the marker from the end
				imageLength := imagesBuffer.Len() - markerLength
				imageBytes := make([]byte, imageLength)
				copy(imageBytes, imagesBuffer.Bytes()[:imageLength])

				// try to send new found image to the channel
				select {
				case ch <- imageBytes:
				default:
				}

				// reset image buffer and add marker to the beginning (that was cut above)
				imagesBuffer.Reset()
				imagesBuffer.Write(marker)
			}
		}
	}
}

// Start returns a channel of images
func (s *ImageStream) Start() (chan []byte, error) {
	fmt.Printf("[raspivid ImageStream] start...\n")

	options := s.makeOptions() //TODO validate options

	cmd := exec.Command("raspivid", options...) //TODO use context to cancel command
	fmt.Printf("[raspivid ImageStream] command to run: raspivid %s\n", strings.Join(options, " "))

	// log errors to stdout
	cmd.Stderr = os.Stdout

	// pipe raspivid output to parser
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[raspivid ImageStream] piping error: %s\n", err)
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf("[raspivid ImageStream] command starting error: %s\n", err)
		return nil, err
	}

	// channel for images
	ch := make(chan []byte)

	// loop packs images from raspivid stdout and sends them to the channel
	go s.parseRaspividOutput(stdout, ch)

	//TODO cmd.Wait()
	// Wait will close the pipe after seeing the command exit,
	// so most callers need not close the pipe themselves;
	// however, an implication is that it is incorrect to call
	/// Wait before all reads from the pipe have completed.

	return ch, nil
}
