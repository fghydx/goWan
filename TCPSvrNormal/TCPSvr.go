package TCPSvrNormal

/*
	TCP
*/

import (
	"github.com/fghydx/gobone/ToolsOther"
	"io"
	"net"
	"reflect"
	"strconv"
	"sync"
)

type TcpNet struct {
	chanConnect chan tcpStatusMsg
	listen      net.Listener
	ListenAddr  string
	ListenPort  int
	NetObjPool  sync.Pool
	frameIntf   reflect.Type
}

type TcpNetObj struct {
	TcpNetObj ITcpNet
	conn      net.Conn
}

///枚举类型
type tcpStatus int

const (
	_          tcpStatus = iota
	connected            = 1 //连接上来
	disconnect           = 2 //对端关闭连接
	errorex              = 3 //出错
	shutdown             = 4 //主动关闭连接
)

type tcpStatusMsg struct {
	netObj  TcpNetObj
	status  tcpStatus
	errorex error
	tag     chan byte
}

func NewNetFrame(ListenAddr string, ListenPort int, impl interface{}) *TcpNet {
	if _, ok := impl.(ITcpNet); !ok {
		panic("没有实现接口")
	}
	result := &TcpNet{chanConnect: make(chan tcpStatusMsg), ListenAddr: ListenAddr, ListenPort: ListenPort, frameIntf: ToolsOther.GetObjRefType(impl)}
	result.NetObjPool = sync.Pool{New: func() any {
		var TcpObj TcpNetObj
		TcpObj.conn = nil
		TcpObj.TcpNetObj = (ToolsOther.CreateObjFromType(result.frameIntf)).(ITcpNet)
		return TcpObj
	}}
	return result
}

func (NetFrame *TcpNet) msgbordcast() {
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
			} else if (tcpMsg.status == disconnect) || (tcpMsg.status == shutdown) {
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

func (NetFrame *TcpNet) Start() error {
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
		TcpObj := NetFrame.NetObjPool.Get().(TcpNetObj)
		TcpObj.conn = conn
		go TcpObj.handleConnection(NetFrame)
	}
}

func (NetFrame *TcpNet) Stop() {
	NetFrame.listen.Close()
	close(NetFrame.chanConnect)
}

func (Netobj *TcpNetObj) handleConnection(NetFrame *TcpNet) {
	defer func() {
		Netobj.conn = nil
		NetFrame.NetObjPool.Put(*Netobj)
	}()
	conn := Netobj.conn
	defer conn.Close()
	tagchan := make(chan byte)
	NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: nil, status: connected}
	if <-tagchan == 1 {
		return //println("不允许连接!")
	}

	closed := false
	var err error
	for {
		closed, err = Netobj.TcpNetObj.OnReadData(conn)
		if err != nil {
			if err == io.EOF {
				NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: err, status: disconnect}
			} else {
				NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: err, status: errorex}
			}
			break
		}

		if closed {
			NetFrame.chanConnect <- tcpStatusMsg{netObj: *Netobj, tag: tagchan, errorex: nil, status: shutdown}
			break
		}
	}
}
