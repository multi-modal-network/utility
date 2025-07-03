package postflow

import (
	"fmt"
	"onosutil/utils/packetparse"
	"strings"
)

func ApplyGEOFlow(flows []string, url string, params packetparse.GEOParams) (string, error) {
    if len(flows) == 0 {
        return "no flows found", nil
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
            TableID:     "3", // GEO表ID
            DeviceID:    deviceID,
            Treatment: Treatment{
                Instructions: []Instruction{
                    {
                        Type:         "PI_ACTION",
                        ActionID:     "ingress.geo_ucast_route",
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
                            {Field: "hdr.ethernet.ether_type", Match: "exact", Value: params.EtherType}, // GEO EtherType
                            {Field: "hdr.gbc.geo_area_pos_lat", Match: "exact", Value: params.GeoAreaPosLat},
                            {Field: "hdr.gbc.geo_area_pos_lon", Match: "exact", Value: params.GeoAreaPosLon},
                            {Field: "hdr.gbc.disa", Match: "exact", Value: params.Disa},
                            {Field: "hdr.gbc.disb", Match: "exact", Value: params.Disb},
                        },
                    },
                },
            },
        }
        
        res = append(res, newFlow)
    }
    return PostToOnos(url, res)
}