package calc

import (
	"fmt"
	"math"
	"onosutil/model"
)

var portUp int32 = 1
var portLeft int32 = 2
var portRight int32 = 3

// 烽火Tofino交换机端口设置
var domain2TofinoPorts = []int32{132, 140, 148, 164}
var domain4TofinoPorts = []int32{132, 140, 164}
var domain6TofinoPorts = []int32{132, 140, 148, 164}

// 武大Tofino交换机端口设置
//var domain2TofinoPorts = []int32{128,144,160,176}
//var domain4TofinoPorts = []int32{128,144,176}
//var domain6TofinoPorts = []int32{128,144,160,176}

var domain2TofinoSwitch int32 = 2000
var domain4TofinoSwitch int32 = 4000
var domain6TofinoSwitch int32 = 6000
var domain3SatelliteSwitch1 int32 = 3100
var domain3SatelliteSwitch2 int32 = 3200
var domain3SatelliteSwitch3 int32 = 3300

func GetDomain(vmx int32) int32 {
	if vmx >= 0 && vmx <= 2 {
		return 1
	} else if vmx >= 3 && vmx <= 4 {
		return 5
	}
	return 7
}

func GetGroup(vmx int32) int32 {
	if vmx == 0 || vmx == 3 || vmx == 5 {
		return 1
	} else if vmx == 1 || vmx == 4 || vmx == 6 {
		return 2
	}
	return 3
}

func GetLevel(switchID int32) int32 {
	return int32(math.Floor(math.Log2(float64(switchID))) + 1)
}

func GetSwitchID(deviceID string) int32 {
	if deviceID == "device:domain2:p1" {
		return domain2TofinoSwitch
	} else if deviceID == "device:domain4:p4" {
		return domain4TofinoSwitch
	} else if deviceID == "device:domain6:p6" {
		return domain6TofinoSwitch
	} else if deviceID == "device:satellite1" {
		return domain3SatelliteSwitch1
	} else if deviceID == "device:satellite2" {
		return domain3SatelliteSwitch2
	} else {
		return domain3SatelliteSwitch3
	}
}

// addUplinkDevices 上行链路 srcSwitch向父节点转发
func addUplinkDevices(vmx, srcSwitch int32, crossGroup bool) []model.DevicePort {
	res := make([]model.DevicePort, 0)
	domain, group := GetDomain(vmx), GetGroup(vmx)
	end := map[bool]int32{false: 1, true: 0}[crossGroup] // 一行赋值
	for s := srcSwitch; s != end; s /= 2 {
		res = append(res, model.DevicePort{
			DeviceName: fmt.Sprintf("device:domain%d:group%d:level%d:s%d", domain, group, GetLevel(s), s+255*vmx),
			Port:       portUp,
		})
	}
	return res
}

// addDownlinkDevices 下行链路 dstSwitch的父节点向dstSwitch转发
func addDownlinkDevices(vmx, dstSwitch int32) []model.DevicePort {
	domain, group := GetDomain(vmx), GetGroup(vmx)
	res := make([]model.DevicePort, 0)
	for t := dstSwitch; t != 1; t /= 2 {
		if t/2*2 == t {
			res = append(res, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain%d:group%d:level%d:s%d", domain, group, GetLevel(t/2), t/2+255*vmx),
				Port:       portLeft,
			})
		} else {
			res = append(res, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain%d:group%d:level%d:s%d", domain, group, GetLevel(t/2), t/2+255*vmx),
				Port:       portRight,
			})
		}
	}
	// 反转切片
	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}
	// 最后dstSwitch向dstHost转发
	res = append(res, model.DevicePort{
		DeviceName: fmt.Sprintf("device:domain%d:group%d:level%d:s%d", domain, group, GetLevel(dstSwitch), dstSwitch+255*vmx),
		Port:       portLeft,
	})
	return res
}

// GetPathDevices 算路
func GetPathDevices(srcHost, dstHost int32) []model.DevicePort {
	srcVmx, dstVmx := srcHost/256, dstHost/256
	srcDomain, dstDomain := GetDomain(srcVmx), GetDomain(dstVmx)
	srcSwitch, dstSwitch := (srcHost-1)%255+1, (dstHost-1)%255+1
	devices := make([]model.DevicePort, 0)
	if srcVmx == dstVmx { // 容器内通信
		devices = append(devices, addUplinkDevices(srcVmx, srcSwitch, false)...) // 处理上行链路
		devices = append(devices, addDownlinkDevices(dstVmx, dstSwitch)...)      // 处理下行链路
	} else if srcDomain == dstDomain { // 跨容器通信
		devices = append(devices, addUplinkDevices(srcVmx, srcSwitch, true)...) // 处理上行链路
		// tofino交换机下发流表
		switch srcDomain {
		case 1:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain2:p1"),
				Port:       domain2TofinoPorts[dstVmx%3],
			})
			break
		case 5:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain4:p4"),
				Port:       domain4TofinoPorts[dstVmx%3],
			})
			break
		case 7:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain6:p6"),
				Port:       domain6TofinoPorts[(dstVmx+1)%3],
			})
			break
		}
		// 目的groupS1直接发至目的主机
		devices = append(devices, addDownlinkDevices(dstVmx, dstSwitch)...) // 处理下行链路
	} else { // 跨域通信
		devices = append(devices, addUplinkDevices(srcVmx, srcSwitch, true)...) // 处理上行链路
		// tofino交换机下发流表 （首先查询Tofino交换机模态对应转发端口）
		switch srcDomain {
		case 1:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain2:p1"),
				Port:       0,
			})
			if dstDomain == 5 {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain4:p4"),
					Port:       domain4TofinoPorts[dstVmx%3],
				})
			} else {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain6:p6"),
					Port:       domain6TofinoPorts[(dstVmx+1)%3],
				})
			}
			break
		case 5:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain4:p4"),
				Port:       0,
			})
			if dstDomain == 1 {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain2:p1"),
					Port:       domain2TofinoPorts[dstVmx%3],
				})
			} else {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain6:p6"),
					Port:       domain6TofinoPorts[(dstVmx+1)%3],
				})
			}
			break
		case 7:
			devices = append(devices, model.DevicePort{
				DeviceName: fmt.Sprintf("device:domain6:p6"),
				Port:       0,
			})
			if dstDomain == 1 {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain2:p1"),
					Port:       domain2TofinoPorts[dstVmx%3],
				})
			} else {
				devices = append(devices, model.DevicePort{
					DeviceName: fmt.Sprintf("device:domain4:p4"),
					Port:       domain4TofinoPorts[dstVmx%3],
				})
			}
		}
		devices = append(devices, addDownlinkDevices(dstVmx, dstSwitch)...) // 处理下行链路
	}
	return devices
}
