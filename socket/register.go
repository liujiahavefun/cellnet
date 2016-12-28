package socket

import "github.com/davyxu/cellnet"

// 注册连接消息
func RegisterSessionMessage(eq cellnet.EventQueue, msgName string, userHandler func(interface{}, cellnet.Session)) *cellnet.MessageMeta {

	msgMeta := cellnet.MessageMetaByName(msgName)

	if msgMeta == nil {
		log.Errorf("message register failed, %s", msgName)
		return nil
	}

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*SessionEvent); ok {

			rawMsg, err := cellnet.ParsePacket(ev.Packet, msgMeta.Type)

			if err != nil {
				log.Errorln("unmarshaling error:\n", err)
				return
			}

			userHandler(rawMsg, ev.Ses)

		}

	})

	return msgMeta
}

type Context struct {
	index int
	list  []Handler
}

func (self *Context) Next() interface{} {
	return self.list[self.index+1]
}

type IMessageContext interface {
	set(interface{}, cellnet.Session)
}

type Handler interface {
	Exec(context *Context)
}

type MessageContext struct {
	Msg interface{}

	Ses cellnet.Session
}

func (self *MessageContext) set(msg interface{}, ses cellnet.Session) {
	self.Msg = msg
	self.Ses = ses
}

// 注册连接消息
func RegisterSessionMessage2(eq cellnet.EventQueue, msgName string, userHandlers ...Handler) *cellnet.MessageMeta {

	msgMeta := cellnet.MessageMetaByName(msgName)

	if msgMeta == nil {
		log.Errorf("message register failed, %s", msgName)
		return nil
	}

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*SessionEvent); ok {

			rawMsg, err := cellnet.ParsePacket(ev.Packet, msgMeta.Type)

			if err != nil {
				log.Errorln("unmarshaling error:\n", err)
				return
			}

			ctx := &Context{list: userHandlers}

			for index, h := range userHandlers {

				ctx.index = index

				mc := h.(IMessageContext)
				mc.set(rawMsg, ev.Ses)

				h.Exec(ctx)

			}

		}

	})

	return msgMeta
}
