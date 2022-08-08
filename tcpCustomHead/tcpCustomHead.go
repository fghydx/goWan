package tcpCustomHead

/*
	实现了协议为ID+内容长度+内容的传输
*/

import (
	"bytes"
	"encoding/binary"
	tcp "github.com/fghydx/goWan/tcp"
	"github.com/fghydx/goWan/tcpReadPacket"
	"io"
	"sync/atomic"
)

type HandFunc func(conn *tcp.Connector, packetEnd bool, data []byte)

var HandlerMap = make(map[uint32]IHandler)

func RegisterHandler(ID uint32, handler HandFunc) {
	HandlerMap[ID] = &handler
}

type IHandler interface {
	Handler(conn *tcp.Connector, packetEnd bool, data []byte)
}

type Header struct {
	iD      uint32
	dataLen uint32
}

const bufflen = uint32(4096)

type Packet struct {
	header        Header
	readBuff      [bufflen]byte
	readedDataLen uint32
}

func (p *Packet) PackData(dataEx any, data []byte) []byte {
	datalen := uint32(len(data))
	buff := bytes.NewBuffer([]byte{})
	id := uint32(dataEx.(int))
	_ = binary.Write(buff, binary.LittleEndian, id)
	_ = binary.Write(buff, binary.LittleEndian, datalen)
	_ = binary.Write(buff, binary.LittleEndian, data)
	return buff.Bytes()
}

func (f HandFunc) Handler(conn *tcp.Connector, packetEnd bool, data []byte) {
	f(conn, packetEnd, data)
}

func (p *Packet) ReadHead(connector *tcp.Connector) (ok bool, closed bool, err error) {
	ok = true
	closed = false
	err = binary.Read(connector.Conn, binary.LittleEndian, &p.header.iD)
	err = binary.Read(connector.Conn, binary.LittleEndian, &p.header.dataLen)
	p.readedDataLen = 0 //读完头， 把已读内容长度设置为0
	return ok, closed, err
}

func (p *Packet) ReadContent(connector *tcp.Connector) (ok bool, closed bool, err error) {
	closed = false
	ok = false
	tmpdata := p.readBuff[:len(p.readBuff)]
	remainLen := p.header.dataLen - p.readedDataLen
	if remainLen <= bufflen {
		tmpdata = p.readBuff[:p.header.dataLen]
	}
	_, err = io.ReadFull(connector.Conn, tmpdata)
	if remainLen > bufflen {
		p.readedDataLen = p.readedDataLen + bufflen
	} else {
		p.readedDataLen = p.readedDataLen + remainLen
	}
	ok = p.header.dataLen == p.readedDataLen //是否已读完一个请求包的内容
	handle, exist := HandlerMap[p.header.iD]
	if exist {
		connector.RLock()
		if !connector.Closed {
			atomic.AddInt32(&connector.RefCount, 1)
		} else {
			return ok, true, err
		}
		defer connector.RUnlock()
		packData := make([]byte, len(tmpdata))
		copy(packData, tmpdata)
		go func() {
			defer func() {
				atomic.AddInt32(&connector.RefCount, -1)
				connector.HandleEnd()
			}()
			handle.Handler(connector, ok, packData)
		}()
	}
	return ok, false, err
}

func (p *Packet) NewPacket() tcpReadPacket.IPacket {
	return &Packet{}
}

func (p *Packet) Init() {
	p.readedDataLen = 0
}

func NewtcpServe(addr string) *tcp.Server {
	return tcpReadPacket.NewtcpServe(addr, &Packet{})
}
