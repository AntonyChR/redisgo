package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	utils "redisgo/utils"
)

// server roles
const (
	MASTER = "master" 
	SLAVE = "slave"
)

type Data struct {
	Id string
	Data []byte
}

func CreateNewServer(port string, role string, masterAddr string) (*TcpServer, error){
	if role != SLAVE && role != MASTER {
		errorMsg := fmt.Sprintf("invalid role option, expected 0(master) or 1(slave) got: %d", role)
		return nil, errors.New(errorMsg)
	}
	if role == SLAVE {
		//TODO: validate masterAddr
		//return slave server
	}

	return &TcpServer{
		Read: make(chan Data),
		Write: make(chan Data),
		RegisterNewSlaveChan: make(chan string),
		conn: make(map[string]net.Conn),
		context: context.Background(),
	}, nil

}

type TcpServer struct{
	Read chan Data 
	Write chan Data 
	port string
	RegisterNewSlaveChan chan string
	conn  map[string] net.Conn
	context context.Context
}

func (s *TcpServer) Listen() error{
	listener, err := net.Listen("tcp", "0.0.0.0:" + s.port)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s* TcpServer) handleConnection(conn net.Conn){
	buff := make([]byte, 1024)

	n, err := conn.Read(buff)

	if err == io.EOF {
		log.Println("EOF, connection closed", err)
		return
	}

	if err != nil {
		log.Println("error reading data, ", err)
		return
	}

	if n == 0 {
		log.Println("data length: 0")
		return
	}

	conId := utils.GenerateUUID()

	s.conn[conId] = conn
	
	s.Read <- Data{conId,buff[:n]}

}

func (s *TcpServer) send(){
	for{
		select{
		case data := <- s.Read:

			if conn, ok := s.conn[data.Id]; ok {
				n, err := conn.Write(data.Data)
				if err != nil {
					log.Println("error writting in connection, ", err)
					continue
				}

				if n == 0 {
					log.Println("data writted in connection was 0")
					continue
				}
				conn.Close()
				delete(s.conn, data.Id)
			}

		case <- s.context.Done():
			return
		}
	}
}

func (s *TcpServer) recv(){
	for{
		select{
		case data := <- s.Write:
			log.Println(string(data.Data))
			// send to parser/command extractor -> execute
		case <- s.context.Done():
			return
		}
	}
}

func (s *TcpServer) Start(){
	go s.send()
	go s.recv()
}
