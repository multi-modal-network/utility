package postflow

import (
    "fmt"
    "onosutil/utils/packetparse"
    "strings"
)

func ApplyNdnFlow(flows []string, url string, params packetparse.NDNParams) (string, error) {
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
            TableID:     "4", // NDN表ID
            DeviceID:    deviceID,
            Treatment: Treatment{
                Instructions: []Instruction{
                    {
                        Type:         "PI_ACTION",
                        ActionID:     "ingress.set_next_ndn_hop",
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
                            {Field: "hdr.ethernet.ether_type", Match: "exact", Value:params.EtherType}, // NDN协议EtherType
                            {Field: "hdr.ndn.ndn_prefix.code", Match: "exact", Value: "6"}, // NDN代码固定为6
                            {Field: "hdr.ndn.name_tlv.components[0].value", Match: "exact", Value: params.SrcNDNName},
                            {Field: "hdr.ndn.name_tlv.components[1].value", Match: "exact", Value: params.DstNDNName},
                            {Field: "hdr.ndn.content_tlv.value", Match: "exact", Value: params.NdnContent},
                        },
                    },
                },
            },
        }
        
        res = append(res, newFlow)
    }
    return PostToOnos(url, res)
}