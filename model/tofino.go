package model

import (
	_ "github.com/go-sql-driver/mysql"
)

// TofinoPort tofino交换机模态端口对应表
type TofinoPort struct {
	ID        int32  `orm:"pk;auto;column(id)"` // 主键
	SwitchID  int32  `orm:"column(switch_id)"`  // Tofino交换机ID
	ModalType string `orm:"column(modal_type)"` // 模态类型
	Port      int32  `orm:"column(port)"`       // 转发端口
	OldPort   int32  `orm:"column(old_port)"`   // 旧转发端口
}

func (t *TofinoPort) TableName() string {
	return "t_tofino_ports"
}

// TableUnique 建立多字段唯一索引
func (t *TofinoPort) TableUnique() [][]string {
	return [][]string{
		{"SwitchID", "ModalType"},
	}
}
