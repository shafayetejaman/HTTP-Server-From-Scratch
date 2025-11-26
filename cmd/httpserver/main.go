package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	server "tcpTohttp/internal"
	"tcpTohttp/internal/headers"
	"tcpTohttp/internal/request"
	"tcpTohttp/internal/response"
	"time"
)

const PORT = 42069

func getResponseBody(requestUrl string) (io.ReadCloser, error) {
	url := strings.TrimPrefix(requestUrl, "/httpbin/")
	res, err := http.Get("https://httpbin.org/" + url)

	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func handlerfunc(w *response.Writer, req *request.Request) {
	url := req.RequestLine.RequestTarget

	if strings.HasPrefix(url, "/httpbin") {
		body, err := getResponseBody(req.RequestLine.RequestTarget)
		if err != nil {
			log.Println(err)
			return
		}

		w.WriteStatusLine(200)
		headers := headers.NewHeaders()
		headers.Set("Transfer-Encoding", "chunked")
		w.WriteHeaders(headers, []string{"Content-Length"})

		buffer := make([]byte, 32)
		read := 0
		isEof := false
		for {
			if !isEof {
				n, err := body.Read(buffer[read:])

				if err != nil {
					if err != io.EOF {

						log.Println(err)
						return
					} else {
						isEof = true
					}
				}
				read += n
			}

			n, err := w.WriteChunkedBody(buffer[:read])
			if err != nil {
				log.Println(err)
				return
			}

			if n == 0 {
				_, err = w.WriteChunkedBodyDone()
				if err != nil {
					log.Println(err)
				}
				return
			}
			copy(buffer, buffer[n:read])
			read -= n
			time.Sleep(1 * time.Second)
		}
	} else {

		data := []byte(`
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
    			<p>Your request was an absolute banger.</p>
			</body>
		</html>`)
		headers := headers.NewHeaders()
		headers.Set("Content-Length", strconv.Itoa(len(data)))
		headers.Set("Content-Type", "text/html")
		w.WriteStatusLine(200)
		w.WriteHeaders(headers, nil)
		w.WriteBody(data)
	}

}
func main() {
	server, err := server.Serve(PORT, handlerfunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", PORT)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
