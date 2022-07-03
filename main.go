package main

import (
	tcp "github.com/fghydx/goWan/tcp"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
)
import . "github.com/fghydx/goWan/tcpCustomHead"

//获取当前路径
func GetCurrentPath() string {
	return filepath.Dir(os.Args[0]) + string(filepath.Separator)
}
func RedirectStdOut(fileName string) {
	f, _ := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND,
		0755)
	os.Stdout = f
}

func main() {
	RegisterHandler(1, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(1, data)
	})
	RegisterHandler(2, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(2, append([]byte("2"), data...))
		//conn.Stop()
	})
	RegisterHandler(3, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(3, append([]byte("3"), data...))
	})
	RegisterHandler(0, func(conn *tcp.Connector, packetEnd bool, data []byte) {
		conn.SendData(0, data)
	})

	RedirectStdOut(GetCurrentPath() + "log.txt")

	go func() {
		http.ListenAndServe("127.0.0.1:8888", nil)
	}()

	svr := NewtcpServe("127.0.0.1:8889")
	if err := svr.Start(); err != nil {
		println(err)
	}
}
