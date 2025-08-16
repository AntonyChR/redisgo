package network

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// server roles
const (
	MASTER = "master"
	SLAVE  = "slave"
)

type Data struct {
	Id   string
	Data []byte
}

func CreateNewServer(port string, role string, masterAddr string) (*TcpServer, error) {
	if role != SLAVE && role != MASTER {
		errorMsg := fmt.Sprintf("invalid role option, expected 0(master) or 1(slave) got: %s", role)
		return nil, errors.New(errorMsg)
	}
	if role == SLAVE {
		//TODO: validate masterAddr
		//return slave server
	}

	return &TcpServer{
		RegisterNewSlaveChan: make(chan string),
		conn:                 make(map[string]net.Conn),
		context:              context.Background(),
		port:                 port,
	}, nil

}

type TcpServer struct {
	port                 string
	RegisterNewSlaveChan chan string
	conn                 map[string]net.Conn
	context              context.Context
}

func (s *TcpServer) Start(handleConn func(net.Conn)) error {
	listener, err := net.Listen("tcp", "0.0.0.0:"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	errChan := make(chan error, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			}
			go handleConn(conn)
		}
	}()

	select {
	case <-s.context.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
