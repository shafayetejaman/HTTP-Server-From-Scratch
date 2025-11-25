package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	channel := make(chan string)

	go func() {
		defer f.Close()
		defer close(channel)

		var s string

		for {
			newBytes := make([]byte, 8)
			_, err := f.Read(newBytes)
			if err != nil {
				break
			}
			s += string(newBytes)
			if i := strings.Index(s, "\n"); i != -1 {

				channel <- s[:i]
				s = s[i+1:]
			}
			// time.Sleep(200 * time.Millisecond)
		}

		if len(s) != 0 {
			channel <- s
		}
	}()

	return channel

}
func main() {
	address := "localhost:42069"
	udpadd, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, udpadd)
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Write([]byte(msg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("message sent : ", msg)
	}

}
