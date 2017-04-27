package socket

import (
	"net"

	"cellnet"
	"cellnet/proto/gamedef"
)

type TcpServer struct {
	*peerBase
	*sessionMgr

	listener net.Listener
	running bool //TODO: 用atomic代替
}

func NewTcpServer(evq cellnet.EventQueue) cellnet.Peer {
	self := &TcpServer{
		sessionMgr: newSessionManager(),
		peerBase:   newPeerBase(evq),
	}

	return self
}

func (self *TcpServer) Start(address string) cellnet.Peer {
	ln, err := net.Listen("tcp", address)
	self.listener = ln
	if err != nil {
		logErrorf("#listen failed(%s) %v", self.name, err.Error())
		return self
	}

	self.running = true
	logInfof("#listen(%s) %s ", self.name, address)

	// 接受线程
	go func() {
		for self.running {
			conn, err := ln.Accept()
			if err != nil {
				logErrorf("#accept failed(%s) %v", self.name, err.Error())
				self.Post(self, newSessionEvent(Event_SessionAcceptFailed, nil, &gamedef.SessionAcceptFailed{Reason: err.Error()}))
				break
			}

			//处理连接进入独立线程, 防止accept无法响应
			go func() {
				session := newSession(NewPacketStream(conn), self, self)

				//添加到管理器
				self.sessionMgr.Add(session)

				//断开后从管理器移除
				//TODO: 这里可以再给外部一个回调，或者post一个事件
				session.OnClose = func() {
					self.sessionMgr.Remove(session)
				}

				logInfof("#accepted(%s) sid: %d", self.name, session.GetID())

				//通知逻辑
				self.Post(self, NewSessionEvent(Event_SessionAccepted, session, nil))
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
