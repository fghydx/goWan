package tcp

import (
	"fmt"
	"io"
	"net"
	"testing"
)

type TcpReader struct {
	i int
}

func (t *TcpReader) WriteData(connector *Connector, dataEx any, data []byte) {
	connector.SendDataChan <- data
}

func (t *TcpReader) ReadData(connector *Connector) (closed bool, err error) {
	closed = false
	b := make([]byte, 1)
	_, err = io.ReadFull(connector.Conn, b)
	println(fmt.Sprintf("源：%s,值：%v, i:%d, ID:%d", connector.Conn.RemoteAddr().String(), b, t.i, connector.ConID))
	t.i++
	connector.SendData(nil, []byte(fmt.Sprintf("源：%s,值：%v, i:%d, ID:%d", connector.Conn.RemoteAddr().String(), b, t.i, connector.ConID)))
	return
}

func (t *TcpReader) Init() {
	t.i = 0
}

func (t *TcpReader) NewReaderWriter() ITcpReaderWriter {
	return &TcpReader{
		i: 0,
	}
}

func TestNewTcpServe(t *testing.T) {
	tcp := NewServe("127.0.0.1:8889", &TcpReader{})
	tcp.FOnConnect = func(conn net.Conn) bool {
		println(fmt.Sprintf("连接上来了：%s", conn.RemoteAddr().String()))
		return true
	}
	tcp.FOnDisConnect = func(conn net.Conn) {
		println(fmt.Sprintf("断开了连接：%s", conn.RemoteAddr().String()))
	}
	tcp.FOnError = func(conn net.Conn, err error) {
		println(fmt.Sprintf("%s发生了错误", conn.RemoteAddr().String()), err)
	}
	err := tcp.Start()
	if err != nil {
		println(err)
	}
}
