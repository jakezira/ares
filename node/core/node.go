package ares

import (
	"context"
	"io"
	ares "local/ares/protos"
	pb "local/ares/protos"
	alog "local/util/log"

	zipkin "github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"google.golang.org/grpc"
)

type Node struct {
	name   string
	tracer *zipkin.Tracer
	logger alog.Logger
	conn   *grpc.ClientConn
	cl     pb.AresServiceClient

	ch chan string
}

func newNode(name string, tracer *zipkin.Tracer, logger alog.Logger) *Node {
	return &Node{
		name:   name,
		tracer: tracer,
		logger: logger,

		ch: make(chan string),
	}
}

func (n *Node) connect(connString string) error {
	var (
		conn *grpc.ClientConn
		err  error
	)

	if n.tracer == nil {
		conn, err = grpc.Dial(
			connString,
			grpc.WithInsecure(),
		)

	} else {
		conn, err = grpc.Dial(
			connString,
			grpc.WithInsecure(),
			grpc.WithStatsHandler(zipkingrpc.NewClientHandler(n.tracer)),
		)
	}
	if err != nil {
		return err
	}

	n.conn = conn

	n.cl = pb.NewAresServiceClient(conn)
	n.login()
	n.logout()

	return nil
}

func (n Node) login() {
	n.cl.Login(
		context.Background(),
		&pb.Node{
			Name: n.name,
		},
	)
}

func (n Node) logout() {
	n.cl.Logout(
		context.Background(),
		&pb.Empty{},
	)
}

func (n Node) registerApps(apps map[string]*SApp) {

	req := &pb.Apps{}
	req.Apps = make([]*pb.App, 0)

	for _, app := range apps {
		req.Apps = append(req.Apps, &pb.App{Name: app.name})
	}

	n.cl.RegisterApps(
		context.Background(),
		req,
	)
}

func (n *Node) subscribeCommands() {
	req := &pb.Node{Name: n.name}
	stream, err := n.cl.SubscribeCommands(
		context.Background(),
		req,
	)

	if err != nil {
		n.logger.Warn("Failed to subscribe commands")
		return
	}

	go func() {
		for {
			cmd, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if cmd == nil {
				n.logger.Warn("Disconnected the connection to the manager")
				break
			}
			if err != nil {
				n.logger.Warn("subscribeCommands(_) = _, %v", err)
			}
			n.ch <- cmd.Name
		}
	}()
}

func (n Node) heartbeat() {
	req := &pb.Empty{}
	n.cl.HeartBeat(
		context.Background(),
		req,
	)
}

func (n Node) sendStatus(status map[string]int) {
	req := &pb.AppsStatus{
		Status: make(map[string]*pb.AppStatus),
	}
	for appName, st := range status {
		req.Status[appName] = &pb.AppStatus{Mode: (ares.AppStatusMode)(st)}
	}
	n.cl.NotifyAppsStatus(
		context.Background(),
		req,
	)
}
