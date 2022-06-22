package tcpReadPacket

import (
	"encoding/binary"
	"github.com/fghydx/goWan/tcp"
	"io"
	"testing"
)

type Packet struct {
	head     uint32
	readBuff [4096]byte
}

func (p *Packet) ReadHead(connector *tcp.Connector) (ok bool, closed bool, err error) {
	ok = true
	closed = false
	err = nil
	err = binary.Read(connector.Conn, binary.LittleEndian, &p.head)
	println("读完头了", p.head)
	return false, closed, err
}

func (p *Packet) ReadContent(connector *tcp.Connector) (ok bool, closed bool, err error) {
	ok = true
	closed = false
	err = nil
	b := p.readBuff[:2]
	_, err = io.ReadFull(connector.Conn, b)
	println("读完内容了", b)
	return ok, closed, err
}

func (p *Packet) NewPacket() IPacket {
	return &Packet{}
}

func (p *Packet) Init() {

}

func TestNewtcpServe(t *testing.T) {
	test := NewtcpServe("127.0.0.1:8889", &Packet{})
	err := test.Start()
	if err != nil {
		println(err)
	}
}
