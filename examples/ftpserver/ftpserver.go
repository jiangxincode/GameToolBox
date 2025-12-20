package main

import (
	"fmt"
	"net"
	"os"
)

const (
	DIR = "DIR"
	CD  = "CD"
	PWD = "PWD"
)

func main() {

	server := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error ", err.Error())
		os.Exit(1)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	var buf [512]byte

	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			conn.Close()
			return
		}

		s := string(buf[0:n])
		switch {
		case s[0:2] == CD:
			chdir(conn, s[3:])
		case s[0:3] == DIR:
			listdir(conn)
		case s[0:3] == PWD:
			pwd(conn)
		}
	}
}

func chdir(conn net.Conn, dir string) {
	if os.Chdir(dir) == nil {
		conn.Write([]byte("OK"))
	} else {
		conn.Write([]byte("ERROR"))
	}
}

func listdir(conn net.Conn) {
	defer conn.Write([]byte("\r\n"))

	dir, err := os.Open(".")
	if err != nil {
		conn.Write([]byte("\r\n"))
		return
	}

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return
	}

	for _, mm := range names {
		conn.Write([]byte(mm + "\r\n"))
	}
}

func pwd(conn net.Conn) {
	s, err := os.Getwd()
	if err != nil {
		conn.Write([]byte(""))
		return
	}
	conn.Write([]byte(s))
}
