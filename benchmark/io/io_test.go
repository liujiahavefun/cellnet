package benchmark

import (
	"testing"
	"time"
	"sync/atomic"
	"syscall"
	"fmt"
	"os"

	"cellnet"
	"cellnet/benchmark"
	"cellnet/example"
	"cellnet/proto/gamedef"
	"cellnet/socket"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

// 测试地址
const benchmarkAddress = "127.0.0.1:7201"

// 客户端并发数量
const clientCount = 20000

// 测试时间(秒)
const benchmarkSeconds = 20

// 多少client connected了
var(
	acceptedServerSession int64
	connectedServerSession int64
	connectedClientCount int64
)

func server() {
	queue := cellnet.NewEventQueue()
	qpsm := benchmark.NewQPSMeter(queue, func(qps int) {
		log.Infof("QPS: %d, Accepted Client: %d, Connected Client: %d", qps, acceptedServerSession, connectedServerSession)
	})

	server := socket.NewTcpServer(queue).Start(benchmarkAddress)
	socket.RegisterMessage(server, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		if qpsm.Acc() > benchmarkSeconds {
			signal.Done(1)
			log.Infof("Average QPS: %d, Accepted Client: %d, Connected Client: %d", qpsm.Average(), acceptedServerSession, connectedServerSession)
		}

		ses.Send(&gamedef.TestEchoACK{})
	})

	socket.RegisterMessage(server, "gamedef.SessionAccepted", func(content interface{}, ses cellnet.Session) {
		atomic.AddInt64(&acceptedServerSession, 1)
	})
	
	socket.RegisterMessage(server, "gamedef.SessionAcceptFailed", func(content interface{}, ses cellnet.Session) {
		msg := content.(*gamedef.SessionAcceptFailed)
		log.Infof("SessionAcceptFailed, err: %v", msg.Reason)
	})

	socket.RegisterMessage(server, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {
		atomic.AddInt64(&connectedServerSession, 1)
	})

	socket.RegisterMessage(server, "gamedef.SessionConnectFailed", func(content interface{}, ses cellnet.Session) {
		log.Infof("SessionConnectFailed")
	})

	queue.StartLoop()
}

func client() {
	queue := cellnet.NewEventQueue()
	connector := socket.NewConnector(queue)

	socket.RegisterMessage(connector, "gamedef.SessionConnected", func(content interface{}, ses cellnet.Session) {
		//log.Infoln("session connected")
		atomic.AddInt64(&connectedClientCount, 1)
		ses.Send(&gamedef.TestEchoACK{})
	})

	socket.RegisterMessage(connector, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		ses.Send(&gamedef.TestEchoACK{})
	})

	connector.Start(benchmarkAddress)

	queue.StartLoop()
}

func TestIO(t *testing.T) {
	EnableManyFiles()
	
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

func EnableManyFiles() {
	var rlim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("get rlimit error: " + err.Error())
		os.Exit(1)
	}

	rlim.Cur = 50000
	rlim.Max = 50000
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("set rlimit error: " + err.Error())
		os.Exit(1)
	}
}
