package tcp

/*
	传入实现ITcpReader的对像，完成对Socket的读取操作
*/
import (
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
)

///枚举类型
type tcpStatus int

const (
	_          tcpStatus = iota
	connecting           //连接上来
	connected            //已连接，正常通信中
	disconnect           //对端关闭连接
	readerr              //读取时出错
	senderr              //发送时出错
	connecterr           //连接时出错
	shutdown             //主动关闭连接
)

type (
	TOnConnect    func(conn net.Conn) bool
	TOnDisConnect func(conn net.Conn)
	TOnError      func(conn net.Conn, err error)
)

type ITcpReaderWriter interface {
	ReadData(connector *Connector) (closed bool, err error)
	WriteData(connector *Connector, dataEx any, data []byte)
	Init()
	NewReaderWriter() ITcpReaderWriter
}

type Connector struct {
	ConID        uint32
	Conn         net.Conn
	SendDataChan chan []byte
	status       tcpStatus
	svr          Server
	//ctx          context.Context
	//cancel       context.CancelFunc
	connectChan  chan bool
	err          error
	closeInt     int32 //1为打开状态，2为关闭状态
	readerwriter ITcpReaderWriter
}

type Server struct {
	addr          string //监听字符串127.0.0.1:1000
	listener      net.Listener
	doChan        chan *Connector
	connectorPool sync.Pool
	FOnConnect    TOnConnect
	FOnDisConnect TOnDisConnect
	FOnError      TOnError
}

func (connector *Connector) init() {
	connector.Conn = nil
	connector.status = connecting
	connector.err = nil
	connector.closeInt = 1
	connector.readerwriter.Init()
}

func NewServe(addr string, readImpl ITcpReaderWriter) *Server {
	result := &Server{
		addr:   addr,
		doChan: make(chan *Connector),
	}
	result.connectorPool = sync.Pool{New: func() any {
		return &Connector{status: connecting, connectChan: make(chan bool), SendDataChan: make(chan []byte, 10000), readerwriter: readImpl.NewReaderWriter()}
	}}
	return result
}

func (svr *Server) broadcast() {
	ok := false
	var connector *Connector
	for {
		select {
		case connector, ok = <-svr.doChan:
			if connector.status == connecting {
				if svr.FOnConnect != nil {
					connector.connectChan <- svr.FOnConnect(connector.Conn)
				} else {
					connector.connectChan <- true
				}
			} else if (connector.status == disconnect) || (connector.status == shutdown) {
				if svr.FOnDisConnect != nil {
					svr.FOnDisConnect(connector.Conn)
				}
			} else if connector.status == readerr || connector.status == connecterr {
				if svr.FOnError != nil {
					svr.FOnError(connector.Conn, connector.err)
				}
				if connector.status == connecterr {
					connector.init()
					svr.connectorPool.Put(connector)
				}
			}
		}
		if !ok {
			break
		}
	}
}

func (svr *Server) Start() error {
	var err error
	svr.listener, err = net.Listen("tcp", svr.addr)
	if err != nil {
		return err
	}

	go svr.broadcast()

	var id uint32 = 0
	for {
		conn, err := svr.listener.Accept()
		connector := svr.connectorPool.Get().(*Connector)
		connector.Conn = conn

		id++
		connector.ConID = id
		if id == math.MaxUint32 {
			id = 0
		}

		if err != nil {
			connector.err = err
			connector.status = connecterr
			svr.doChan <- connector
			continue
		}
		//connector.ctx, connector.cancel = context.WithCancel(context.Background())
		go svr.handlerConnector(connector) //处理接收数据
		go connector.connectorSendData()   //处理发送数据
	}
}

func (svr *Server) Stop() {
	_ = svr.listener.Close()
	close(svr.doChan)
}

func (svr *Server) handlerConnector(connector *Connector) {
	defer func() {
		connector.Close()
		connector.init()
		svr.connectorPool.Put(connector)
	}()
	svr.doChan <- connector
	if !<-connector.connectChan {
		return
	}

	connector.status = connected
	closed := false
	var err error
	for {
		select {
		default:
			{
				closed, err = connector.readerwriter.ReadData(connector)
				if err != nil {
					if err == io.EOF {
						connector.status = disconnect
						svr.doChan <- connector
					} else {
						connector.status = readerr
						connector.err = err
						svr.doChan <- connector
					}
					return
				}

				if closed {
					connector.status = shutdown
					svr.doChan <- connector
					return
				}
			}
		}
	}
}

func (connector *Connector) connectorSendData() {
	for {
		select {
		case senddata, ok := <-connector.SendDataChan:
			{
				if ok {
					_, err := connector.Conn.Write(senddata)
					if err != nil {
						connector.status = senderr
						connector.err = err
						connector.svr.doChan <- connector
						return
					}
				} else {
					return
				}

			}
		}

	}
}

func (connector *Connector) SendData(dataEx any, data []byte) {
	connector.readerwriter.WriteData(connector, dataEx, data)
}

func (connector *Connector) Close() {
	if !atomic.CompareAndSwapInt32(&connector.closeInt, 1, 2) {
		return
	}
	_ = connector.Conn.Close()
	close(connector.SendDataChan)
	//connector.cancel()
}
