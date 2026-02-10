package ui

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"radio-tui/internal/ipc"
)

type ipcServer struct {
	endpoint ipc.Endpoint
	listener net.Listener
	messages chan ipcMsg
	done     chan struct{}
}

type ipcMsg struct {
	cmd   string
	reply chan ipcReply
}

type ipcReply struct {
	ok   bool
	data string
	err  string
}

type ipcReadyMsg struct {
	server *ipcServer
	err    error
}

type ipcClosedMsg struct{}

func newIPCServer() (*ipcServer, error) {
	listener, endpoint, err := ipc.Listen()
	if err != nil {
		return nil, err
	}

	server := &ipcServer{
		endpoint: endpoint,
		listener: listener,
		messages: make(chan ipcMsg, 8),
		done:     make(chan struct{}),
	}
	go server.acceptLoop()
	return server, nil
}

func (s *ipcServer) Close() {
	select {
	case <-s.done:
		return
	default:
		close(s.done)
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	_ = ipc.Cleanup(s.endpoint)
}

func (s *ipcServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *ipcServer) handleConn(conn net.Conn) {
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewScanner(conn)
	if !reader.Scan() {
		s.writeReply(conn, ipcReply{ok: false, err: "empty command"})
		return
	}
	cmd := strings.TrimSpace(reader.Text())
	if cmd == "" {
		s.writeReply(conn, ipcReply{ok: false, err: "empty command"})
		return
	}

	replyChan := make(chan ipcReply, 1)
	msg := ipcMsg{cmd: cmd, reply: replyChan}

	select {
	case <-s.done:
		s.writeReply(conn, ipcReply{ok: false, err: "server shutting down"})
		return
	case s.messages <- msg:
	default:
		s.writeReply(conn, ipcReply{ok: false, err: "busy"})
		return
	}

	select {
	case reply := <-replyChan:
		s.writeReply(conn, reply)
	case <-time.After(2 * time.Second):
		s.writeReply(conn, ipcReply{ok: false, err: "timeout"})
	}
}

func (s *ipcServer) writeReply(conn net.Conn, reply ipcReply) {
	if reply.ok {
		if strings.TrimSpace(reply.data) != "" {
			_, _ = fmt.Fprintln(conn, reply.data)
			return
		}
		_, _ = fmt.Fprintln(conn, "OK")
		return
	}

	message := strings.TrimSpace(reply.err)
	if message == "" {
		message = "error"
	}
	_, _ = fmt.Fprintln(conn, "ERR "+message)
}

func parseIPCCommand(cmd string) (string, error) {
	cmd = strings.TrimSpace(strings.ToUpper(cmd))
	if cmd == "" {
		return "", errors.New("empty command")
	}
	return cmd, nil
}
