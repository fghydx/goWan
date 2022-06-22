package tcpCustomHead

/*
	实现了协议为ID+内容长度+内容的传输
*/

import (
	"encoding/binary"
	"github.com/fghydx/goWan/tcp"
	"github.com/fghydx/goWan/tcpReadPacket"
	"io"
)

type THandler = func()

func RegisterHandler(ID uint32, handler THandler) {

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
	handler       IHandler
}

func (p *Packet) ReadHead(connector *tcp.Connector) (ok bool, closed bool, err error) {
	ok = true
	closed = false
	err = binary.Read(connector.Conn, binary.LittleEndian, &p.header.iD)
	err = binary.Read(connector.Conn, binary.LittleEndian, &p.header.dataLen)
	p.readedDataLen = 0 //读完头， 把已读内容长度设置为0
	println(p.header.iD, p.header.dataLen)
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
	p.handler.Handler(connector, ok, tmpdata)
	return
}

func (p *Packet) NewPacket() tcpReadPacket.IPacket {
	return &Packet{}
}

func (p *Packet) Init() {
	p.readedDataLen = 0
}

func NewtcpServe(addr string, handler IHandler) *tcp.Server {
	return tcpReadPacket.NewtcpServe(addr, &Packet{})
}
