package controllers

import (
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/lib/rate"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego"
)

type ExtController struct {
	beego.Controller
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
			Crypt:    false,
		},
		ConfigConnAllow: false,
		RateLimit:       0,
		MaxConn:         0,
		Flow: &file.Flow{
			ExportFlow: 0,
			InletFlow:  0,
			FlowLimit:  int64(0),
		},
	}
	if t.RateLimit > 0 {
		t.Rate = rate.NewRate(int64(t.RateLimit * 1024))
		t.Rate.Start()
	}
	if err := file.GetCsvDb().NewClient(t); err != nil {
	}
}
