package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const headlen = 8

func PackData(dataEx any, data []byte) []byte {
	datalen := uint32(len(data))
	var result []byte
	buff := bytes.NewBuffer(result)
	id := uint32(dataEx.(int))
	_ = binary.Write(buff, binary.LittleEndian, id)
	_ = binary.Write(buff, binary.LittleEndian, datalen)
	_ = binary.Write(buff, binary.LittleEndian, data)
	return buff.Bytes()
}

func main() {
	i := 0
	k := 0
	for {
		i++
		k++
		if k > 500 {
			time.Sleep(time.Second)
			continue
		}
		println("创建go程", k, i)
		go func(i int) {
			conns, err := net.Dial("tcp4", "127.0.0.1:8889")
			if err != nil {
				println("创建连接失败", err.Error())
				return
			}
			j := rand.Int()
			for {
				j++
				var id, len uint32
				b := PackData(i%4, []byte("测试"+strconv.Itoa(i)))
				conns.Write(b)
				err = binary.Read(conns, binary.LittleEndian, &id)
				if err != nil {
					println("读取数据失败", err.Error())
					return
				}
				err = binary.Read(conns, binary.LittleEndian, &len)
				if err != nil {
					return
				}
				c := make([]byte, len)
				_, err = io.ReadFull(conns, c)
				if err != nil {
					println("读取数据失败", err.Error())
					return
				}
				println(id, string(c))
				if j%5 == 0 {
					conns.Close()
					k--
					println("关闭")
					break
				}
			}
		}(i)
	}
	println("退出程序")
	select {}
}
