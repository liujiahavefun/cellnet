package echo

import (
	"testing"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/gamedef"
	"github.com/davyxu/cellnet/socket"
	"github.com/davyxu/cellnet/test"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

type Player struct {
	name string
}

type PlayerConvertor struct {
	socket.MessageContext
}

func (self *PlayerConvertor) Exec(context *socket.Context) {
	context.Next().(interface {
		SetPlayer(*Player)
	}).SetPlayer(&Player{name: "pp"})
}

type PlayerContext struct {
	player *Player
}

func (self *PlayerContext) SetPlayer(p *Player) {
	self.player = p
}

type HandlerTestEchoACKWithPlayer struct {
	socket.MessageContext
	PlayerContext
}

func (self *HandlerTestEchoACKWithPlayer) Exec(context *socket.Context) {

	msg := self.Msg.(*gamedef.TestEchoACK)

	log.Debugln("server recv:", msg.String(), self.player)
	self.Ses.Send(&gamedef.TestEchoACK{
		Content: msg.String(),
	})
}

func server() {

	pipe := cellnet.NewEventPipe()

	evq := socket.NewAcceptor(pipe).Start("127.0.0.1:7201")

	socket.RegisterSessionMessage2(evq, "gamedef.TestEchoACK", new(PlayerConvertor), new(HandlerTestEchoACKWithPlayer))

	pipe.Start()

}

func client() {

	pipe := cellnet.NewEventPipe()

	evq := socket.NewConnector(pipe).Start("127.0.0.1:7201")

	socket.RegisterSessionMessage(evq, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.TestEchoACK)

		log.Debugln("client recv:", msg.String())

		signal.Done(1)
	})

	socket.RegisterSessionMessage(evq, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {

		ses.Send(&gamedef.TestEchoACK{
			Content: "hello",
		})

	})

	pipe.Start()

	signal.WaitAndExpect(1, "not recv data")

}

func TestEcho(t *testing.T) {

	signal = test.NewSignalTester(t)

	server()

	client()

}
