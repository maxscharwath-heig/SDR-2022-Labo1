package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sdr/labo1/types"
	"strings"
	"time"
)

type HeaderResponse struct {
	Valid     bool
	NeedsAuth bool
}

type AuthResponse struct {
	Success bool
	Auth    any
}

type AuthFunc func(credentials types.Credentials) (bool, any)

type Request struct {
	Path   string
	Header HeaderResponse
	Auth   any
	Data   string
}

func (r Request) GetJson(data any) {
	_ = json.Unmarshal([]byte(r.Data), data)
}

type Endpoint struct {
	Path        string
	NeedsAuth   bool
	HandlerFunc func(request Request) any
}

type connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

func (c connection) sendData(data string) {
	log("send", data)
	fmt.Fprintln(c.conn, data)

}

func (c connection) sendJSON(data any) {
	bytes, _ := json.Marshal(data)
	c.sendData(string(bytes))
}

func (c connection) getLine() (string, error) {
	data, err := c.reader.ReadString('\n')
	data = strings.Trim(data, "\n")
	log("recv", data)
	return data, err
}

func (c connection) getJson(data any) {
	jsonString, _ := c.getLine()
	_ = json.Unmarshal([]byte(jsonString), data)
}

func (c connection) getHeader() HeaderResponse {
	var header HeaderResponse
	c.getJson(&header)
	return header
}

type ServerProtocol struct {
	AuthFunc  AuthFunc
	Endpoints []Endpoint
}

func (p ServerProtocol) Process(c net.Conn) {
	conn := connection{c, bufio.NewReader(c)}
	for {
		request := Request{}
		request.Path, _ = conn.getLine()

		var endpoint *Endpoint
		for _, e := range p.Endpoints {
			if e.Path == request.Path {
				request.Header.Valid = true
				request.Header.NeedsAuth = e.NeedsAuth
				endpoint = &e
			}
		}
		conn.sendJSON(request.Header)

		if !request.Header.Valid {
			continue
		}
		if request.Header.NeedsAuth {
			var credentials types.Credentials
			conn.getJson(&credentials)
			isValid, auth := p.AuthFunc(credentials)
			conn.sendJSON(AuthResponse{isValid, auth})
			request.Auth = auth
			if !isValid {
				continue
			}
		}
		request.Data, _ = conn.getLine()

		response := endpoint.HandlerFunc(request)
		conn.sendJSON(response)
	}
}

type ClientProtocol struct {
	Conn     net.Conn
	AuthFunc func() types.Credentials
}

func (p ClientProtocol) SendRequest(path string, data func(auth any) any) (string, error) {
	conn := connection{p.Conn, bufio.NewReader(p.Conn)}
	conn.sendData(path)
	header := conn.getHeader()
	if !header.Valid {
		return "", fmt.Errorf("invalid path")
	}
	authResponse := AuthResponse{}
	if header.NeedsAuth {
		conn.sendJSON(p.AuthFunc())
		conn.getJson(&authResponse)
		if !authResponse.Success {
			return "", fmt.Errorf("invalid credentials")
		}
	}
	conn.sendJSON(data(authResponse.Auth))
	return conn.getLine()
}

func log(prefix string, data any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	color := "\033[33m"
	reset := "\033[0m"
	fmt.Println(color, fmt.Sprintf("[%s] (%s):", date, prefix), reset, data)
}
