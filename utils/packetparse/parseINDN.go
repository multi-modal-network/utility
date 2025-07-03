package packetparse

import "fmt"

type NDNParams struct {
	EtherType  string
	NdnCode    string
	SrcNDNName string
	DstNDNName string
	NdnContent string
}

func (p *NDNParams) Parse(buffer []byte) (string, error) {
	// 检查buffer长度，至少需要36字节（根据Java代码中最大位置34+2字节）
	if len(buffer) < 36 {
		return "failed", fmt.Errorf("buffer too short, need at least 36 bytes, got %d", len(buffer))
	}

	// 提取源NDN名称 (从位置8开始，4个字节)
	srcNDNName := make([]byte, 4)
	copy(srcNDNName, buffer[8:12])

	// 提取目标NDN名称 (从位置14开始，4个字节)
	dstNDNName := make([]byte, 4)
	copy(dstNDNName, buffer[14:18])

	// 提取NDN内容 (从位置34开始，2个字节)
	ndnContent := make([]byte, 2)
	copy(ndnContent, buffer[34:36])

	// 设置固定值和解析出的字段
	p.EtherType = "8624"                         // 固定值，对应Java代码中的0x8624
	p.NdnCode = "6"                              // 固定值，对应Java代码中的6
	p.SrcNDNName = fmt.Sprintf("%x", srcNDNName) // 源NDN名称
	p.DstNDNName = fmt.Sprintf("%x", dstNDNName) // 目标NDN名称
	p.NdnContent = fmt.Sprintf("%x", ndnContent) // NDN内容

	return "success", nil
}
