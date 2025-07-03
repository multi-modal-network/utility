package postflow

import (
	"fmt"
	"onosutil/utils/packetparse"
	"strings"
)

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
                        Type:         "PI_ACTION",
                        ActionID:     "ingress.set_next_v4_hop",
                        ActionParams: map[string]interface{}{"dst_port": port},
                    },
                },
            },
            ClearDeferred: false,
            Selector: Selector{
                Criteria: []Criteria{
                    {
                        Type: "PI_EXACT",
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