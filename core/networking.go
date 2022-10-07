package core

import (
	"bufio"
	"fmt"
	"net"
)

func SendRequest(conn net.Conn, path string, body string) {
	fmt.Fprintln(conn, path)
	fmt.Fprintln(conn, body)
}

type Message struct {
	Path string
	Body string
}

func ReceiveData(conn net.Conn, entryMessages chan Message) {
	for {
		path, _ := bufio.NewReader(conn).ReadString('\n')
		body, _ := bufio.NewReader(conn).ReadString('\n')
		entryMessages <- Message{path, body}
	}
}
