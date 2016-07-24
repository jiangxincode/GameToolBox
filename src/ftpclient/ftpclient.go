package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	uiDir  = "dir"
	uiCd   = "cd"
	uiPwd  = "pwd"
	uiQuit = "quit"
)

const (
	DIR = "DIR"
	PWD = "PWD"
	CD  = "CD"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s host:port", os.Args[0])
		os.Exit(1)
	}

	server := os.Args[1]
	conn, err := net.Dial("tcp", server)
	checkError(err)

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, " \t\r\n")
		if err != nil {
			break
		}

		strs := strings.SplitN(line, " ", 2)
		switch strs[0] {
		case uiDir:
			listdir(conn)
		case uiPwd:
			pwd(conn)
		case uiCd:
			if len(strs) != 2 {
				fmt.Println("cd <dir>")
				continue
			}
			fmt.Println("CD \"", strs[1], "\"")
			cd(conn, strs[1])
		case uiQuit:
			conn.Close()
			os.Exit(0)
		default:
			fmt.Println("unknow command")
		}
	}
}

func listdir(conn net.Conn) {
	conn.Write([]byte(DIR + " "))
	var buf [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		result.Write(buf[0:n])
		length := result.Len()
		content := result.Bytes()
		if string(content[length-4:]) == "\r\n\r\n" {
			fmt.Println("dir content is :", "\n", string(content[0:length-4]))
			return
		}

	}

}
func pwd(conn net.Conn) {
	conn.Write([]byte(PWD))
	var res [512]byte
	n, _ := conn.Read(res[0:])
	s := string(res[0:n])
	fmt.Println("current dir is \"", s, "\"")

}
func cd(conn net.Conn, dir string) {
	conn.Write([]byte(CD + " " + dir))
	var res [512]byte
	n, _ := conn.Read(res[0:])
	s := string(res[0:n])
	if s != "OK" {
		fmt.Println("faild to change dir")
		return
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error ", err.Error())
		os.Exit(1)
	}
}
