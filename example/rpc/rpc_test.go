package rpc

import (
	"testing"

	"cellnet"
	"cellnet/example"
	"cellnet/proto/gamedef"
	"cellnet/rpc"
	"cellnet/socket"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

func server() {
	queue := cellnet.NewEventQueue()

	server := socket.NewTcpServer(queue)
	server.SetName("server")
	server.Start("127.0.0.1:9201")

	rpc.RegisterMessage(server, "gamedef.TestEchoACK", func(content interface{}, resp rpc.Response) {
		msg := content.(*gamedef.TestEchoACK)
		log.Debugln("server recv:", msg.String())

		resp.Feedback(&gamedef.TestEchoACK{
			Content: "server recv:" + msg.Content,
		})
	})

	queue.StartLoop()
}

func client() {
	queue := cellnet.NewEventQueue()

	connector := socket.NewConnector(queue)
	connector.SetName("client")
	connector.Start("127.0.0.1:9201")

	socket.RegisterMessage(connector, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {
		rpc.Call(connector, &gamedef.TestEchoACK{
			Content: "rpc async call",
		}, func(msg *gamedef.TestEchoACK) {
			log.Debugln("client recv:", msg.Content)
			signal.Done(1)
		})
	})

	queue.StartLoop()

	signal.WaitAndExpect(1, "not recv data")
}

func TestRPC(t *testing.T) {
	signal = test.NewSignalTester(t)

	server()
	client()
}
