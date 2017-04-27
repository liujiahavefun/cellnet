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
	//获取一个连接
	GetSession(int64) Session

	//遍历连接
	VisitSession(func(Session) bool)

	//连接数量
	SessionCount() int
}

type Session interface {
	//发包
	Send(interface{})

	//直接发送封包
	RawSend(*Packet)

	//断开
	Close()

	//标示ID
	GetID() int64

	//归属端
	FromPeer() Peer
}