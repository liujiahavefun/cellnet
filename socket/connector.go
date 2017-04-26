package socket

import (
	"net"
	"time"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/gamedef"
)

type socketConnector struct {
	*peerBase
	*sessionMgr

	//底层的net.Conn
	conn net.Conn

	//重连间隔时间, 0为不重连
	autoReconnectSec int

	//尝试连接次数
	tryConnTimes int

	//重入锁
	working bool

	//等待关闭的chan
	closeSignal chan bool

	defaultSes cellnet.Session
}

const (
	DEFAULT_CONNECT_RETRY_TIMES = 3
)

func NewConnector(evq cellnet.EventQueue) cellnet.Peer {
	self := &socketConnector{
		sessionMgr:  newSessionManager(),
		peerBase:    newPeerBase(evq),
		closeSignal: make(chan bool),
	}

	return self
}

//设置自动重连间隔, 秒为单位，0表示不重连
func (self *socketConnector) SetAutoReconnectSec(sec int) {
	self.autoReconnectSec = sec
}

//启动，去连接
func (self *socketConnector) Start(address string) cellnet.Peer {
	if self.working {
		return self
	}

	go self.connect(address)
	return self
}

//连接，注意重连是会阻塞的，并且连上之后也是阻塞的，所以这个函数要在单独的goroutine里被调用
func (self *socketConnector) connect(address string) {
	self.working = true

	for {
		self.tryConnTimes++

		//去连接
		conn, err := net.Dial("tcp", address)
		if err != nil {
			if self.tryConnTimes <= DEFAULT_CONNECT_RETRY_TIMES {
				log.Errorf("#connect failed(%s) %v", self.name, err.Error())
			}

			if self.tryConnTimes == DEFAULT_CONNECT_RETRY_TIMES {
				log.Errorf("(%s) continue reconnecting, but mute log", self.name)
			}

			//没重连就退出
			if self.autoReconnectSec == 0 {
				self.Post(self, newSessionEvent(Event_SessionConnectFailed, nil, &gamedef.SessionConnectFailed{Reason: err.Error()}))
				break
			}

			//有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			//继续连接
			continue
		}

		self.tryConnTimes = 0

		//连上了, 记录连接
		self.conn = conn

		//创建Session
		ses := newSession(NewPacketStream(conn), self, self)
		self.sessionMgr.Add(ses)
		self.defaultSes = ses

		log.Infof("#connected(%s) %s sid: %d", self.name, address, ses.id)

		//内部断开回调
		ses.OnClose = func() {
			self.sessionMgr.Remove(ses)
			self.closeSignal <- true
		}

		// 抛出事件
		self.Post(self, NewSessionEvent(Event_SessionConnected, ses, nil))

		//等待连接关闭
		if <-self.closeSignal {
			self.conn = nil

			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			// 继续连接
			continue
		}
	}

	self.working = false
}

func (self *socketConnector) Stop() {
	if self.conn != nil {
		//这个调用会导致session的Close，从而调用了我们设置的OnClose回调，最终self.closeSignal收到信号，关闭
		self.conn.Close()
	}
}

func (self *socketConnector) DefaultSession() cellnet.Session {
	return self.defaultSes
}
