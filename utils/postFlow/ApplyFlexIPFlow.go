package postflow

import (
    "fmt"
    "onosutil/utils/packetparse"
    "strings"
)

func ApplyFlexIPFlow(flows []string, url string, params packetparse.FLEXIPParams) (string, error) {
    if len(flows) == 0 {
        return "no flows found", nil
    }
    // 构建流表
    res:=[]Flow{}
    for _, flow := range flows {
        //从flow解析deviceID和端口
        flowInfo:= strings.Split(flow, "/")
        if len(flowInfo) < 2 {
            return "invalid flow format", fmt.Errorf("invalid flow format: %s", flow)
        }
        deviceID, port := flowInfo[0], flowInfo[1]
        // 创建流表项
        newFlow := Flow{
            Priority:    10,
            Timeout:     0,
            IsPermanent: true,
            TableID:     "6",
            DeviceID:    deviceID,
            Treatment: Treatment{
                Instructions: []Instruction{
                    {
                        Type:         "PI_ACTION",
                        ActionID:     "ingress.set_next_flexip_hop",
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
                            {Field: "hdr.ethernet.ether_type", Match: "exact", Value: params.EtherType}, // 添加EtherType匹配
                            {Field: "hdr.flexip.src_format", Match: "exact", Value: params.SrcFormat},
                            {Field: "hdr.flexip.dst_format", Match: "exact", Value: params.DstFormat},
                            {Field: "hdr.flexip.src_addr", Match: "exact", Value: params.SrcAddr},
                            {Field: "hdr.flexip.dst_addr", Match: "exact", Value: params.DstAddr},
                        },
                    },
                },
            },
        }
        
        res = append(res, newFlow)
    }
	// 将流表项发送到ONOS
    return PostToOnos(url, res)
}