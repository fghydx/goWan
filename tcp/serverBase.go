package tcp

/*
	传入实现ITcpReader的对像，完成对Socket的读取操作
*/
import (
	"context"
	"fmt"
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
	sync.RWMutex
	logidx       uint32
	ConID        uint32
	Conn         net.Conn
	SendDataChan chan []byte
	RefCount     int32
	status       tcpStatus
	svr          *Server
	ctx          context.Context
	cancel       context.CancelFunc
	connectChan  chan bool
	err          error
	Closed       bool
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

var ConnectCount int32 = 0
var conidx uint32 = 0

func (connector *Connector) init() {
	connector.Conn = nil
	connector.status = connecting
	connector.err = nil
	connector.Closed = false
	connector.readerwriter.Init()
}

func NewServe(addr string, readImpl ITcpReaderWriter) *Server {
	result := &Server{
		addr:   addr,
		doChan: make(chan *Connector, 1000),
	}
	result.connectorPool = sync.Pool{New: func() any {
		return &Connector{status: connecting, readerwriter: readImpl.NewReaderWriter(), svr: result, logidx: atomic.AddUint32(&conidx, 1)}
	}}
	return result
}

func (svr *Server) broadcast() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	ok := false
	var connector *Connector
	for {
		select {
		case connector, ok = <-svr.doChan:

			if connector.status == connecting {
				//fmt.Println(connector.ConID, connector.logidx, "connecting", connector.status)
				if svr.FOnConnect != nil {
					connector.connectChan <- svr.FOnConnect(connector.Conn)
				} else {
					connector.connectChan <- true
				}
				//fmt.Println(connector.ConID, connector.logidx, "connecting Over", connector.status)
			} else {
				close(connector.SendDataChan)
				close(connector.connectChan)
				if (connector.status == disconnect) || (connector.status == shutdown) {
					if svr.FOnDisConnect != nil {
						svr.FOnDisConnect(connector.Conn)
					}
				} else if connector.status == readerr || connector.status == connecterr {
					if svr.FOnError != nil {
						svr.FOnError(connector.Conn, connector.err)
					}
					if connector.status == connecterr {

					}
				}
				svr.connectorPool.Put(connector)
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
		if atomic.LoadInt32(&ConnectCount) > 10000 {
			//fmt.Println("当前连接数，", ConnectCount)
			continue
		}
		connector := svr.connectorPool.Get().(*Connector)
		connector.init()
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
		go connector.start()
	}
}

func (svr *Server) Stop() {
	_ = svr.listener.Close()
	close(svr.doChan)
}

func (connector *Connector) sendData() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err, connector.logidx)
		}
	}()
	for {
		select {
		case <-connector.ctx.Done():
			{
				//fmt.Println("离开sendData", connector.logidx)
				return
			}

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

func (connector *Connector) recvData() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err, connector.logidx)
		}
	}()
	connector.status = connected
	closed := false
	var err error
	atomic.AddInt32(&connector.RefCount, 1)
end:
	for {
		select {
		case <-connector.ctx.Done():
			{
				//fmt.Println("ctx取消离开recvData", connector.logidx)
				break end
			}
		default:
			{
				closed, err = connector.readerwriter.ReadData(connector)
				if err != nil {
					if err == io.EOF {
						connector.status = disconnect
						connector.svr.doChan <- connector
					} else {
						connector.status = readerr
						connector.err = err
						connector.svr.doChan <- connector
					}
					break end
				}

				if closed {
					connector.status = shutdown
					connector.svr.doChan <- connector
					break end
				}
			}
		}
	}
	connector.Stop()
	atomic.AddInt32(&connector.RefCount, -1)
	connector.HandleEnd()
}

func (connector *Connector) SendData(dataEx any, data []byte) {
	connector.readerwriter.WriteData(connector, dataEx, data)
}

func (connector *Connector) start() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err, connector.logidx)
		}
	}()
	connector.connectChan = make(chan bool)
	//fmt.Println(connector.ConID, connector.logidx, "连接上来了")
	connector.svr.doChan <- connector
	if !<-connector.connectChan {
		connector.status = disconnect
		connector.svr.doChan <- connector
		return
	}
	connector.SendDataChan = make(chan []byte, 1000)
	atomic.AddInt32(&ConnectCount, 1)
	connector.ctx, connector.cancel = context.WithCancel(context.Background())
	go connector.recvData()
	go connector.sendData() //处理发送数据
}

func (connector *Connector) Stop() {
	connector.Lock()
	defer connector.Unlock()
	if connector.Closed {
		return
	}
	connector.Closed = true

	//fmt.Println(connector.ConID, connector.logidx, "Stop()")
	connector.cancel()
	atomic.AddInt32(&ConnectCount, -1)
}

func (connector *Connector) CheckClosed() bool {
	connector.RLock()
	defer connector.RUnlock()
	if !connector.Closed {
		return false
	}
	return true
}

func (connector *Connector) HandleEnd() {
	if connector.CheckClosed() {
		if atomic.LoadInt32(&connector.RefCount) == 0 {
			connector.Conn.Close()
		}
	}
}
