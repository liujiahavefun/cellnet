package main

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
	"cellnet/proto/session"
	"cellnet/socket"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

// 测试地址
const benchmarkAddress = "127.0.0.1:7201"

// 客户端并发数量
const clientCount = 10000

// 测试时间(秒)
const benchmarkSeconds = 2000

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

	socket.RegisterMessage(server, "session.SessionAccepted", func(content interface{}, ses cellnet.Session) {
		atomic.AddInt64(&acceptedServerSession, 1)
	})

	socket.RegisterMessage(server, "session.SessionAcceptFailed", func(content interface{}, ses cellnet.Session) {
		msg := content.(*session.SessionAcceptFailed)
		log.Infof("SessionAcceptFailed, err: %v", msg.Reason)
	})

	socket.RegisterMessage(server, "session.SessionConnected", func(content interface{}, ses cellnet.Session) {
		atomic.AddInt64(&connectedServerSession, 1)
	})

	socket.RegisterMessage(server, "session.SessionConnectFailed", func(content interface{}, ses cellnet.Session) {
		log.Infof("SessionConnectFailed")
	})

	socket.RegisterMessage(server, "session.SessionError", func(content interface{}, ses cellnet.Session) {
		msg := content.(session.SessionError)
		log.Infof("SessionError: ", msg.Reason)
	})

	queue.StartLoop()
}

func client() {
	queue := cellnet.NewEventQueue()
	connector := socket.NewConnector(queue)

	socket.RegisterMessage(connector, "session.SessionConnected", func(content interface{}, ses cellnet.Session) {
		//log.Infoln("session connected")
		atomic.AddInt64(&connectedClientCount, 1)
		ses.Send(&gamedef.TestEchoACK{})
	})

	socket.RegisterMessage(connector, "session.SessionError", func(content interface{}, ses cellnet.Session) {
		msg := content.(*session.SessionError)
		log.Infoln("session error:", msg.Reason)
	})

	socket.RegisterMessage(connector, "gamedef.TestEchoACK", func(content interface{}, ses cellnet.Session) {
		//ses.Send(&gamedef.TestEchoACK{})
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
		time.Sleep(50*time.Millisecond)
		go client()
	}

	log.Infoln("all clients started")

	signal.WaitAndExpect(1, "recv time out")
}

func EnableManyFiles() {
	var rlim syscall.Rlimit
	rlim.Cur = 50000
	rlim.Max = 50000

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("set rlimit error: " + err.Error())
		os.Exit(1)
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println("get rlimit error: " + err.Error())
		os.Exit(1)
	}

	fmt.Println("rlim.Curr", rlim.Cur)
	fmt.Println("rlim.Max", rlim.Max)
}