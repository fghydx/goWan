package TCPSvrNormal

import (
	"io"
	"net"
	"testing"
)

type ReadType uint8

const (
	_ ReadType = iota
	ReadHead
	ReadContent
)

type testTcpObj struct {
	readtype ReadType
}

func (t *testTcpObj) OnConnect(conn net.Conn) bool {
	println("连接上来了")
	return true
}
func (t *testTcpObj) OnDisconnect(conn net.Conn) {
	println("断开连接")
}
func (t *testTcpObj) OnError(conn net.Conn, err error) {
	println("连接错误")
}
func (t *testTcpObj) OnReadData(conn net.Conn) (closed bool, err error) {
	println("接收数据")
	b := make([]byte, 1)
	switch t.readtype {
	case ReadHead:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = ReadContent //接收完头，接收内容
		}
	case ReadContent:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = ReadHead //接收完内容，接收头
		}
	default:
		t.readtype = ReadHead
	}
	return false, err
}

func TestNewNetFrame(t *testing.T) {
	tcpframe := NewNetFrame("127.0.0.1", 9997, &testTcpObj{})
	err := tcpframe.Start()
	if err != nil {
		t.Log(err)
	}
}
