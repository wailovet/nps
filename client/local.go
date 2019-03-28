package client

import (
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/config"
	"github.com/wailovet/nps/lib/conn"
	"github.com/wailovet/nps/lib/crypt"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/lib/mux"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego/logs"
	"github.com/wailovet/nps/vender/github.com/xtaci/kcp"
	"net"
	"net/http"
)

var LocalServer []*net.TCPListener
var udpConn net.Conn
var muxSession *mux.Mux
var fileServer []*http.Server

func CloseLocalServer() {
	for _, v := range LocalServer {
		v.Close()
	}
	for _, v := range fileServer {
		v.Close()
	}
}

func startLocalFileServer(config *config.CommonConfig, t *file.Tunnel, vkey string) {
	remoteConn, err := NewConn(config.Tp, vkey, config.Server, common.WORK_FILE, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	srv := &http.Server{
		Handler: http.StripPrefix(t.StripPre, http.FileServer(http.Dir(t.LocalPath))),
	}
	logs.Info("start local file system, local path %s, strip prefix %s ,remote port %s ", t.LocalPath, t.StripPre, t.Ports)
	fileServer = append(fileServer, srv)
	listener := mux.NewMux(remoteConn.Conn, common.CONN_TCP)
	logs.Warn(srv.Serve(listener))
}

func StartLocalServer(l *config.LocalServer, config *config.CommonConfig) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), l.Port, ""})
	if err != nil {
		logs.Error("local listener startup failed port %d, error %s", l.Port, err.Error())
		return err
	}
	LocalServer = append(LocalServer, listener)
	logs.Info("successful start-up of local monitoring, port", l.Port)
	conn.Accept(listener, func(c net.Conn) {
		if l.Type == "secret" {
			processSecret(c, config, l)
		} else {
			processP2P(c, config, l)
		}
	})
	return nil
}

func processSecret(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) {
	remoteConn, err := NewConn(config.Tp, config.VKey, config.Server, common.WORK_SECRET, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	if _, err := remoteConn.Write([]byte(crypt.Md5(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	conn.CopyWaitGroup(remoteConn.Conn, localTcpConn, false, false, nil, nil, false, nil)
}

func processP2P(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) {
	if udpConn == nil {
		newUdpConn(config, l)
		if udpConn == nil {
			return
		}
		muxSession = mux.NewMux(udpConn, "kcp")
	}
	nowConn, err := muxSession.NewConn()
	if err != nil {
		logs.Error(err)
		return
	}
	link := conn.NewLink(common.CONN_TCP, l.Target, config.Cnf.Crypt, config.Cnf.Compress, localTcpConn.LocalAddr().String())
	if _, err := conn.NewConn(nowConn).SendLinkInfo(link); err != nil {
		logs.Error(err)
		return
	}
	conn.CopyWaitGroup(nowConn, localTcpConn, config.Cnf.Crypt, config.Cnf.Compress, nil, nil, false, nil)
}

func newUdpConn(config *config.CommonConfig, l *config.LocalServer) {
	remoteConn, err := NewConn(config.Tp, config.VKey, config.Server, common.WORK_P2P, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	if _, err := remoteConn.Write([]byte(crypt.Md5(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	var rAddr []byte
	//读取服务端地址、密钥 继续做处理
	if rAddr, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	//与服务端udp建立连接
	tmpConn, err := common.GetLocalUdpAddr()
	if err != nil {
		logs.Warn(err)
		return
	}
	//与服务端建立udp连接
	localAddr, _ := net.ResolveUDPAddr("udp", tmpConn.LocalAddr().String())
	localConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		logs.Error(err)
		return
	}
	localKcpConn, err := kcp.NewConn(string(rAddr), nil, 150, 3, localConn)
	if err != nil {
		logs.Error(err)
		return
	}
	conn.SetUdpSession(localKcpConn)
	//写入密钥、provider身份
	if _, err := localKcpConn.Write([]byte(crypt.Md5(l.Password))); err != nil {
		logs.Error(err)
		return
	}
	if _, err := localKcpConn.Write([]byte(common.WORK_P2P_VISITOR)); err != nil {
		logs.Error(err)
		return
	}
	//接收服务端传的visitor地址
	if b, err := conn.NewConn(localKcpConn).GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	} else {
		//关闭与服务端连接
		localConn.Close()
		//建立新的连接
		localConn, err = net.ListenUDP("udp", localAddr)
		udpTunnel, err := kcp.NewConn(string(b), nil, 150, 3, localConn)
		if err != nil || udpTunnel == nil {
			logs.Warn(err)
			return
		}
		conn.SetUdpSession(udpTunnel)
		udpConn = udpTunnel
	}
}
