package model

import (
	_ "github.com/go-sql-driver/mysql"
)

// Device 设备结构
type Device struct {
	ID                int32  `orm:"pk;auto;column(id)"`
	DeviceID          string `orm:"column(deviceID);unique"`
	Domain            int32  `orm:"column(domain)"`
	Group             int32  `orm:"column(group)"`
	SwitchID          string `orm:"column(switchID)"`
	ManagementAddress string `orm:"column(management_address)"`
	Driver            string `orm:"column(driver)"`
	Pipeconf          string `orm:"column(pipeconf)"`
}

type Link struct {
	ID        int32  `orm:"pk;auto;column(id)"`
	EndPoint1 string `orm:"column(endPoint1)"`
	EndPoint2 string `orm:"column(endPoint2)"`
}

func (d *Device) TableName() string {
	return "t_devices"
}

// TableUnique 建立多字段唯一索引
func (l *Link) TableUnique() [][]string {
	return [][]string{
		{"EndPoint1", "EndPoint2"},
	}
}

func (l *Link) TableName() string {
	return "t_links"
}
