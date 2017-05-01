package socket

import (
	"net"

	"cellnet"
	"cellnet/proto/gamedef"
	"fmt"
)

type TcpServer struct {
	*peerBase
	*sessionMgr

	cellnet.EventDispatcher
	evq cellnet.EventQueue

	//sessionCallback *SessionCallback

	address string
	running bool //TODO: 用atomic代替
	listener net.Listener
}

func NewTcpServer(evq cellnet.EventQueue) cellnet.Server {
	self := &TcpServer{
		EventDispatcher: cellnet.NewEventDispatcher(),
		evq: 		evq,
		peerBase:   newPeerBase(),
		sessionMgr: newSessionManager(),
	}

	/*
	self.sessionCallback = NewSessionCallback(self.onSessionConnectedFunc,
											  self.onSessionClosedFunc,
											  self.onSessionErrorFunc,
											  self.onSessionRecvPacketFunc)
	*/

	return self
}

func (self *TcpServer) Start(address string) cellnet.Server {
	ln, err := net.Listen("tcp", address)
	self.listener = ln
	if err != nil {
		logErrorf("#listen failed(%s) %v", self.name, err.Error())
		return self
	}

	self.address = address
	self.running = true
	logInfof("#listen(%s) %s ", self.name, address)

	// 接受线程
	go func() {
		for self.running {
			conn, err := ln.Accept()
			if err != nil {
				//TODO: onErrorFunc instead, 别随地胡逼抛事件
				logErrorf("#accept failed(%s) %v", self.name, err.Error())
				self.evq.Post(self, newSessionEvent(Event_SessionAcceptFailed, nil, &gamedef.SessionAcceptFailed{Reason: err.Error()}))
				break
			}

			//处理连接进入独立线程, 防止accept无法响应
			go func() {
				session := newServerSession(conn, self)

				//断开后从管理器移除
				//TODO: 这里可以再给外部一个回调，或者post一个事件
				session.onSessionClosedFunc = self.onSessionClosedFunc
				session.onSessionErrorFunc = self.onSessionErrorFunc
				session.onSessionRecvPacketFunc = self.onSessionRecvPacketFunc

				//添加到管理器
				self.sessionMgr.Add(session)

				logInfof("#accepted(%s) sid: %d", self.name, session.GetID())

				//通知逻辑
				//self.Post(self, NewSessionEvent(Event_SessionAccepted, session, nil))
				self.onSessionConnectedFunc(session)
			}()
		}
	}()

	return self
}

func (self *TcpServer) Stop() {
	if !self.running {
		return
	}

	self.running = false
	self.listener.Close()
}

func (self *TcpServer) IsRunning() bool {
	return self.running
}

func (self *TcpServer) GetAddress() string {
	return self.address
}

func (self *TcpServer) onSessionConnectedFunc(session cellnet.Session) {
	fmt.Println("liujia, tcp_server onSessionConnectedFunc: ", session)
	self.evq.Post(self, NewSessionEvent(Event_SessionAccepted, session, nil))
}

func (self *TcpServer) onSessionClosedFunc(session cellnet.Session) {
	fmt.Println("liujia, tcp_server onSessionClosedFunc: ", session)
	self.sessionMgr.Remove(session)

	/*
	ev := newSessionEvent(Event_SessionClosed, session, &gamedef.SessionClosed{Reason: err.Error()})
	msgLog("recv", session, ev.Packet)

	//post断开事件
	self.evq.Post(self, ev)
	*/
}

func (self *TcpServer) onSessionErrorFunc(session cellnet.Session, err error) {
	fmt.Println("liujia, tcp_server onSessionErrorFunc: ", session, err)
	//TODO: Event_SessionClosed to Event_SessionError
	ev := newSessionEvent(Event_SessionClosed, session, &gamedef.SessionClosed{Reason: err.Error()})

	//post断开事件
	self.evq.Post(self, ev)
}

func (self *TcpServer) onSessionRecvPacketFunc(session cellnet.Session, packet *cellnet.Packet) {
	fmt.Println("liujia, tcp_server onSessionRecvPacketFunc: ", session, packet)
	msgLog("recv", session, packet)
	self.evq.Post(self, &SessionEvent{
		Packet: packet,
		Ses:    session,
	})
}