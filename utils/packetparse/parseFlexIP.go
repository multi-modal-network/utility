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

	// 合理性校验
	if srcByteLength < 0 || dstByteLength < 0 || srcByteLength > 48 || dstByteLength > 48 || len(buffer) < 4 {
		return "failed", fmt.Errorf("FlexIP buffer invalid length or address length too long")
	}

	var srcAddr, dstAddr []byte

	if srcByteLength > 0 {
		srcStartPos := 52 - srcByteLength
		if srcStartPos < 0 || srcStartPos+srcByteLength > len(buffer) {
			// Java 代码此时会抛异常，Go 这里返回空字符串
			srcAddr = []byte{}
		} else {
			srcAddr = buffer[srcStartPos : srcStartPos+srcByteLength]
		}
	}

	if dstByteLength > 0 {
		dstStartPos := 100 - dstByteLength
		if dstStartPos < 0 || dstStartPos+dstByteLength > len(buffer) {
			dstAddr = []byte{}
		} else {
			dstAddr = buffer[dstStartPos : dstStartPos+dstByteLength]
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
