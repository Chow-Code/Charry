package tcp

import (
	"net"
	"time"
)

// 心跳相关常量
const (
	HeartbeatModule uint32 = 0 // 心跳模块号
	HeartbeatCmd    uint32 = 1 // 心跳命令号
	HeartbeatCode   uint32 = 0 // 心跳响应码
)

// HeartbeatInterval 心跳发送间隔（10秒）
var HeartbeatInterval = 10 * time.Second

// SendHeartbeat 发送心跳消息
func SendHeartbeat(conn net.Conn) error {
	req := &ClusterReqMsg{
		Module:    HeartbeatModule,
		Cmd:       HeartbeatCmd,
		SessionId: "heartbeat", // 心跳固定 sessionId
		Payload:   []byte{},    // 空 payload
	}

	data := EncodeClusterReqMsg(req)
	_, err := conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// IsHeartbeatMsg 判断是否为心跳消息
func IsHeartbeatMsg(module, cmd uint32) bool {
	return module == HeartbeatModule && cmd == HeartbeatCmd
}

// HandleHeartbeatReq 处理心跳请求
func HandleHeartbeatReq(conn net.Conn, req *ClusterReqMsg) error {
	// 响应心跳
	resp := &ClusterRespMsg{
		Module:    req.Module,
		Cmd:       req.Cmd,
		SessionId: req.SessionId,
		Code:      HeartbeatCode,
		Payload:   []byte{},
	}

	data := EncodeClusterRespMsg(resp)
	_, err := conn.Write(data)
	return err
}
