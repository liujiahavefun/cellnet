package socket

import "cellnet"

//Peer间的共享数据
type peerBase struct {
	cellnet.EventDispatcher
	cellnet.EventQueue

	name          string
	maxPacketSize int
}

func newPeerBase(evq cellnet.EventQueue) *peerBase {
	self := &peerBase{
		EventDispatcher: cellnet.NewEventDispatcher(),
		EventQueue:      evq,
	}

	return self
}

func (self *peerBase) SetName(name string) {
	self.name = name
}

func (self *peerBase) Name() string {
	return self.name
}

func (self *peerBase) SetMaxPacketSize(size int) {
	self.maxPacketSize = size
}

func (self *peerBase) MaxPacketSize() int {
	return self.maxPacketSize
}
