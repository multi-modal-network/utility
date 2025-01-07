package main

import (
	"github.com/beego/beego/v2/server/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"onosutil/logic"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	// 初始化数据库
	ormer, err := SetupORM(viper.Sub("db"))
	if err != nil {
		log.Fatal(err)
	}
	// 初始化Manager
	manager, err := logic.NewManager(logic.Options{Ormer: ormer})
	if err != nil {
		log.Fatal(err)
	}
	// 初始化beego路由
	setupRouter(manager)

	web.Run(":8188")
}

type MainController struct {
	web.Controller
}
