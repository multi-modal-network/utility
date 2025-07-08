package packetparse

import "fmt"

type GEOParams struct {
	EtherType     string
	GeoAreaPosLat string
	GeoAreaPosLon string
	Disa          string
	Disb          string
}

func (p *GEOParams) Parse(buffer []byte) (string, error) {
	// 检查buffer长度，至少需要48字节（根据Java代码中最大位置44+4字节）
	if len(buffer) < 48 {
		return "failed", fmt.Errorf("buffer too short, need at least 48 bytes, got %d", len(buffer))
	}

	// 提取地理区域位置纬度 (从位置40开始，4个字节)
	geoAreaPosLat := make([]byte, 4)
	copy(geoAreaPosLat, buffer[40:44])

	// 提取地理区域位置经度 (从位置44开始，4个字节)
	geoAreaPosLon := make([]byte, 4)
	copy(geoAreaPosLon, buffer[44:48])

	// 设置固定值和解析出的字段
	p.EtherType = "8947"                               // 固定值，对应Java代码中的0x8947
	p.GeoAreaPosLat = fmt.Sprintf("%x", geoAreaPosLat) // 地理区域位置纬度
	p.GeoAreaPosLon = fmt.Sprintf("%x", geoAreaPosLon) // 地理区域位置经度
	p.Disa = "00"                                      // 固定值，对应Java代码中的new byte[]{0}
	p.Disb = "00"                                      // 固定值，对应Java代码中的new byte[]{0}

	return "success", nil
}
