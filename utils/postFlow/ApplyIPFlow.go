package postflow

import (
	"fmt"
	"onosutil/utils/packetparse"
	"strconv"
	"strings"
)

func portToHex(port string) (string, error) {
	// 将字符串转换为整数
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid port number: %v", err)
	}

	// 检查端口号是否在有效范围内
	if portNum < 0 || portNum > 65535 {
		return "", fmt.Errorf("port number out of range (0-65535)")
	}

	// 将整数转换为十六进制字符串，去掉前缀"0x"
	hexStr := fmt.Sprintf("%X", portNum)

	// 如果十六进制字符串长度为1，前面补0（可选，根据需求）
	if len(hexStr) == 1 {
		hexStr = "0" + hexStr
	}

	return hexStr, nil
}

func ApplyIPFlow(flows []string, url string, params packetparse.IPParams) (string, error) {
    if len(flows) == 0 {
        return "no flow found", nil
    }
    // 构建流表
    res := []Flow{}
    for _, flow := range flows {
        // 从flow解析deviceID和端口
        flowInfo := strings.Split(flow, "/")
        if len(flowInfo) < 2 {
            return "invalid flow format", fmt.Errorf("invalid flow format: %s", flow)
        }
        deviceID, port := flowInfo[0], flowInfo[1]
        // 把port修改为对应的十六进制字符串
        if strings.Contains(deviceID,"domain2")||strings.Contains(deviceID,"domain4")||strings.Contains(deviceID,"domain6"){
            port,_ = portToHex(port)
        }
        // 创建流表项
        newFlow := Flow{
            Priority:    10,
            Timeout:     0,
            IsPermanent: true,
            TableID:     "1", // IPv4表ID
            DeviceID:    deviceID,
            Treatment: Treatment{
                Instructions: []Instruction{
                    {
                        Type:         "PROTOCOL_INDEPENDENT",
                        Subtype:    "ACTION",
                        ActionID:     "ingress.set_next_v4_hop",
                        ActionParams: map[string]string{"dst_port": port},
                    },
                },
            },
            // ClearDeferred: false,
            Selector: Selector{
                Criteria: []Criteria{
                    {
                        Type: "PROTOCOL_INDEPENDENT",
                        Matches: []Match{
                            {Field: "hdr.ethernet.ether_type", Match: "exact", Value: params.EtherType}, // IPv4 EtherType
                            {Field: "hdr.ipv4.src_addr", Match: "exact", Value: params.SrcIPHex},
                            {Field: "hdr.ipv4.dst_addr", Match: "exact", Value: params.DstIPHex},
                        },
                    },
                },
            },
        }
        
        res = append(res, newFlow)
    }
    return PostToOnos(url, res)
}