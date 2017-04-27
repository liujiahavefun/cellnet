package socket

import (
	"sync"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/gamedef"
)

type tcpClientSession struct {
	id int64

	fromPeer cellnet.Peer

	endSync sync.WaitGroup

	needNotifyWrite bool // 是否需要通知写线程关闭

	stream *ltvStream

	sendList *PacketList

	OnClose func() // 关闭函数回调
}

func newSession(stream *ltvStream, eq cellnet.EventQueue, peer cellnet.Peer) *tcpClientSession {
	self := &tcpClientSession{
		stream:          stream,
		fromPeer:        peer,
		needNotifyWrite: true,
		sendList:        NewPacketList(),
	}

	// 使用peer的统一设置
	self.stream.maxPacketSize = peer.MaxPacketSize()

	// 布置接收和发送2个任务
	// bug fix感谢viwii提供的线索
	self.endSync.Add(2)

	go func() {

		// 等待2个任务结束
		self.endSync.Wait()

		// 在这里断开session与逻辑的所有关系
		if self.OnClose != nil {
			self.OnClose()
		}
	}()

	// 接收线程
	go self.recvThread(eq)

	// 发送线程
	go self.sendThread()

	return self
}

func (self *tcpClientSession) ID() int64 {
	return self.id
}

func (self *tcpClientSession) FromPeer() cellnet.Peer {
	return self.fromPeer
}

func (self *tcpClientSession) Close() {
	self.sendList.Add(&cellnet.Packet{})
}

func (self *tcpClientSession) Send(data interface{}) {

	pkt, _ := cellnet.BuildPacket(data)

	msgLog("send", self, pkt)

	self.RawSend(pkt)
}

func (self *tcpClientSession) RawSend(pkt *cellnet.Packet) {

	if pkt != nil {
		self.sendList.Add(pkt)
	}
}

//发送线程
func (self *tcpClientSession) sendThread() {
	var writeList []*cellnet.Packet

	for {
		willExit := false
		writeList = writeList[0:0]

		//从队列中拷贝所有待发送的packet
		packetList := self.sendList.BeginPick()

		for _, packet := range packetList {
			//用特殊的msg来使发送线程退出
			if packet.MsgID == 0 {
				willExit = true
			} else {
				writeList = append(writeList, packet)
			}
		}

		self.sendList.EndPick()

		//依次发送每一个packet
		for _, packet := range writeList {
			if err := self.stream.Write(packet); err != nil {
				//TODO: 这里应该日志记录，并且onError()回调
				willExit = true
				break
			}
		}

		//flush socket
		if err := self.stream.Flush(); err != nil {
			willExit = true
		}

		if willExit {
			goto EXIT_SEND_LOOP
		}
	}

EXIT_SEND_LOOP:
	//不需要读线程再次通知写线程
	self.needNotifyWrite = false

	//关闭socket,触发读错误, 结束读循环
	self.stream.Close()

	//通知发送线程退出
	self.endSync.Done()
}

//接收线程
func (self *tcpClientSession) recvThread(eq cellnet.EventQueue) {
	var err error
	var pkt *cellnet.Packet

	for {
		//从Socket读取封包
		pkt, err = self.stream.Read()
		if err != nil {
			ev := newSessionEvent(Event_SessionClosed, self, &gamedef.SessionClosed{Reason: err.Error()})
			msgLog("recv", self, ev.Packet)

			//post断开事件
			eq.Post(self.fromPeer, ev)
			break
		}

		// 消息日志要多损耗一次解析性能
		msgLog("recv", self, pkt)

		// 逻辑封包
		eq.Post(self.fromPeer, &SessionEvent{
			Packet: pkt,
			Ses:    self,
		})
	}

	if self.needNotifyWrite {
		self.Close()
	}

	//通知接收线程退出
	self.endSync.Done()
}