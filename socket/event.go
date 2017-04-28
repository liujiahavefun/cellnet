package socket

import (
	"fmt"

	"cellnet"
	_ "cellnet/proto/gamedef"
)

var (
	Event_SessionAccepted      = uint32(cellnet.MessageMetaByName("gamedef.SessionAccepted").ID)
	Event_SessionAcceptFailed  = uint32(cellnet.MessageMetaByName("gamedef.SessionAcceptFailed").ID)
	Event_SessionConnected     = uint32(cellnet.MessageMetaByName("gamedef.SessionConnected").ID)
	Event_SessionClosed        = uint32(cellnet.MessageMetaByName("gamedef.SessionClosed").ID)
	Event_SessionConnectFailed = uint32(cellnet.MessageMetaByName("gamedef.SessionConnectFailed").ID)
)

//会话事件
type SessionEvent struct {
	*cellnet.Packet
	Ses cellnet.Session
}

func (self SessionEvent) String() string {
	return fmt.Sprintf("SessionEvent msgid: %d data: %v", self.MsgID, self.Data)
}

func NewSessionEvent(msgid uint32, s cellnet.Session, data []byte) *SessionEvent {
	return &SessionEvent{
		Packet: &cellnet.Packet{MsgID: msgid, Data: data},
		Ses:    s,
	}
}

func newSessionEvent(msgid uint32, s cellnet.Session, msg interface{}) *SessionEvent {
	pkt, _ := cellnet.BuildPacket(msg)
	return &SessionEvent{
		Packet: pkt,
		Ses:    s,
	}
}
