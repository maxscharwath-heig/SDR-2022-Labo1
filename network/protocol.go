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
	NeedsAuth   bool
	HandlerFunc func(request Request) any
}

type connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

func (c connection) isClosed() bool {
	_, err := c.reader.Peek(1)
	return err != nil
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
	log("recv", data, err)
	return data, err
}

func (c connection) getJson(data any) error {
	jsonString, err1 := c.getLine()
	if err1 != nil {
		return err1
	}
	return json.Unmarshal([]byte(jsonString), data)
}

func (c connection) getHeader() HeaderResponse {
	var header HeaderResponse
	_ = c.getJson(&header)
	return header
}

type ServerProtocol struct {
	AuthFunc  AuthFunc
	Endpoints map[string]Endpoint
}

func (p ServerProtocol) Process(c net.Conn) {
	log("new connection", c.RemoteAddr())
	defer func() {
		log("close connection", c.RemoteAddr())
		_ = c.Close()
	}()

	conn := connection{
		conn:   c,
		reader: bufio.NewReader(c),
	}
	for {
		if conn.isClosed() {
			break
		}

		request := Request{}
		request.Path, _ = conn.getLine()

		endpoint, ok := p.Endpoints[request.Path]
		if ok {
			request.Header.Valid = true
			request.Header.NeedsAuth = endpoint.NeedsAuth
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
	conn := connection{
		conn:   p.Conn,
		reader: bufio.NewReader(p.Conn),
	}
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

func log(prefix string, data ...any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	color := "\033[33m"
	reset := "\033[0m"
	fmt.Println(color, fmt.Sprintf("[%s] (%s):", date, prefix), reset, data)
}
