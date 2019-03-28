package main

import (
	"flag"
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/crypt"
	"github.com/wailovet/nps/lib/daemon"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/lib/install"
	"github.com/wailovet/nps/lib/version"
	"github.com/wailovet/nps/server"
	"github.com/wailovet/nps/server/connection"
	"github.com/wailovet/nps/server/test"
	"github.com/wailovet/nps/server/tool"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego/logs"
	_ "github.com/wailovet/nps/web/routers"
	"log"
	"os"
	"path/filepath"
)

var (
	level   string
	logType = flag.String("log", "stdout", "Log output mode（stdout|file）")
)

func main() {
	flag.Parse()
	beego.LoadAppConfig("ini", filepath.Join(common.GetRunPath(), "conf", "nps.conf"))
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "test":
			test.TestServerConfig()
			log.Println("test ok, no error")
			return
		case "start", "restart", "stop", "status", "reload":
			daemon.InitDaemon("nps", common.GetRunPath(), common.GetTmpPath())
		case "install":
			install.InstallNps()
			return
		}
	}
	if level = beego.AppConfig.String("log_level"); level == "" {
		level = "7"
	}
	logs.Reset()
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	if *logType == "stdout" {
		logs.SetLogger(logs.AdapterConsole, `{"level":`+level+`,"color":true}`)
	} else {
		logs.SetLogger(logs.AdapterFile, `{"level":`+level+`,"filename":"nps_log.log","daily":false,"color":true}`)
	}
	task := &file.Tunnel{
		Mode: "webServer",
	}
	bridgePort, err := beego.AppConfig.Int("bridge_port")
	if err != nil {
		logs.Error("Getting bridge_port error", err)
		os.Exit(0)
	}
	logs.Info("the version of server is %s ,allow client version to be %s", version.VERSION, version.GetVersion())
	connection.InitConnectionService()
	crypt.InitTls(filepath.Join(beego.AppPath, "conf", "server.pem"), filepath.Join(beego.AppPath, "conf", "server.key"))
	tool.InitAllowPort()
	server.StartNewServer(bridgePort, task, beego.AppConfig.String("bridge_type"))
}
