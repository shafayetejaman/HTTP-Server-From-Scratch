package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"tcpTohttp/internal/request"
	"tcpTohttp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
type Server struct {
	listener net.Listener
	runing   atomic.Bool
	Handler  Handler
}
type Handler func(w *response.Writer, req *request.Request)

func WriteError(h HandlerError, w io.Writer) {
	response.WriteStatusLine(w, h.StatusCode)
	headers := response.GetDefaultHeaders(len(h.Message))
	response.WriteHeaders(w, headers)
	w.Write([]byte(h.Message))
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server := &Server{listener: listener, Handler: handler}

	server.runing.Store(true)
	go server.listen()

	return server, nil
}

func (s *Server) isClosed() bool {
	return !s.runing.Load()
}

func (s *Server) Close() error {
	s.runing.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.isClosed() {
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	request, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{StatusCode: 500, Message: err.Error()}
		log.Println(err)
		WriteError(*handlerError, conn)
		return
	}
	writer := &response.Writer{Conn: conn, Status: response.StatusWriteStatusLine}
	s.Handler(writer, request)
}

