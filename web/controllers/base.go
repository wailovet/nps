package controllers

import (
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/crypt"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/server"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego"
	"math"
	"strconv"
	"strings"
	"time"
)

type BaseController struct {
	beego.Controller
	controllerName string
	actionName     string
}

//初始化参数
func (s *BaseController) Prepare() {
	controllerName, actionName := s.GetControllerAndAction()
	s.controllerName = strings.ToLower(controllerName[0 : len(controllerName)-10])
	s.actionName = strings.ToLower(actionName)
	// web api verify
	// param 1 is md5(authKey+Current timestamp)
	// param 2 is timestamp (It's limited to 20 seconds.)
	md5Key := s.GetString("auth_key")
	timestamp := s.GetIntNoErr("timestamp")
	configKey := beego.AppConfig.String("auth_key")
	timeNowUnix := time.Now().Unix()
	if !((math.Abs(float64(timeNowUnix-int64(timestamp))) <= 20) && (crypt.Md5(configKey+strconv.Itoa(timestamp)) == md5Key)) {
		if s.GetSession("auth") != true {
			s.Redirect("/login/index", 302)
		}
	}
	if s.GetSession("isAdmin") != nil && !s.GetSession("isAdmin").(bool) {
		s.Ctx.Input.SetData("client_id", s.GetSession("clientId").(int))
		s.Ctx.Input.SetParam("client_id", strconv.Itoa(s.GetSession("clientId").(int)))
		s.Data["isAdmin"] = false
		s.CheckUserAuth()
	} else {
		s.Data["isAdmin"] = true
	}
}

//加载模板
func (s *BaseController) display(tpl ...string) {
	var tplname string
	if s.Data["menu"] == nil {
		s.Data["menu"] = s.actionName
	}
	if len(tpl) > 0 {
		tplname = strings.Join([]string{tpl[0], "html"}, ".")
	} else {
		tplname = s.controllerName + "/" + s.actionName + ".html"
	}
	ip := s.Ctx.Request.Host
	s.Data["ip"] = common.GetIpByAddr(ip)
	s.Data["bridgeType"] = beego.AppConfig.String("bridge_type")
	if common.IsWindows() {
		s.Data["win"] = ".exe"
	}
	s.Data["p"] = server.Bridge.TunnelPort
	s.Data["proxyPort"] = beego.AppConfig.String("hostPort")
	s.Layout = "public/layout.html"
	s.TplName = tplname
}

//错误
func (s *BaseController) error() {
	s.Layout = "public/layout.html"
	s.TplName = "public/error.html"
}

//去掉没有err返回值的int
func (s *BaseController) GetIntNoErr(key string, def ...int) int {
	strv := s.Ctx.Input.Query(key)
	if len(strv) == 0 && len(def) > 0 {
		return def[0]
	}
	val, _ := strconv.Atoi(strv)
	return val
}

//获取去掉错误的bool值
func (s *BaseController) GetBoolNoErr(key string, def ...bool) bool {
	strv := s.Ctx.Input.Query(key)
	if len(strv) == 0 && len(def) > 0 {
		return def[0]
	}
	val, _ := strconv.ParseBool(strv)
	return val
}

//ajax正确返回
func (s *BaseController) AjaxOk(str string) {
	s.Data["json"] = ajax(str, 1)
	s.ServeJSON()
	s.StopRun()
}

//ajax错误返回
func (s *BaseController) AjaxErr(str string) {
	s.Data["json"] = ajax(str, 0)
	s.ServeJSON()
	s.StopRun()
}

//组装ajax
func ajax(str string, status int) map[string]interface{} {
	json := make(map[string]interface{})
	json["status"] = status
	json["msg"] = str
	return json
}

//ajax table返回
func (s *BaseController) AjaxTable(list interface{}, cnt int, recordsTotal int) {
	json := make(map[string]interface{})
	json["rows"] = list
	json["total"] = recordsTotal
	s.Data["json"] = json
	s.ServeJSON()
	s.StopRun()
}

//ajax table参数
func (s *BaseController) GetAjaxParams() (start, limit int) {
	return s.GetIntNoErr("offset"), s.GetIntNoErr("limit")
}

func (s *BaseController) SetInfo(name string) {
	s.Data["name"] = name
}

func (s *BaseController) SetType(name string) {
	s.Data["type"] = name
}

func (s *BaseController) CheckUserAuth() {
	if s.controllerName == "client" {
		if s.actionName == "add" {
			s.StopRun()
			return
		}
		if id := s.GetIntNoErr("id"); id != 0 {
			if id != s.GetSession("clientId").(int) {
				s.StopRun()
				return
			}
		}
	}
	if s.controllerName == "index" {
		if id := s.GetIntNoErr("id"); id != 0 {
			belong := false
			if strings.Contains(s.actionName, "h") {
				if v, ok := file.GetCsvDb().Hosts.Load(id); ok {
					if v.(*file.Host).Client.Id == s.GetSession("clientId").(int) {
						belong = true
					}
				}
			} else {
				if v, ok := file.GetCsvDb().Tasks.Load(id); ok {
					if v.(*file.Tunnel).Client.Id == s.GetSession("clientId").(int) {
						belong = true
					}
				}
			}
			if !belong {
				s.StopRun()
			}
		}
	}
}
