package main

import (
	"fmt"
	"log"
	"net"
	"tcpTohttp/internal/request"
)

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	channel := make(chan string)

// 	go func() {
// 		defer f.Close()
// 		defer close(channel)

// 		var s string

// 		for {
// 			newBytes := make([]byte, 8)
// 			_, err := f.Read(newBytes)
// 			if err != nil {
// 				break
// 			}
// 			s += string(newBytes)
// 			if i := strings.Index(s, "\n"); i != -1 {

// 				channel <- s[:i]
// 				s = s[i+1:]
// 			}
// 			// time.Sleep(200 * time.Millisecond)
// 		}

// 		if len(s) != 0 {
// 			channel <- s
// 		}
// 	}()

// 	return channel

// }
func main() {
	listener, err := net.Listen("tcp", ":42069")
	defer listener.Close()

	if err != nil {
		log.Fatal(err)
	}

	for {
		cnn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("connection has been accepted!")
		req, err := request.RequestFromReader(cnn)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Request line:\n - Method: %v\n - Target: %v\n - Version: %v\n Headers:\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion)
		for key, val := range req.Headers.Headers {
			fmt.Printf("- %v: %v\n", key, val)
		}
		fmt.Printf("Body:\n%v\n", string(req.Body))
		cnn.Close()
	}

}
