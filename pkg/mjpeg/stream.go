package mjpeg

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
)

var mjpegBoundary = "--CUT-HERE"

// Stream is a HTTP handler for MJPEG stream
type Stream struct {
	sync.Mutex
	Source  chan []byte
	clients []chan []byte
}

func (s *Stream) addClient(client chan []byte) {
	s.Lock()
	defer s.Unlock()

	s.clients = append(s.clients, client)
}

func (s *Stream) removeClient(clientToRemove chan []byte) {
	s.Lock()
	defer s.Unlock()

	for i, client := range s.clients {
		if client == clientToRemove {
			s.clients[i] = nil
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

func (s *Stream) logClients() {
	s.Lock()
	defer s.Unlock()

	fmt.Printf("[MJPEG Stream] client count: %d\n", len(s.clients))
}

// Broadcast - broadcasts the stream to clients
func (s *Stream) Broadcast() {
	go func() {
		for {
			image := <-s.Source
			s.Lock()
			for _, updateClientCh := range s.clients {
				select {
				case updateClientCh <- image:
				default:
				}
			}
			s.Unlock()
		}
	}()
}

// HTTPHandler is a handler for HTTP server
func (s *Stream) HTTPHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mjpegBoundary))
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "close") //TODO or "keep-alive"?

	updateClientCh := make(chan []byte)

	s.addClient(updateClientCh)
	s.logClients()

	defer s.logClients()
	defer s.removeClient(updateClientCh)

	// if client closed the connection
	var clientClosed <-chan bool
	if i, ok := res.(http.CloseNotifier); ok {
		clientClosed = i.CloseNotify()
	}

	resBuffer := new(bytes.Buffer)
	for {
		select {
		case <-clientClosed:
			return
		case image := <-updateClientCh:
			// JPEG headers
			fmt.Fprintf(resBuffer, "%s\r\n", mjpegBoundary)
			fmt.Fprint(resBuffer, "Content-Type: image/jpeg\r\n")
			fmt.Fprintf(resBuffer, "Content-Length: %d\r\n", len(image))
			fmt.Fprint(resBuffer, "\r\n")

			// add image
			resBuffer.Write(image)

			// send image with headers
			_, err := resBuffer.WriteTo(res)

			resBuffer.Reset()

			if err != nil { // likely connection is close by client
				return
			}
		}
	}
}

// NewHandler create new HTTP handler
func NewHandler(source chan []byte) *Stream {
	stream := &Stream{
		Source: source,
	}

	stream.Broadcast()

	return stream
}
