// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sdr/labo1/src/utils"
	"strings"
)

// Connection
// is used to handle the Connection and create a wrapper around it.
type Connection struct {
	net.Conn
	reader *bufio.Reader
}

func CreateConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (c Connection) IsClosed() bool {
	_, err := c.Read(make([]byte, 0))
	return err != nil
}

func (c Connection) SendData(data string) error {
	utils.LogInfo(false, fmt.Sprintf("📤SEND TO  %s", c.RemoteAddr().String()), data)
	_, err := fmt.Fprintln(c, data)
	return err
}

func (c Connection) SendJSON(data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.SendData(string(bytes))
}

func (c Connection) SendResponse(endpointId string, success bool, data any) error {
	if err := c.SendData(endpointId); err != nil {
		return err
	}
	return c.SendJSON(CreateResponse(success, data))
}

func (c Connection) GetLine() (string, error) {
	data, err := c.reader.ReadString('\n')
	data = strings.TrimSpace(data)
	utils.LogInfo(false, fmt.Sprintf("📥GOT FROM %s", c.RemoteAddr().String()), data)
	return data, err
}

func (c Connection) GetJson(data any) error {
	jsonString, err := c.GetLine()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonString), data)
}

// Request defines the format of client to server communication
type Request[Header any] struct {
	Conn       net.Conn
	EndpointId string
	Header     Header
	Data       string
}

func (r Request[T]) GetJson(data any) {
	_ = json.Unmarshal([]byte(r.Data), data)
}

// Response defines the format of server to client communication
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateResponse creates a response to be sent to a client
func CreateResponse(success bool, data any) (response Response[any]) {
	response.Success = success
	if success {
		response.Data = data
	} else {
		response.Error = data.(string)
	}
	return
}

// ParseResponse parse a response to a struct
func ParseResponse[T any](data string) (res T, err error) {
	var result Response[T]
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return
	}
	if !result.Success {
		err = fmt.Errorf(result.Error)
		return
	}
	return result.Data, nil
}

func GetJson[T any](conn Connection) (data T, err error) {
	err = conn.GetJson(&data)
	return
}

func GetResponse[T any](conn Connection, endpointId string) (res T, err error) {
	if endpoint, e := conn.GetLine(); e != nil || endpoint != endpointId {
		return res, fmt.Errorf("invalid endpoint")
	}
	var data string
	if data, err = conn.GetLine(); err != nil {
		return
	}
	return ParseResponse[T](data)
}
