package ares

import (
	"fmt"
	alog "local/util/log"
	"time"

	pb "local/ares/protos"

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
) *Server {

	server := Server{
		quit:   quit,
		logger: logger,

		clientIDDic: make(map[string]string),
		tracer:      tracer,
		apps:        make(map[string]*SApp),
		nodes:       make(map[string]*Node),
	}

	return &server
}

type Server struct {
	logger      alog.Logger
	clientIDDic map[string]string
	quit        *quitInformation

	tracer *alog.Tracer
	apps   map[string]*SApp
	nodes  map[string]*Node
}

func (s *Server) Close() {
	if s.tracer != nil {
		s.tracer.Close()
	}
}

// Init the ares environment.
func (s *Server) Init() {
	s.logger.Info("Initializing ...")
	s.init()
}

func (s *Server) init() {
	go func() {
		for {
			select {
			case <-time.After(5 * time.Second):
				s.checkAvailability()
			}
		}
	}()
}

// Start the service
func (s *Server) start() {
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

func (s *Server) getNodeName(ctx context.Context) string {
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

//
func (s *Server) Login(ctx context.Context, in *pb.Node) (*pb.Empty, error) {
	if s.tracer != nil {
		span := s.tracer.StartTraceWithCtx(ctx, "Login")
		defer span.Finish()
	}

	s.registerClient(ctx, in.Name)
	s.nodes[in.Name] = newNode(in.Name)

	s.logger.Info("The client[%v] has been logged in.", in.Name)
	return &pb.Empty{}, nil
}

func (s *Server) Logout(ctx context.Context, in *pb.Empty) (*pb.Empty, error) {
	if s.tracer != nil {
		span := s.tracer.StartTraceWithCtx(ctx, "Logout")
		defer span.Finish()
	}

	s.logger.Info("The client[%v] has been logged out.", s.getNodeName(ctx))
	return &pb.Empty{}, nil
}

func (s *Server) RegisterApps(ctx context.Context, in *pb.Apps) (*pb.Empty, error) {
	if s.tracer != nil {
		span := s.tracer.StartTraceWithCtx(ctx, "RegisterApps")
		defer span.Finish()
	}

	node := s.getNodeName(ctx)
	s.resetNode(node)
	for _, app := range in.Apps {
		s.registerApp(node, app.Name)
	}

	return &pb.Empty{}, nil
}

func (s *Server) resetNode(node string) {
	for _, app := range s.apps {
		app.resetNode(node)
	}
}

func (s *Server) registerApp(node, appName string) {

	if _, ok := s.apps[appName]; !ok {
		s.apps[appName] = newSApp(appName)
	}
	s.apps[appName].registerNode(node)
}

func (s *Server) NotifyAppsStatus(ctx context.Context, in *pb.AppsStatus) (*pb.Empty, error) {
	if s.tracer != nil {
		span := s.tracer.StartTraceWithCtx(ctx, "NodifyAppStatus")
		defer span.Finish()
	}

	nodeName := s.getNodeName(ctx)

	for appName, status := range in.Status {
		if app, ok := s.apps[appName]; ok {
			app.setStatus(nodeName, (int)(status.Mode))
		}
	}

	return &pb.Empty{}, nil
}

func (s *Server) HeartBeat(ctx context.Context, in *pb.Empty) (*pb.Empty, error) {
	if s.tracer != nil {
		span := s.tracer.StartTraceWithCtx(ctx, "HeartBeat")
		defer span.Finish()
	}

	nodeName := s.getNodeName(ctx)
	if node, ok := s.nodes[nodeName]; ok {
		node.onHeartBeat()
	}

	return &pb.Empty{}, nil
}

func (s *Server) SubscribeCommands(in *pb.Node, stream pb.AresService_SubscribeCommandsServer) error {

	s.logger.Debug("SubscribeCommands node[%s]", in.Name)
	c1 := make(chan int)
	go func(nodeName string) {
		s.quit.add()
		defer s.quit.done()

		if node, ok := s.nodes[nodeName]; ok {

			for {
				select {
				case app := <-node.cmd:
					err := stream.SendMsg(&pb.App{Name: app})
					if err != nil {
						c1 <- 1
						//return err
						s.logger.Error("failed to send command: %s", err)
					}
					break
				}
			}
		}
	}(in.Name)
	_ = <-c1
	return nil
}

func (s *Server) checkAvailability() {
	for _, app := range s.apps {
		if app.status == AppStatus_Off {
			continue
		}
		if !s.nodeAvailable(app.lastNode) && app.status == AppStatus_On {
			app.status = AppStatus_Fail
		}

		if app.status == AppStatus_Fail {
			// run new app
			s.turnOnApp(app.name)
		}
	}
}

func (s *Server) turnOnApp(appName string) {

	app, ok := s.apps[appName]
	if !ok {
		return
	}

	node := s.getNextNode(appName)
	if node != nil {
		app.setStatus(node.name, AppStatus_On)
		node.sendCmd(appName)
	} else {
		s.logger.Info("no new node found for app: %s", appName)
	}
}

func (s *Server) nodeAvailable(nodeName string) bool {
	if node, ok := s.nodes[nodeName]; ok {
		if node.isActive() {
			return true
		}
	}
	return false
}

func (s *Server) getNextNode(appName string) *Node {
	app, ok := s.apps[appName]
	if !ok {
		return nil
	}

	for _, nodeName := range app.nodes {
		node, ok := s.nodes[nodeName]
		if !ok {
			continue
		}
		if node.isActive() {
			return node
		}
	}

	return nil
}

func (s *Server) showNodes() string {
	ret := ""
	if len(s.nodes) == 0 {
		return "No nodes found."
	}
	ret = ret + fmt.Sprintf("%4s|%15s|%12s\n", "No", "Node", "Availability")

	index := 1
	for nodeName, node := range s.nodes {
		statusStr := "On"
		if !node.isActive() {
			statusStr = "Off"
		}
		ret = ret + fmt.Sprintf("%4d|%15s|%12s\n", index, nodeName, statusStr)
		index = index + 1
	}
	return ret
}

func (s *Server) showApps() string {
	ret := ""
	if len(s.apps) == 0 {
		return "No apps found."
	}
	ret = ret + fmt.Sprintf("%4s|%15s|%12s\n", "No", "App", "Availability")

	index := 1
	for appName, app := range s.apps {
		statusStr := "On"
		if app.status == AppStatus_On {
			statusStr = "On"
		} else if app.status == AppStatus_Fail {
			statusStr = "Pending"
		} else {
			statusStr = "Off"
		}
		ret = ret + fmt.Sprintf("%4d|%15s|%12s\n", index, appName, statusStr)
		index = index + 1
	}
	return ret
}

func (s *Server) reqTurnOnApp(appName string) string {
	app, ok := s.apps[appName]
	if !ok {
		return fmt.Sprintf(`Can't find the app: %s`, appName)
	}

	if app.status != AppStatus_Off {
		return fmt.Sprintf(`App already on: %s`, appName)
	}

	s.turnOnApp(appName)

	return ""
}
