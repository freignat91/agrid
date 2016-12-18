package gnode

import (
	"fmt"
	"log"
	"net"
)

type udpServer struct {
	gnode         *GNode
	server        *net.UDPConn
	leaderManager *gnodeLeader
}

func (s *udpServer) start(gnodeLeader *gnodeLeader) error {
	s.leaderManager = gnodeLeader
	serverAddr, err := net.ResolveUDPAddr("udp", ":"+config.udpPort)
	if err != nil {
		return err
	}
	serv, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		return err
	}
	s.server = serv
	go func() {
		for {
			buf := make([]byte, 1024)
			n, addr, err := s.server.ReadFromUDP(buf)
			if err != nil {
				log.Printf("UDP receive error: %v\n", err)
			}
			mes := string(buf[0:n])
			//log.Printf("UDP received from %s mes=%s\n", addr, mes)
			s.received(mes, addr)
		}
	}()
	return nil
}

func (s *udpServer) close() {
	s.server.Close()
}

func (s *udpServer) send(addr *net.UDPAddr, mes string) error {
	//logf.printf("udp send to %s mes=%s\n", addr.String(), mes)
	addr.Port = 3010
	if _, err := s.server.WriteTo([]byte(mes), addr); err != nil {
		fmt.Errorf("UDP send error: %s to %s", mes, addr.String())
	}
	//logf.printf("send %s to %s\n", mes, addr.String())
	return nil
}

func (s *udpServer) sendFromIP(addr *net.IP, mes string) error {
	//logf.printf("udp SendFromIP to %s mes=%s\n", addr.String(), mes)
	uAddr := &net.UDPAddr{
		IP:   *addr,
		Port: 3010,
	}
	conn, err := net.DialUDP("udp4", nil, uAddr)
	if err != nil {
		return fmt.Errorf("UDP broadcast dial error: %v\n", err)
	}
	defer conn.Close()
	buf := []byte(mes)
	_, errConn := conn.Write(buf)
	if errConn != nil {
		logf.error("Error writing UDP message %s: %v", mes, errConn)
	}
	return nil
}

func (s *udpServer) received(mes string, addr *net.UDPAddr) {
	if mes == "UpdateGrid" {
		s.leaderManager.updateGrid(true, false)
	}
}
