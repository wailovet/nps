package controllers

import (
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/server"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego"
	"time"
)

type LoginController struct {
	beego.Controller
}

func (self *LoginController) Index() {
	self.TplName = "login/index.html"
}
func (self *LoginController) Verify() {
	var auth bool
	if self.GetString("password") == beego.AppConfig.String("web_password") && self.GetString("username") == beego.AppConfig.String("web_username") {
		self.SetSession("isAdmin", true)
		auth = true
		server.Bridge.Register.Store(common.GetIpByAddr(self.Ctx.Request.RemoteAddr), time.Now().Add(time.Hour*time.Duration(2)))
	}
	b, err := beego.AppConfig.Bool("allow_user_login")
	if err == nil && b && self.GetString("username") == "user" && !auth {
		file.GetCsvDb().Clients.Range(func(key, value interface{}) bool {
			v := value.(*file.Client)
			if v.VerifyKey == self.GetString("password") && v.Status {
				self.SetSession("isAdmin", false)
				self.SetSession("clientId", v.Id)
				auth = true
				return false
			}
			return true
		})
	}
	if auth {
		self.SetSession("auth", true)
		self.Data["json"] = map[string]interface{}{"status": 1, "msg": "login success"}
	} else {
		self.Data["json"] = map[string]interface{}{"status": 0, "msg": "username or password incorrect"}
	}
	self.ServeJSON()
}
func (self *LoginController) Out() {
	self.SetSession("auth", false)
	self.Redirect("/login/index", 302)
}
