package packetparse

import "fmt"

type IPParams struct {
	EtherType string // 以太网类型字段，通常为0x0800表示IPv4
	SrcIPHex  string // 源IP地址的16进制字符串
	DstIPHex  string // 目标IP地址的16进制字符串
}

func (p *IPParams) Parse(buffer []byte) (string, error) {
	// 检查buffer长度是否足够
	if len(buffer) < 20 {
		return "", fmt.Errorf("buffer too short, need at least 20 bytes, got %d", len(buffer))
	}

	// 解析源IPv4地址 (位置12-15)
	srcIPv4Address := make([]byte, 4)
	for i := 0; i < 4; i++ {
		srcIPv4Address[i] = buffer[12+i]
	}

	// 解析目标IPv4地址 (位置16-19)
	dstIPv4Address := make([]byte, 4)
	for i := 0; i < 4; i++ {
		dstIPv4Address[i] = buffer[16+i]
	}

	// 转换为16进制字符串
	srcIPHex := fmt.Sprintf("%02x%02x%02x%02x",
		srcIPv4Address[0], srcIPv4Address[1], srcIPv4Address[2], srcIPv4Address[3])
	dstIPHex := fmt.Sprintf("%02x%02x%02x%02x",
		dstIPv4Address[0], dstIPv4Address[1], dstIPv4Address[2], dstIPv4Address[3])
	// 例如 srcIPHex = "c0a80001" 表示

	p.EtherType = "0800" // 设置以太网类型为IPv4
	p.SrcIPHex = srcIPHex
	p.DstIPHex = dstIPHex

	return "success", nil
}
