package network

import (
	"bufio"
	"fmt"
	"net"
)

type Message struct {
	Path string
	Body string
}

func SendRequest[T any](conn net.Conn, path string, body Request[T]) {
	SendData(conn, path, body.ToJson())
}

func SendData(conn net.Conn, path string, body string) {
	fmt.Fprintln(conn, path)
	fmt.Fprintln(conn, body)
}

func HandleReceiveData(conn net.Conn, entryMessages chan Message) {
	for {
		path, _ := bufio.NewReader(conn).ReadString('\n')
		body, _ := bufio.NewReader(conn).ReadString('\n')
		entryMessages <- Message{path[:len(path)-1], body[:len(body)-1]}
	}
}
