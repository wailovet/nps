package main

import (
	"flag"
	"github.com/wailovet/nps/client"
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/daemon"
	"github.com/wailovet/nps/lib/version"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego/logs"
	"os"
	"strings"
	"time"
)

var (
	serverAddr   = flag.String("server", "", "Server addr (ip:port)")
	configPath   = flag.String("config", "", "Configuration file path")
	verifyKey    = flag.String("vkey", "", "Authentication key")
	logType      = flag.String("log", "stdout", "Log output mode（stdout|file）")
	connType     = flag.String("type", "tcp", "Connection type with the server（kcp|tcp）")
	proxyUrl     = flag.String("proxy", "", "proxy socks5 url(eg:socks5://111:222@127.0.0.1:9007)")
	logLevel     = flag.String("log_level", "7", "log level 0~7")
	registerTime = flag.Int("time", 2, "register time long /h")
)

func main() {
	flag.Parse()
	if len(os.Args) > 2 {
		switch os.Args[1] {
		case "status":
			path := strings.Replace(os.Args[2], "-config=", "", -1)
			client.GetTaskStatus(path)
		case "register":
			flag.CommandLine.Parse(os.Args[2:])
			client.RegisterLocalIp(*serverAddr, *verifyKey, *connType, *proxyUrl, *registerTime)
		}
	}
	daemon.InitDaemon("npc", common.GetRunPath(), common.GetTmpPath())
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	if *logType == "stdout" {
		logs.SetLogger(logs.AdapterConsole, `{"level":`+*logLevel+`,"color":true}`)
	} else {
		logs.SetLogger(logs.AdapterFile, `{"level":`+*logLevel+`,"filename":"npc_log.log","daily":false,"color":true}`)
	}
	env := common.GetEnvMap()
	if *serverAddr == "" {
		*serverAddr, _ = env["NPC_SERVER_ADDR"]
	}
	if *verifyKey == "" {
		*verifyKey, _ = env["NPC_SERVER_VKEY"]
	}
	logs.Info("the version of client is %s", version.VERSION)
	if *verifyKey != "" && *serverAddr != "" && *configPath == "" {
		for {
			client.NewRPClient(*serverAddr, *verifyKey, *connType, *proxyUrl, nil).Start()
			logs.Info("It will be reconnected in five seconds")
			time.Sleep(time.Second * 5)
		}
	} else {
		if *configPath == "" {
			*configPath = "npc.conf"
		}
		client.StartFromFile(*configPath)
	}
}
