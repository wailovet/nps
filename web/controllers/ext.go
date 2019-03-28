package controllers

import (
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/lib/rate"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego"
	"strconv"
)

type ExtController struct {
	beego.Controller
}

func (s *ExtController) Test() {

	s.AjaxOk("success")
}

//添加客户端
func (s *ExtController) Add() {
	t := &file.Client{
		VerifyKey: s.GetString("vkey"),
		Id:        int(file.GetCsvDb().GetClientId()),
		Status:    true,
		Remark:    s.GetString("remark"),
		Cnf: &file.Config{
			U:        s.GetString("u"),
			P:        s.GetString("p"),
			Compress: common.GetBoolByStr(s.GetString("compress")),
			Crypt:    s.GetBoolNoErr("crypt"),
		},
		ConfigConnAllow: s.GetBoolNoErr("config_conn_allow"),
		RateLimit:       s.GetIntNoErr("rate_limit"),
		MaxConn:         s.GetIntNoErr("max_conn"),
		Flow: &file.Flow{
			ExportFlow: 0,
			InletFlow:  0,
			FlowLimit:  int64(s.GetIntNoErr("flow_limit")),
		},
	}
	if t.RateLimit > 0 {
		t.Rate = rate.NewRate(int64(t.RateLimit * 1024))
		t.Rate.Start()
	}
	if err := file.GetCsvDb().NewClient(t); err != nil {
		s.AjaxErr(err.Error())
	}
	s.AjaxOk("add success")
}

//去掉没有err返回值的int
func (s *ExtController) GetIntNoErr(key string, def ...int) int {
	strv := s.Ctx.Input.Query(key)
	if len(strv) == 0 && len(def) > 0 {
		return def[0]
	}
	val, _ := strconv.Atoi(strv)
	return val
}

//获取去掉错误的bool值
func (s *ExtController) GetBoolNoErr(key string, def ...bool) bool {
	strv := s.Ctx.Input.Query(key)
	if len(strv) == 0 && len(def) > 0 {
		return def[0]
	}
	val, _ := strconv.ParseBool(strv)
	return val
}

//ajax正确返回
func (s *ExtController) AjaxOk(str string) {
	s.Data["json"] = ajax(str, 1)
	s.ServeJSON()
	s.StopRun()
}

//ajax错误返回
func (s *ExtController) AjaxErr(str string) {
	s.Data["json"] = ajax(str, 0)
	s.ServeJSON()
	s.StopRun()
}
