package main

import (
	"encoding/binary"
	tcp "github.com/fghydx/goWan/tcp"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewtcpServe(t *testing.T) {
	go func() {
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()
	go func() {
		svr := NewtcpServe("127.0.0.1:8889")
		svr.RegisterHandler(1, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(1, data)
		})
		svr.RegisterHandler(2, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(2, append([]byte("我是2"), data...))
		})
		svr.RegisterHandler(3, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(3, append([]byte("我是3"), data...))
		})
		svr.RegisterHandler(0, func(conn *tcp.Connector, packetEnd bool, data []byte) {
			conn.SendData(0, data)
		})
		if err := svr.Start(); err != nil {
			println(err)
		}
	}()
	gocount := 5000
	wait := sync.WaitGroup{}
	wait.Add(gocount)
	a := int32(0)
	time.Sleep(2 * time.Second)
	for i := 0; i < gocount; i++ {
		go func(index int) {
			con, err := net.Dial("tcp4", "127.0.0.1:8889")
			defer func() {
				err := recover()
				if err != nil {
					println("错误", err.(error).Error())
				}
				wait.Done()
			}()
			if err != nil {
				println("连接错误", err.Error())
				return
			}
			defer con.Close()
			var id, len uint32
			pack := Packet{}
			for i := 0; i < 100; i++ {
				b := pack.PackData(i%4, []byte("测试"+strconv.Itoa(i)))
				_, err := con.Write(b)
				if err != nil {
					println("发送错误", err.Error())
					continue
				}
				err = binary.Read(con, binary.LittleEndian, &id)
				if err != nil {
					println("读取错误", err.Error())
					continue
				}
				err = binary.Read(con, binary.LittleEndian, &len)
				if err != nil {
					println("读取错误1", err.Error())
					continue
				}
				c := make([]byte, len)
				_, err = io.ReadFull(con, c)
				if err != nil {
					println("读取错误2", err.Error())
					continue
				}
				atomic.AddInt32(&a, 1)
				//println("我是客户端", index, "号，请求ID:", id, string(c), a)
			}
		}(i)
	}
	wait.Wait()
	println(a)
	select {}
}
