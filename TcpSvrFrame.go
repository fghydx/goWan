package main

import (
	"github.com/fghydx/gobone/ToolsOther"
	"io"
	"net"
	"reflect"
	"strconv"
	"sync"
)

type TcpNetIntf interface {
	OnConnect(conn net.Conn) bool
	OnDisconnect(conn net.Conn)
	OnError(conn net.Conn, err error)
	OnRecv(conn net.Conn, Adata []byte, len int) bool
}

type TcpNetFrame struct {
	chanConnect chan tcpStatusMsg
	listen      net.Listener
	ListenAddr  string
	ListenPort  int
	NetObjPool  sync.Pool
	frameIntf   reflect.Type
}

type GLTcpNetObj struct {
	TcpNetObj TcpNetIntf
	conn      net.Conn
}

///枚举类型
type tcpStatus int

const (
	_          tcpStatus = iota
	connected            = 1 //连接上来
	disconnect           = 2 //对端关闭连接
	errorex              = 3 //出错
	close                = 4 //主动关闭连接
)

type tcpStatusMsg struct {
	netObj  GLTcpNetObj
	status  tcpStatus
	errorex error
	tag     chan byte
}

func NewNetFrame(ListenAddr string, ListenPort int, impl interface{}) *TcpNetFrame {
	if _, ok := impl.(TcpNetIntf); !ok {
		panic("没有实现接口")
	}
	result := &TcpNetFrame{chanConnect: make(chan tcpStatusMsg), ListenAddr: ListenAddr, ListenPort: ListenPort, frameIntf: ToolsOther.GetObjType(impl)}
	result.NetObjPool = sync.Pool{New: func() any {
		var TcpObj GLTcpNetObj
		TcpObj.conn = nil
		TcpObj.TcpNetObj = ToolsOther.CreateObjFromType(result.frameIntf).(TcpNetIntf)
		return TcpObj
	}}
	return result
}

func (NetFrame *TcpNetFrame) msgbordcast() {
	ok := false
	var tcpMsg tcpStatusMsg
	for {
		select {
		case tcpMsg, ok = <-NetFrame.chanConnect:
			if tcpMsg.status == connected {
				if !tcpMsg.netObj.TcpNetObj.OnConnect(tcpMsg.netObj.conn) {
					tcpMsg.tag <- 1
				} else {
					tcpMsg.tag <- 0
				}
			} else if (tcpMsg.status == disconnect) || (tcpMsg.status == close) {
				tcpMsg.netObj.TcpNetObj.OnDisconnect(tcpMsg.netObj.conn)
			} else if tcpMsg.status == errorex {
				tcpMsg.netObj.TcpNetObj.OnError(tcpMsg.netObj.conn, tcpMsg.errorex)
			}
		}
		if !ok {
			break
		}
	}
}

func (NetFrame *TcpNetFrame) Start() error {
	var err error
	NetFrame.listen, err = net.Listen("tcp", NetFrame.ListenAddr+":"+strconv.Itoa(NetFrame.ListenPort))
	if err != nil {
		return err
	}
	defer NetFrame.listen.Close()

	go NetFrame.msgbordcast()

	for {
		conn, err := NetFrame.listen.Accept()
		if err != nil {
			panic(err)
		}
		TcpObj := NetFrame.NetObjPool.Get().(GLTcpNetObj)
		TcpObj.conn = conn
		go TcpObj.handleConnection(NetFrame)
	}
}

func (NetFrame *TcpNetFrame) Stop() {
	NetFrame.listen.Close()
}

func (Netobj *GLTcpNetObj) handleConnection(NetFrame *TcpNetFrame) {
	defer func() {
		Netobj.conn = nil
		NetFrame.NetObjPool.Put(*Netobj)
	}()
	conn := Netobj.conn
	defer conn.Close()
	tagchan := make(chan byte)
	NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: nil, status: connected}
	if <-tagchan == 1 {
		//println("不允许连接!")
		return
	}

	CurData := make([]byte, 1024)
	for {
		readlen, err := conn.Read(CurData)
		if err != nil {
			if err == io.EOF {
				NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: err, status: disconnect}
			} else {
				NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: err, status: errorex}
			}
			break
		}

		if !Netobj.TcpNetObj.OnRecv(conn, CurData, readlen) {
			NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: nil, status: close}
			break
		}
	}
}
