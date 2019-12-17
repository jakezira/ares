package ares

import (
	alog "local/util/log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
)

//
// NewServer Create server instance
//
func NewServer(
	quit *quitInformation,
	logger alog.Logger,
	tracer *alog.Tracer,
	nodeName string,
	managerConnString string,
) *Server {

	node := newNode(nodeName, tracer.GetTracer(), logger)

	server := Server{
		quit:   quit,
		logger: logger,

		clientIDDic: make(map[string]string),
		tracer:      tracer,

		node:              node,
		managerConnString: managerConnString,
		apps:              make(map[string]*SApp),
	}

	return &server
}

type Server struct {
	logger      alog.Logger
	clientIDDic map[string]string
	quit        *quitInformation

	tracer *alog.Tracer

	node              *Node
	managerConnString string
	apps              map[string]*SApp
}

func (s *Server) Close() {
	if s.tracer != nil {
		s.tracer.Close()
	}
}

// Init the ares environment.
func (s *Server) Init(appsConf []*AppConf) {
	s.logger.Info("Initializing ...")
	s.init(false)

	for _, conf := range appsConf {
		app := newSApp(conf.Name, conf.Run, conf.Dir, s.logger)
		s.apps[app.name] = app
	}
}

func (s *Server) init(isReset bool) {

}

// Start the service
func (s *Server) start() {
	s.connect(s.managerConnString)
	s.node.subscribeCommands()

	go func() {
		for {
			select {
			case cmd := <-s.node.ch:
				s.runCommand(cmd)
				break
			}
		}
	}()

	go func() {
		for {
			select {
			case <-time.After(time.Second * 5):
				s.onHeartBeat()
				break
			}
		}
	}()
}

func (s *Server) connect(connString string) {
	s.node.connect(connString)
	s.node.registerApps(s.apps)
}

func getPeerAddress(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}

	return peer.Addr.String()
}

func (s *Server) getClientId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := s.clientIDDic[getPeerAddress(ctx)]; ok {
		return val
	}
	return ""
}

func (s *Server) registerClient(ctx context.Context, clientName string) {

	s.clientIDDic[getPeerAddress(ctx)] = clientName
}

func (s *Server) checkPanic() {
	if r := recover(); r != nil {
		s.logger.Error("Recovered: %v", r)
	}
}

func (s *Server) runCommand(cmd string) {
	app, ok := s.apps[cmd]
	if !ok {
		s.logger.Error("runCommand: failed to run command[%s]", cmd)
		return
	}

	app.runApp()
}

func (s *Server) onHeartBeat() {
	status := make(map[string]int)
	for name, app := range s.apps {
		status[name] = app.status
	}
	s.node.heartbeat()
	s.node.sendStatus(status)
}
