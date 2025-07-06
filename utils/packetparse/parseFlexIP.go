package packetparse

import "fmt"

// ModalParser 接口
type ModalParser interface {
	Parse(buffer []byte) (string, error)
}

type FLEXIPParams struct {
	EtherType string
	SrcFormat string
	DstFormat string
	SrcAddr   string
	DstAddr   string
}

// Parse 解析 buffer，填充 FLEXIPParams 字段（全部为十六进制字符串）
func (p *FLEXIPParams) Parse(buffer []byte) (string, error) {
	if len(buffer) < 4 {
		return "failed", fmt.Errorf("FlexIP buffer too short, need at least 4 bytes for prefix, got %d", len(buffer))
	}

	// 完全按照Java代码的方式解析
	flexipPrefix := int((buffer[0]&0xff)<<24 | (buffer[1]&0xff)<<16 | (buffer[2]&0xff)<<8 | (buffer[3] & 0xff))

	// 完全按照Java代码的位操作
	srcFormat := (flexipPrefix >> 26) & 0x3
	dstFormat := (flexipPrefix >> 24) & 0x3
	srcLength := (flexipPrefix >> 12) & 0x7ff
	dstLength := flexipPrefix & 0x7ff

	// 添加调试信息
	fmt.Printf("Debug: flexipPrefix=0x%08x, srcFormat=%d, dstFormat=%d, srcLength=%d, dstLength=%d\n",
		flexipPrefix, srcFormat, dstFormat, srcLength, dstLength)

	// 计算字节长度
	srcByteLength := srcLength / 8
	dstByteLength := dstLength / 8

	// 验证长度的合理性
	if srcByteLength > 48 || dstByteLength > 48 {
		return "failed", fmt.Errorf("FlexIP address length too long: src=%d, dst=%d bytes", srcByteLength, dstByteLength)
	}

	// 验证buffer长度
	if len(buffer) < 100 {
		return "failed", fmt.Errorf("FlexIP buffer too short, need at least 100 bytes, got %d", len(buffer))
	}

	var srcAddr, dstAddr []byte

	// 按照Java代码的方式提取地址
	if srcByteLength > 0 {
		srcAddr = make([]byte, srcByteLength)
		srcStartPos := 52 - srcByteLength
		for i := 0; i < srcByteLength; i++ {
			srcAddr[i] = buffer[srcStartPos+i]
		}
	}

	if dstByteLength > 0 {
		dstAddr = make([]byte, dstByteLength)
		dstStartPos := 100 - dstByteLength
		for i := 0; i < dstByteLength; i++ {
			dstAddr[i] = buffer[dstStartPos+i]
		}
	}

	// 填充结构体字段
	p.EtherType = fmt.Sprintf("%04x", 0x3690)
	p.SrcFormat = fmt.Sprintf("%x", srcFormat)
	p.DstFormat = fmt.Sprintf("%x", dstFormat)

	if len(srcAddr) > 0 {
		p.SrcAddr = fmt.Sprintf("%x", srcAddr)
	} else {
		p.SrcAddr = ""
	}

	if len(dstAddr) > 0 {
		p.DstAddr = fmt.Sprintf("%x", dstAddr)
	} else {
		p.DstAddr = ""
	}

	return "success", nil
}
