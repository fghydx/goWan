package tcpReadPacket

/*
	传入实现IPacket的对像，完成对区分Socket的头与内容的读取操作,有自定义的头时可用这个单元
*/

import (
	"github.com/fghydx/goWan/tcp"
)

type EnumReadType uint8

const (
	_ EnumReadType = iota
	ReadHead
	ReadContent
)

type IPacket interface {
	ReadHead(connector *tcp.Connector) (ok bool, closed bool, err error)
	ReadContent(connector *tcp.Connector) (ok bool, closed bool, err error)
	PackData(head any, data []byte) []byte
	NewPacket() IPacket
	Init()
}

type TcpReaderWriter struct {
	readType EnumReadType
	ipacket  IPacket
}

func (t *TcpReaderWriter) WriteData(connector *tcp.Connector, dataEx any, data []byte) {
	connector.SendDataChan <- t.ipacket.PackData(dataEx, data)
}

func (t *TcpReaderWriter) ReadData(connector *tcp.Connector) (closed bool, err error) {
	closed = false
	err = nil
	ok := false
	switch t.readType {
	case ReadHead:
		{
			ok, closed, err = t.ipacket.ReadHead(connector)
			if ok {
				t.readType = ReadContent
			}
		}
	case ReadContent:
		{
			ok, closed, err = t.ipacket.ReadContent(connector)
			if ok {
				t.readType = ReadHead
			}
		}
	}
	return closed, err
}

func NewtcpServe(addr string, impl IPacket) *tcp.Server {
	tcpreader := &TcpReaderWriter{
		readType: ReadHead,
		ipacket:  impl,
	}
	result := tcp.NewServe(addr, tcpreader)
	return result
}

func (t *TcpReaderWriter) Init() {
	t.readType = ReadHead
	t.ipacket.Init()
}

func (t *TcpReaderWriter) NewReaderWriter() tcp.ITcpReaderWriter {
	result := &TcpReaderWriter{
		readType: ReadHead,
	}
	result.ipacket = t.ipacket.NewPacket()
	return result
}
