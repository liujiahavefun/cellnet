// Generated by github.com/davyxu/cellnet/protoc-gen-msg
// DO NOT EDIT!
// Source: core.proto

package gamedef

import (
	"cellnet"
)

func init() {
	// core.proto
	cellnet.RegisterMessageMeta("gamedef.SessionAccepted", (*SessionAccepted)(nil), 4151444293)
	cellnet.RegisterMessageMeta("gamedef.SessionConnected", (*SessionConnected)(nil), 459632229)
	cellnet.RegisterMessageMeta("gamedef.SessionAcceptFailed", (*SessionAcceptFailed)(nil), 399042283)
	cellnet.RegisterMessageMeta("gamedef.SessionConnectFailed", (*SessionConnectFailed)(nil), 1644962508)
	cellnet.RegisterMessageMeta("gamedef.SessionClosed", (*SessionClosed)(nil), 1412646790)
	cellnet.RegisterMessageMeta("gamedef.RemoteCallREQ", (*RemoteCallREQ)(nil), 1469566342)
	cellnet.RegisterMessageMeta("gamedef.RemoteCallACK", (*RemoteCallACK)(nil), 1020080612)
	cellnet.RegisterMessageMeta("gamedef.TestEchoACK", (*TestEchoACK)(nil), 1899977859)
}
