package model

import (
	_ "github.com/go-sql-driver/mysql"
)

// Device 设备结构
type Device struct {
	ID                int32  `orm:"pk;auto;column(id)"`         // 主键
	DeviceID          string `orm:"column(device_id);unique"`   // 设备ID，形如"device:domain1:group2:level6:s305_ID_IP_NDN_GEO_FLEXIP"
	DeviceName        string `orm:"column(device_name)"`        // 设备名称，设备ID的前缀，形如"device:domain1:group2:level6:s305"
	Domain            int32  `orm:"column(domain)"`             // 域，domain后接的数字
	Group             int32  `orm:"column(group)"`              // 组，group后接的数字
	SwitchID          int32  `orm:"column(switch_id)"`          // 交换机ID，s后接的数字
	ManagementAddress string `orm:"column(management_address)"` // grpc地址，netcfg中配置
	Driver            string `orm:"column(driver)"`             // 设备驱动，netcfg中配置
	Pipeconf          string `orm:"column(pipeconf)"`           // 设备流水线，netcfg中配置，例如"org.stratumproject.IP_ID_GEO_MF_NDN_FLEXIP.bmv2"
	SupportModal      string `orm:"column(support_modal)"`      // 设置支持的模态类型，形如"ID_IP_NDN_GEO_FLEXIP"
}

type Link struct {
	ID        int32  `orm:"pk;auto;column(id)"`
	EndPoint1 string `orm:"column(endPoint1)"`
	EndPoint2 string `orm:"column(endPoint2)"`
}

// DevicePort 交换机及转发端口结构
type DevicePort struct {
	DeviceName string
	Port       int32
}

func (d *Device) TableName() string {
	return "t_devices"
}

// TableUnique 建立（设备名称_支持模态）的唯一索引
func (d *Device) TableUnique() [][]string {
	return [][]string{
		{"DeviceName", "SupportModal"},
	}
}

func (l *Link) TableName() string {
	return "t_links"
}

// TableUnique 建立（EndPoint1_EndPoint2）唯一索引
func (l *Link) TableUnique() [][]string {
	return [][]string{
		{"EndPoint1", "EndPoint2"},
	}
}
