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
	HeaderLenSize    = 4 // Len 字段长度
	HeaderIsRespSize = 1 // IsResp 字段长度
	HeaderModuleSize = 4 // Module 字段长度
	HeaderCmdSize    = 4 // Cmd 字段长度
	HeaderCodeSize   = 4 // Code 字段长度（仅响应消息）

	// 请求消息头长度：4 + 1 + 4 + 4 = 13
	ReqHeaderSize = HeaderLenSize + HeaderIsRespSize + HeaderModuleSize + HeaderCmdSize

	// 响应消息头长度：4 + 1 + 4 + 4 + 4 = 17
	RespHeaderSize = HeaderLenSize + HeaderIsRespSize + HeaderModuleSize + HeaderCmdSize + HeaderCodeSize
)

// ReqMsg 请求消息
type ReqMsg struct {
	Module  uint32 // 模块号
	Cmd     uint32 // 命令号
	Payload []byte // 消息体（PB 序列化）
}

// RespMsg 响应消息
type RespMsg struct {
	Module  uint32 // 模块号
	Cmd     uint32 // 命令号
	Code    uint32 // 错误码（0 为正常）
	Payload []byte // 消息体（PB 序列化）
}

// EncodeReqMsg 编码请求消息
func EncodeReqMsg(msg *ReqMsg) []byte {
	payloadLen := len(msg.Payload)
	totalLen := ReqHeaderSize + payloadLen

	buf := make([]byte, totalLen)

	// Len (4字节) - 消息体长度（不包含 Len 字段本身）
	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen-4))

	// IsResp (1字节) - 0 表示请求
	buf[4] = MsgTypeRequest

	// Module (4字节)
	binary.BigEndian.PutUint32(buf[5:9], msg.Module)

	// Cmd (4字节)
	binary.BigEndian.PutUint32(buf[9:13], msg.Cmd)

	// Payload (N字节)
	copy(buf[13:], msg.Payload)

	return buf
}

// EncodeRespMsg 编码响应消息
func EncodeRespMsg(msg *RespMsg) []byte {
	payloadLen := len(msg.Payload)
	totalLen := RespHeaderSize + payloadLen

	buf := make([]byte, totalLen)

	// Len (4字节) - 消息体长度（不包含 Len 字段本身）
	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen-4))

	// IsResp (1字节) - 1 表示响应
	buf[4] = MsgTypeResponse

	// Module (4字节)
	binary.BigEndian.PutUint32(buf[5:9], msg.Module)

	// Cmd (4字节)
	binary.BigEndian.PutUint32(buf[9:13], msg.Cmd)

	// Code (4字节) - 错误码
	binary.BigEndian.PutUint32(buf[13:17], msg.Code)

	// Payload (N字节)
	copy(buf[17:], msg.Payload)

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
		return decodeReqMsg(reader, msgLen)
	case MsgTypeResponse:
		return decodeRespMsg(reader, msgLen)
	default:
		return nil, fmt.Errorf("未知消息类型: %d", isResp)
	}
}

// decodeReqMsg 解码请求消息
func decodeReqMsg(reader io.Reader, msgLen uint32) (*ReqMsg, error) {
	// 读取剩余部分：Module(4) + Cmd(4) + Payload(N)
	remainLen := msgLen - 1 // 减去已读的 IsResp
	buf := make([]byte, remainLen)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("读取请求消息失败: %w", err)
	}

	msg := &ReqMsg{
		Module:  binary.BigEndian.Uint32(buf[0:4]),
		Cmd:     binary.BigEndian.Uint32(buf[4:8]),
		Payload: buf[8:],
	}

	return msg, nil
}

// decodeRespMsg 解码响应消息
func decodeRespMsg(reader io.Reader, msgLen uint32) (*RespMsg, error) {
	// 读取剩余部分：Module(4) + Cmd(4) + Code(4) + Payload(N)
	remainLen := msgLen - 1 // 减去已读的 IsResp
	buf := make([]byte, remainLen)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("读取响应消息失败: %w", err)
	}

	msg := &RespMsg{
		Module:  binary.BigEndian.Uint32(buf[0:4]),
		Cmd:     binary.BigEndian.Uint32(buf[4:8]),
		Code:    binary.BigEndian.Uint32(buf[8:12]),
		Payload: buf[12:],
	}

	return msg, nil
}
