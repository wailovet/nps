package proxy

import (
	"errors"
	"github.com/wailovet/nps/bridge"
	"github.com/wailovet/nps/lib/common"
	"github.com/wailovet/nps/lib/conn"
	"github.com/wailovet/nps/lib/file"
	"github.com/wailovet/nps/vender/github.com/astaxie/beego/logs"
	"net"
	"net/http"
	"sync"
)

type Service interface {
	Start() error
	Close() error
}

//Server BaseServer struct
type BaseServer struct {
	id           int
	bridge       *bridge.Bridge
	task         *file.Tunnel
	errorContent []byte
	sync.Mutex
}

func NewBaseServer(bridge *bridge.Bridge, task *file.Tunnel) *BaseServer {
	return &BaseServer{
		bridge:       bridge,
		task:         task,
		errorContent: nil,
		Mutex:        sync.Mutex{},
	}
}

func (s *BaseServer) FlowAdd(in, out int64) {
	s.Lock()
	defer s.Unlock()
	s.task.Flow.ExportFlow += out
	s.task.Flow.InletFlow += in
}

func (s *BaseServer) FlowAddHost(host *file.Host, in, out int64) {
	s.Lock()
	defer s.Unlock()
	host.Flow.ExportFlow += out
	host.Flow.InletFlow += in
}

func (s *BaseServer) writeConnFail(c net.Conn) {
	c.Write([]byte(common.ConnectionFailBytes))
	c.Write(s.errorContent)
}

//权限认证
func (s *BaseServer) auth(r *http.Request, c *conn.Conn, u, p string) error {
	if u != "" && p != "" && !common.CheckAuth(r, u, p) {
		c.Write([]byte(common.UnauthorizedBytes))
		c.Close()
		return errors.New("401 Unauthorized")
	}
	return nil
}

func (s *BaseServer) CheckFlowAndConnNum(client *file.Client) error {
	if client.Flow.FlowLimit > 0 && (client.Flow.FlowLimit<<20) < (client.Flow.ExportFlow+client.Flow.InletFlow) {
		return errors.New("Traffic exceeded")
	}
	if !client.GetConn() {
		return errors.New("Connections exceed the current client limit")
	}
	return nil
}

//与客户端建立通道
func (s *BaseServer) DealClient(c *conn.Conn, client *file.Client, addr string, rb []byte, tp string, f func(), flow *file.Flow) error {
	link := conn.NewLink(tp, addr, client.Cnf.Crypt, client.Cnf.Compress, c.Conn.RemoteAddr().String())
	if target, err := s.bridge.SendLinkInfo(client.Id, link, c.Conn.RemoteAddr().String(), s.task); err != nil {
		logs.Warn("task id %d get connection from client id %d  error %s", s.task.Id, client.Id, err.Error())
		c.Close()
		return err
	} else {
		if f != nil {
			f()
		}
		conn.CopyWaitGroup(target, c.Conn, link.Crypt, link.Compress, client.Rate, flow, true, rb)
	}
	client.AddConn()
	return nil
}
