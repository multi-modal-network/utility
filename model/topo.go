package model

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
)

// Device 设备结构
type Device struct {
	ID       int32  `orm:"pk;auto;column(id)"`
	DeviceID string `orm:"column(deviceID);unique"`
	Domain   int32  `orm:"column(domain)"`
	Group    int32  `orm:"column(group)"`
	SwitchID string `orm:"column(switchID)"`
}

type Link struct {
	ID        int32  `orm:"pk;auto;column(id)"`
	EndPoint1 string `orm:"column(endPoint1)"`
	EndPoint2 string `orm:"column(endPoint2)"`
}

// TableUnique 建立多字段唯一索引
func (l *Link) TableUnique() [][]string {
	return [][]string{
		{"EndPoint1", "EndPoint2"},
	}
}

func SetupORM() (orm.Ormer, error) {
	orm.RegisterModel(new(Device))
	orm.RegisterModel(new(Link))
	if err := orm.RegisterDataBase("default", "mysql",
		"root:132311@tcp(127.0.0.1:3306)/onos?charset=utf8"); err != nil {
		return nil, errors.New("ORM RegisterDataBase failed")
	}
	if err := orm.RunSyncdb("default", false, true); err != nil {
		return nil, errors.New("ORM RunSyncdb failed")
	}
	return orm.NewOrm(), nil
}
