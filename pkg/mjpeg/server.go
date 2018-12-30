package mjpeg

import (
	"fmt"
	"net/http"
)

// Server streams MJPEG video to clients from specified source on specified URL
type Server struct {
	// Addr is a TCP address to listen on (can be just port ":8081", or "localhost:8081")
	Addr string

	// StreamURL is a sub-URL for stream, "/stream" means server will stream on "http://localhost:8081/stream"
	StreamURL string

	// Source is a channel of images represented as []byte
	Source chan []byte
}

// FullStreamURL returns full stream URL for links
func (s *Server) FullStreamURL() string {
	return s.Addr + s.StreamURL
}

// index page handler
func (s *Server) indexHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(
		res,
		`<center>
			<style>
				img{border:1px solid red;padding:3px}
				button{margin:10px}
				#box{display:flex;flex-wrap:wrap;justify-content: center}
			</style>
			<h1>MJPEG stream</h1>
			Link: <a href="%s">%s</a><br>
			<button onclick="
				var el = document.createElement('img');
				el.src = '%s?' + Math.random();
				el.setAttribute('onclick', 'this.remove()');
				document.getElementById('box').appendChild(el);
			">+</button><br>
			<div id="box">
				<img onclick="this.remove()" src="%s" style="height:70vh">
			</div>
		</center>`,
		s.FullStreamURL(),
		s.FullStreamURL(),
		s.StreamURL,
		s.StreamURL,
	)
}

// ListenAndServe starts the HTTP server that:
// - has index page with stream demo
// - streams MJPEG video on specified StreamURL
func (s *Server) ListenAndServe() error {
	// index page
	if s.StreamURL != "/" {
		http.HandleFunc("/", s.indexHandler)
	}

	// create MJPEG stream handler from source channel
	stream := NewHandler(s.Source)
	http.HandleFunc(s.StreamURL, stream.HTTPHandler)

	server := &http.Server{
		Addr: s.Addr,
	}

	fmt.Printf("[MJPEG Server] streaming on %s\n", s.FullStreamURL())

	return server.ListenAndServe()
}
