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

	flexipPrefix := int(buffer[0])<<24 | int(buffer[1])<<16 | int(buffer[2])<<8 | int(buffer[3])
	srcFormat := (flexipPrefix >> 26) & 0x3
	dstFormat := (flexipPrefix >> 24) & 0x3
	srcLength := (flexipPrefix >> 12) & 0x7ff
	dstLength := flexipPrefix & 0x7ff

	srcByteLength := srcLength / 8
	dstByteLength := dstLength / 8

	if len(buffer) < 100 || srcByteLength > 48 || dstByteLength > 48 {
		return "failed", fmt.Errorf("FlexIP buffer invalid length or address length too long")
	}

	// 解析源FlexIP地址
	srcAddr := make([]byte, srcByteLength)
	srcStartPos := 52 - srcByteLength
	for i := 0; i < srcByteLength; i++ {
		srcAddr[i] = buffer[srcStartPos+i]
	}

	// 解析目标FlexIP地址
	dstAddr := make([]byte, dstByteLength)
	dstStartPos := 100 - dstByteLength
	for i := 0; i < dstByteLength; i++ {
		dstAddr[i] = buffer[dstStartPos+i]
	}

	// 填充结构体字段（全部转为十六进制字符串）
	p.EtherType = fmt.Sprintf("%x", 0x3690)
	p.SrcFormat = fmt.Sprintf("%x", srcFormat)
	p.DstFormat = fmt.Sprintf("%x", dstFormat)
	p.SrcAddr = fmt.Sprintf("%x", srcAddr)
	p.DstAddr = fmt.Sprintf("%x", dstAddr)

	return "success", nil
}
