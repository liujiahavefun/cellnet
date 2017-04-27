/*
* 本文件定义常用的接口
*/
package cellnet

/*
* 代表通信的一端
*/
type Peer interface {
	// 开启
	Start(address string) Peer

	// 关闭
	Stop()

	// 名字
	SetName(string)
	Name() string

	// Session最大包大小, 超过这个数字, 接收视为错误, 断开连接
	SetMaxPacketSize(size int)
	MaxPacketSize() int

	// 事件
	EventDispatcher

	// 连接管理
	SessionManager
}

/*
* 代表通信的一端，同时可以发起连接
*/
type Connector interface {
	// 连接后的Session
	DefaultSession() Session

	// 自动重连间隔, 0表示不重连, 默认不重连
	SetAutoReconnectSec(sec int)
}

type SessionManager interface {

	// 获取一个连接
	GetSession(int64) Session

	// 遍历连接
	VisitSession(func(Session) bool)

	// 连接数量
	SessionCount() int
}

type Session interface {

	// 发包
	Send(interface{})

	// 直接发送封包
	RawSend(*Packet)

	// 断开
	Close()

	// 标示ID
	GetID() int64

	// 归属端
	FromPeer() Peer
}

/*
TODO: 要实现以下几种回调
//start()时，无论是server启动listen还是client启动connect，都去给个回调
type onConnectFunc func(Connection) bool

//一个connection关闭时给回调，无论是server还是client
type onCloseFunc func(Connection)
//出错时给回调，时机？我倾向于server accept错误，client收发包的错误
type onErrorFunc func()

//下面这两个可以合二为一，一个是收到raw data包(byte[])，另一个是解成具体的message对象(做完反序列化之后)
type onPacketRecvFunc func(Connection, *pool.Buffer) (HandlerProc, bool)
type HandlerProc func()

type onMessageFunc func(Message, Connection)

//定时器回调
type onScheduleFunc func(time.Time, interface{})
*/