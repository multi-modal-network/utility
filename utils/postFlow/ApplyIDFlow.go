package postflow

import (
	"fmt"
	"onosutil/utils/packetparse"
	"strings"
)

func ApplyIDFlow(flows []string, url string, params packetparse.IDParams) (string, error) {
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
            TableID:     "5", // ID表ID
            DeviceID:    deviceID,
            Treatment: Treatment{
                Instructions: []Instruction{
                    {
                        Type:         "PROTOCOL_INDEPENDENT",
                        Subtype:    "ACTION",
                        ActionID:     "ingress.set_next_id_hop",
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
                            {Field: "hdr.ethernet.ether_type", Match: "exact", Value: params.EtherType}, // ID协议EtherType
                            {Field: "hdr.id.src_identity", Match: "exact", Value: params.SrcIdentity},
                            {Field: "hdr.id.dst_identity", Match: "exact", Value: params.DstIdentity},
                        },
                    },
                },
            },
        }
        
        res = append(res, newFlow)
    }
    return PostToOnos(url, res)
}