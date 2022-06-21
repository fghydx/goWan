package TCPSvrNormal

import (
	"github.com/fghydx/goWan"
	"io"
	"net"
	"testing"
)

type testTcpObj struct {
	readtype main.ReadType
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
	case main.ReadHead:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = main.ReadContent //接收完头，接收内容
		}
	case main.ReadContent:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = main.ReadHead //接收完内容，接收头
		}
	default:
		t.readtype = main.ReadHead
	}
	return false, err
}

func TestNewNetFrame(t *testing.T) {
	tcpframe := NewNetFrame("127.0.0.1", 9997, &testTcpObj{})
	tcpframe.Start()
}
