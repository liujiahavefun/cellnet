package echo

import (
	"testing"

	"cellnet"
	"cellnet/example"
	"cellnet/proto/gamedef"
	"cellnet/socket"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

func server() {
	queue := cellnet.NewEventQueue()

	server := socket.NewTcpServer(queue).Start("127.0.0.1:7201")

	socket.RegisterMessage(server, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.TestEchoACK)
		log.Debugln("server recv:", msg.String())
		ses.Send(&gamedef.TestEchoACK{
			Content: msg.String(),
		})
	})

	queue.StartLoop()
}

func client() {
	queue := cellnet.NewEventQueue()

	connector := socket.NewConnector(queue).Start("127.0.0.1:7201")

	socket.RegisterMessage(connector, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.TestEchoACK)
		log.Debugln("client recv:", msg.String())
		signal.Done(1)
	})

	socket.RegisterMessage(connector, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {
		ses.Send(&gamedef.TestEchoACK{
			Content: "hello",
		})
	})

	socket.RegisterMessage(connector, "gamedef.SessionConnectFailed", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.SessionConnectFailed)
		log.Debugln(msg.Reason)
	})

	queue.StartLoop()

	signal.WaitAndExpect(1, "not recv data")
}

func TestEcho(t *testing.T) {
	signal = test.NewSignalTester(t)

	server()
	client()
}
