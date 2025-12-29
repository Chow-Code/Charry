package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
)

// 消息类型
const (
	MsgTypeRequest  byte = 0 // 请求消息
	MsgTypeResponse byte = 1 // 响应消息
)

// 消息头长度
const (
	HeaderLenSize       = 4  // Len 字段长度
	HeaderIsRespSize    = 1  // IsResp 字段长度
	HeaderModuleSize    = 4  // Module 字段长度
	HeaderCmdSize       = 4  // Cmd 字段长度
	HeaderSessionIdSize = 36 // SessionId 字段长度（UUID）
	HeaderCodeSize      = 4  // Code 字段长度（仅响应消息）

	// 请求消息头长度：4 + 1 + 4 + 4 + 36 = 49
	ClusterReqHeaderSize = HeaderLenSize + HeaderIsRespSize + HeaderModuleSize + HeaderCmdSize + HeaderSessionIdSize

	// 响应消息头长度：4 + 1 + 4 + 4 + 36 + 4 = 53
	ClusterRespHeaderSize = HeaderLenSize + HeaderIsRespSize + HeaderModuleSize + HeaderCmdSize + HeaderSessionIdSize + HeaderCodeSize
)

// ClusterReqMsg 集群请求消息
type ClusterReqMsg struct {
	Module    uint32 // 模块号
	Cmd       uint32 // 命令号
	SessionId string // 会话ID（UUID，36字节）
	Payload   []byte // 消息体（PB 序列化）
}

// ClusterRespMsg 集群响应消息
type ClusterRespMsg struct {
	Module    uint32 // 模块号
	Cmd       uint32 // 命令号
	SessionId string // 会话ID（UUID，36字节）
	Code      uint32 // 错误码（0 为正常）
	Payload   []byte // 消息体（PB 序列化）
}

// EncodeClusterReqMsg 编码请求消息
func EncodeClusterReqMsg(msg *ClusterReqMsg) []byte {
	payloadLen := len(msg.Payload)
	totalLen := ClusterReqHeaderSize + payloadLen

	buf := make([]byte, totalLen)

	// Len (4字节) - 消息体长度（不包含 Len 字段本身）
	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen-4))

	// IsResp (1字节) - 0 表示请求
	buf[4] = MsgTypeRequest

	// Module (4字节)
	binary.BigEndian.PutUint32(buf[5:9], msg.Module)

	// Cmd (4字节)
	binary.BigEndian.PutUint32(buf[9:13], msg.Cmd)

	// SessionId (36字节) - UUID
	copy(buf[13:49], []byte(padSessionId(msg.SessionId)))

	// Payload (N字节)
	copy(buf[49:], msg.Payload)

	return buf
}

// EncodeClusterRespMsg 编码响应消息
func EncodeClusterRespMsg(msg *ClusterRespMsg) []byte {
	payloadLen := len(msg.Payload)
	totalLen := ClusterRespHeaderSize + payloadLen

	buf := make([]byte, totalLen)

	// Len (4字节) - 消息体长度（不包含 Len 字段本身）
	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen-4))

	// IsResp (1字节) - 1 表示响应
	buf[4] = MsgTypeResponse

	// Module (4字节)
	binary.BigEndian.PutUint32(buf[5:9], msg.Module)

	// Cmd (4字节)
	binary.BigEndian.PutUint32(buf[9:13], msg.Cmd)

	// SessionId (36字节) - UUID
	copy(buf[13:49], []byte(padSessionId(msg.SessionId)))

	// Code (4字节) - 错误码
	binary.BigEndian.PutUint32(buf[49:53], msg.Code)

	// Payload (N字节)
	copy(buf[53:], msg.Payload)

	return buf
}

// DecodeMsg 解码消息（自动判断请求或响应）
func DecodeMsg(reader io.Reader) (interface{}, error) {
	// 1. 读取 Len (4字节)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(reader, lenBuf); err != nil {
		return nil, fmt.Errorf("读取长度失败: %w", err)
	}
	msgLen := binary.BigEndian.Uint32(lenBuf)

	// 2. 读取 IsResp (1字节)
	isRespBuf := make([]byte, 1)
	if _, err := io.ReadFull(reader, isRespBuf); err != nil {
		return nil, fmt.Errorf("读取消息类型失败: %w", err)
	}
	isResp := isRespBuf[0]

	// 3. 根据类型解码
	switch isResp {
	case MsgTypeRequest:
		return decodeClusterReqMsg(reader, msgLen)
	case MsgTypeResponse:
		return decodeClusterRespMsg(reader, msgLen)
	default:
		return nil, fmt.Errorf("未知消息类型: %d", isResp)
	}
}

// padSessionId 填充 SessionId 到 36 字节
func padSessionId(sessionId string) string {
	if len(sessionId) >= 36 {
		return sessionId[:36]
	}
	// 不足36字节，右侧填充空格
	return sessionId + string(make([]byte, 36-len(sessionId)))
}

// trimSessionId 去除 SessionId 的填充
func trimSessionId(sessionId string) string {
	// 去除右侧空格和 null 字符
	for i := len(sessionId) - 1; i >= 0; i-- {
		if sessionId[i] != ' ' && sessionId[i] != 0 {
			return sessionId[:i+1]
		}
	}
	return ""
}

// decodeClusterReqMsg 解码请求消息
func decodeClusterReqMsg(reader io.Reader, msgLen uint32) (*ClusterReqMsg, error) {
	// 读取剩余部分：Module(4) + Cmd(4) + SessionId(36) + Payload(N)
	remainLen := msgLen - 1 // 减去已读的 IsResp
	buf := make([]byte, remainLen)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("读取请求消息失败: %w", err)
	}

	msg := &ClusterReqMsg{
		Module:    binary.BigEndian.Uint32(buf[0:4]),
		Cmd:       binary.BigEndian.Uint32(buf[4:8]),
		SessionId: trimSessionId(string(buf[8:44])),
		Payload:   buf[44:],
	}

	return msg, nil
}

// decodeClusterRespMsg 解码响应消息
func decodeClusterRespMsg(reader io.Reader, msgLen uint32) (*ClusterRespMsg, error) {
	// 读取剩余部分：Module(4) + Cmd(4) + SessionId(36) + Code(4) + Payload(N)
	remainLen := msgLen - 1 // 减去已读的 IsResp
	buf := make([]byte, remainLen)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("读取响应消息失败: %w", err)
	}

	msg := &ClusterRespMsg{
		Module:    binary.BigEndian.Uint32(buf[0:4]),
		Cmd:       binary.BigEndian.Uint32(buf[4:8]),
		SessionId: trimSessionId(string(buf[8:44])),
		Code:      binary.BigEndian.Uint32(buf[44:48]),
		Payload:   buf[48:],
	}

	return msg, nil
}
