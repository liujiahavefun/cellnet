package benchmark

import (
	"testing"
	"time"

	"cellnet"
	"cellnet/benchmark"
	"cellnet/example"
	"cellnet/proto/gamedef"
	"cellnet/socket"
	"github.com/davyxu/golog"
	"sync"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

// 测试地址
const benchmarkAddress = "127.0.0.1:7201"

// 客户端并发数量
const clientCount = 10000

// 测试时间(秒)
const benchmarkSeconds = 20
cd
func server() {
	queue := cellnet.NewEventQueue()
	qpsm := benchmark.NewQPSMeter(queue, func(qps int) {
		log.Infof("QPS: %d", qps)
	})

	evd := socket.NewTcpServer(queue).Start(benchmarkAddress)
	socket.RegisterMessage(evd, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		if qpsm.Acc() > benchmarkSeconds {
			signal.Done(1)
			log.Infof("Average QPS: %d", qpsm.Average())
		}

		ses.Send(&gamedef.TestEchoACK{})
	})

	queue.StartLoop()
}

func client() {
	queue := cellnet.NewEventQueue()
	evd := socket.NewConnector(queue).Start(benchmarkAddress)

	socket.RegisterMessage(evd, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {
		//log.Infoln("session connected")
		ses.Send(&gamedef.TestEchoACK{})
	})

	socket.RegisterMessage(evd, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		ses.Send(&gamedef.TestEchoACK{})
	})

	queue.StartLoop()
}

func TestIO(t *testing.T) {
	// 屏蔽socket层的调试日志
	golog.SetLevelByString("socket", "error")

	signal = test.NewSignalTester(t)

	// 超时时间为测试时间延迟一会
	signal.SetTimeout((benchmarkSeconds + 5) * time.Second)

	log.Infoln("start server")

	server()

	log.Infoln("start all clients")
	for i := 0; i < clientCount; i++ {
		go client()
	}

	log.Infoln("all clients started")

	signal.WaitAndExpect(1, "recv time out")
}
