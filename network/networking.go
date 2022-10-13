package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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
	reader := bufio.NewReader(conn)
	for {
		path, ePath := reader.ReadString('\n')
		body, eBody := reader.ReadString('\n')
		if ePath != nil || eBody != nil {
			break
		}
		entryMessages <- Message{strings.Trim(path, "\n"), strings.Trim(body, "\n")}
	}
}
