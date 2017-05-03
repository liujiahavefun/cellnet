package rpc

import (
	"errors"
	"reflect"
	"sync"

	"cellnet"
	"cellnet/proto/gamedef"
	"cellnet/socket"
)

//TODO: 这里用map保存一个递增的
var (
	reqByID  = make(map[int64]*request)
	reqGuard sync.RWMutex
	idacc    int64
)

var (
	ErrorInvalidPeerSession   error = errors.New("rpc: invalid peer type, require connector")
	ErrorConnectorSessionNotReady error = errors.New("rpc: connector session not ready")
)

//添加一个rpc的调用信息
func addCall(req *request) {
	reqGuard.Lock()
	defer reqGuard.Unlock()

	idacc++
	req.id = idacc

	//TODO 底层支持timer, 抛出一个超时检测, 清理map
	reqByID[req.id] = req
}

//获取一个rpc调用信息
func getCall(id int64) *request {
	reqGuard.RLock()
	defer reqGuard.RUnlock()

	if v, ok := reqByID[id]; ok {
		return v
	}

	return nil
}

func removeCall(id int64) {
	reqGuard.Lock()
	defer reqGuard.Unlock()

	delete(reqByID, id)
}

//从peer获取rpc使用的session
func getPeerSession(p interface{}) (cellnet.Session, cellnet.EventDispatcher, error) {
	var ses cellnet.Session
	switch p.(type) {
		case cellnet.Peer:
			//liujia: interface {Session() cellnet.Session}判断p这个interface是否实现了"Session() cellnet.Session"这个函数
			//通常如果p的实现了cellnet.Peer，那么p是Connector
			if connPeer, ok := p.(interface {Session() cellnet.Session}); ok {
				ses = connPeer.Session()
			} else {
				return nil, nil, ErrorInvalidPeerSession
			}
		case cellnet.Session:
			ses = p.(cellnet.Session)
	}

	if ses == nil {
		return nil, nil, ErrorConnectorSessionNotReady
	}

	//liujia: FromPeer()返回的是myself，对于server_session就是Server，对于client_session就是Connector，而这俩都是EventDispatcher
	return ses, ses.FromPeer().(cellnet.EventDispatcher), nil
}

//传入peer或者session
func Call(p interface{}, args interface{}, callback interface{}) {
	ses, evd, err := getPeerSession(p)
	if err != nil {
		log.Errorln(err)
		return
	}

	_, msg := newRequest(evd, args, callback)
	ses.Send(msg)

	// TODO rpc日志
}

// 传入peer或者session
func CallSync(p interface{}, args interface{}, callback interface{}) {
	ses, evq, err := getPeerSession(p)
	if err != nil {
		log.Errorln(err)
		return
	}

	req, msg := newRequest(evq, args, callback)
	req.recvied = make(chan bool)

	ses.Send(msg)
	<-req.recvied

	// TODO rpc日志
}

type request struct {
	id        int64
	callback  reflect.Value
	replyType reflect.Type
	recvied   chan bool
}

func (self *request) done(msg *gamedef.RemoteCallACK) {
	rawType, err := cellnet.ParsePacket(&cellnet.Packet{
		MsgID: msg.MsgID,
		Data:  msg.Data,
	}, self.replyType)

	defer removeCall(self.id)

	if err != nil {
		log.Errorln(err)
		return
	}

	//这里的反射, 会影响非常少的效率, 但因为外部写法简单, 就算了
	self.callback.Call([]reflect.Value{reflect.ValueOf(rawType)})

	if self.recvied != nil {
		self.recvied <- true
	}
}

//var needRegisterClient bool = true
//var needRegisterClientGuard sync.Mutex
var once sync.Once

func newRequest(evd cellnet.EventDispatcher, args interface{}, callback interface{}) (*request, interface{}) {
	//第一次请求时注册消息
	/*
	needRegisterClientGuard.Lock()
	if needRegisterClient {
		socket.RegisterMessage(evd, "gamedef.RemoteCallACK", func(content interface{}, ses cellnet.Session) {
			msg := content.(*gamedef.RemoteCallACK)
			c := getCall(msg.CallID)
			if c == nil {
				return
			}

			c.done(msg)
		})

		needRegisterClient = false
	}
	needRegisterClientGuard.Unlock()
	*/

	once.Do(func() {
		socket.RegisterMessage(evd, "gamedef.RemoteCallACK", func(content interface{}, ses cellnet.Session) {
			msg := content.(*gamedef.RemoteCallACK)
			c := getCall(msg.CallID)
			if c == nil {
				return
			}

			c.done(msg)
		})
	})

	req := &request{}

	//liujia:replyType是callback函数的第一个参数的类型
	//如果rpc的对端返回的结果的类型与这个不一致，则会在ParsePacket()时报错
	//大概这个样子：proto: bad wiretype for field gamedef.RemoteCallACK.MsgID: got wiretype 2, want 0
	funcType := reflect.TypeOf(callback)
	req.replyType = funcType.In(0)
	req.callback = reflect.ValueOf(callback)

	pkt, _ := cellnet.BuildPacket(args)
	addCall(req)

	return req, &gamedef.RemoteCallREQ{
		MsgID:  pkt.MsgID,
		Data:   pkt.Data,
		CallID: req.id,
	}
}
