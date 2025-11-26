package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
		headers, trailers := headers.NewHeaders(), headers.NewHeaders()
		headers.Set("Transfer-Encoding", "chunked")
		trailers.Set("X-Content-SHA256", "")
		trailers.Set("X-Content-Length", "")

		w.WriteHeaders(headers, []string{"Content-Length"}, trailers)

		buffer := make([]byte, 64)
		isEof := false
		fullBody := bytes.NewBuffer([]byte{})
		for {
			n, err := body.Read(buffer)

			if err != nil {
				if err != io.EOF {
					log.Println(err)
					return
				} else {
					isEof = true
				}
			}
			fullBody.Write(buffer[:n])

			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				log.Println(err)
				return
			}

			if isEof {
				_, err = w.WriteChunkedBodyDone()
				if err != nil {
					log.Println(err)
				}

				hash := sha256.Sum256(fullBody.Bytes())
				encodedHash := hex.EncodeToString(hash[:])

				trailers.Replace("X-Content-SHA256", encodedHash)
				trailers.Replace("X-Content-Length", strconv.Itoa(len([]byte(encodedHash))))

				err := w.WriteTrailers(trailers)
				if err != nil {
					log.Println(err)
				}
				return
			}
		}

	} else if url == "/video" {
		w.WriteStatusLine(200)
		headers := headers.NewHeaders()
		headers.Set("Content-Type", "video/mp4")
		file, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Println("Error reading file:", err)
			return
		}
		headers.Set("Content-Length", strconv.Itoa(len(file)))
		w.WriteHeaders(headers, nil, nil)
		w.WriteBody(file)
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
		w.WriteHeaders(headers, nil, nil)
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
