package packetparse

import "fmt"

type IDParams struct {
	EtherType   string
	SrcIdentity string
	DstIdentity string
}

func (p *IDParams) Parse(buffer []byte) (string, error) {
	// 检查buffer长度，至少需要8字节（4字节srcIdentity + 4字节dstIdentity）
	if len(buffer) < 8 {
		return "failed", fmt.Errorf("buffer too short, need at least 8 bytes, got %d", len(buffer))
	}

	// 提取源Identity (从位置0开始，4个字节)
	srcIdentity := make([]byte, 4)
	copy(srcIdentity, buffer[0:4])

	// 提取目标Identity (从位置4开始，4个字节)
	dstIdentity := make([]byte, 4)
	copy(dstIdentity, buffer[4:8])

	// 设置固定值和解析出的字段
	p.EtherType = "0812"                           // 固定值，对应Java代码中的0x0812
	p.SrcIdentity = fmt.Sprintf("%x", srcIdentity) // 源Identity
	p.DstIdentity = fmt.Sprintf("%x", dstIdentity) // 目标Identity

	return "success", nil
}
