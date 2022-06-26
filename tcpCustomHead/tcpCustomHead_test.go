package tcpCustomHead

import (
	"encoding/binary"
	tcp "github.com/fghydx/goWan/tcp"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestNewtcpServe(t *testing.T) {
	go func() {
		RegisterHandler(1, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(1, data)
		})
		RegisterHandler(2, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(2, append([]byte("我是2"), data...))
			conn.Stop()
		})
		RegisterHandler(3, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(3, append([]byte("我是3"), data...))
		})
		RegisterHandler(0, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(0, data)
		})

		svr := NewtcpServe("127.0.0.1:8889")
		if err := svr.Start(); err != nil {
			println(err)
		}
	}()

	time.Sleep(2 * time.Second)
	for i := 0; i < 10; i++ {
		go func(index int) {
			con, _ := net.Dial("tcp4", "127.0.0.1:8889")
			var id, len uint32
			pack := Packet{}
			for i := 0; i < 1000; i++ {
				b := pack.PackData(i%4, []byte("测试"+strconv.Itoa(i)))
				con.Write(b)
				binary.Read(con, binary.LittleEndian, &id)
				binary.Read(con, binary.LittleEndian, &len)
				c := make([]byte, len)
				io.ReadFull(con, c)
				println("我是客户端", index, "号，请求ID:", id, string(c))
			}
			select {}
		}(i)
	}
	select {}
}
