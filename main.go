package main

import tcp "github.com/fghydx/goWan/tcp"
import . "github.com/fghydx/goWan/tcpCustomHead"

func main() {
	RegisterHandler(1, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(1, data)
	})
	RegisterHandler(2, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(2, append([]byte("我是2"), data...))
		//conn.Stop()
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
}
