package packetparse

import "fmt"

type MFParams struct {
	EtherType string
    SrcGuid   string
    DstGuid   string
}

func (p *MFParams) Parse(buffer []byte) (string, error) {
// 检查buffer长度，至少需要12字节（4字节srcGuid + 4字节dstGuid + 其他头部）
    if len(buffer) < 12 {
        return "failed", fmt.Errorf("buffer too short, need at least 12 bytes, got %d", len(buffer))
    }
    
    // 提取源GUID (从位置4开始，4个字节)
    srcGuid := make([]byte, 4)
    copy(srcGuid, buffer[4:8])
    
    // 提取目标GUID (从位置8开始，4个字节)
    dstGuid := make([]byte, 4)
    copy(dstGuid, buffer[8:12])
    
	p.EtherType = "27c0" // 固定值，对应Java代码中的0x27c0
	p.SrcGuid = fmt.Sprintf("%x", srcGuid)
	p.DstGuid = fmt.Sprintf("%x", dstGuid)
    
    return "success", nil
}
