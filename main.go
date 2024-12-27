package main

import (
	"github.com/beego/beego/v2/server/web"
	"onosutil/model"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	if err := model.SetupORM(); err != nil {
		panic(err)
	}

	web.Router("/", &MainController{}) //TIP Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined or highlighted text
	// to see how GoLand suggests fixing it.
	web.Run(":8088")
}

type MainController struct {
	web.Controller
}

func (c *MainController) Get() {
	name := c.GetString("name")
	if name == "" {
		c.Ctx.WriteString("Hello World")
		return
	}
	c.Ctx.WriteString("Hello " + name)
}

func (c *MainController) Post() {
	netcfg := NetConf{}
	if err := c.BindJSON(&netcfg); err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}
	//fmt.Println(netcfg)
	//for deviceID, _ := range netcfg.Devices {
	//	c.Ctx.WriteString(deviceID + "\n")
	//}
	//c.Ctx.WriteString(fmt.Sprintf("Number of Links: %d", len(netcfg.Links)))

}

type NetConf struct {
	Devices map[string]Devices `json:"devices"`
	Links   []Links            `json:"links"`
}

type Links struct {
	EndPoint1 string `json:"endpoint1"`
	EndPoint2 string `json:"endpoint2"`
}

type Devices struct {
	Device map[string]DeviceBasic
}

type DeviceBasic struct {
	Basic map[string]DeviceInfo `json:"basic"`
}

type DeviceInfo struct {
	ManagementAddress string `json:"managementAddress"`
	Driver            string `json:"driver"`
	Pipeconf          string `json:"pipeconf"`
}
