package main

import (
	"github.com/beego/beego/v2/server/web"
	"onosutil/logic"
)

func setupRouter(manager *logic.Manager) {
	web.Post("/api/topo", manager.UpdateTopoHandler)
}
