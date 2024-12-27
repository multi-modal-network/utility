package main

import (
	"github.com/beego/beego/v2/server/web"
	log "github.com/sirupsen/logrus"
	"onosutil/logic"
	"onosutil/model"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	ormer, err := model.SetupORM()
	if err != nil {
		log.Fatal(err)
	}

	manager, err := logic.NewManager(logic.Options{
		Ormer: ormer,
	})
	if err != nil {
		log.Fatal(err)
	}

	setupRouter(manager)

	web.Router("/", &MainController{})
	web.Run(":8088")
}

type MainController struct {
	web.Controller
}
