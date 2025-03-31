package model

import "time"

// TrafficHistory 流量路径记录表
type TrafficHistory struct {
	ID       int32     `orm:"pk;auto;column(id)"`           // 主键
	SrcHost  int32     `orm:"column(src_host)"`             // 源主机
	DstHost  int32     `orm:"column(dst_host)"`             // 目的主机
	ModeName string    `orm:"column(mode_name)"`            // 模态类型
	Datetime time.Time `orm:"column(datetime);unique"`      // 发包时间
	PathInfo string    `orm:"column(path_info);size(1000)"` // 流量路径，deviceID/port按照逗号隔开
}

func (t *TrafficHistory) TableName() string {
	return "t_traffic_historys"
}
