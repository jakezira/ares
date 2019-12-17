package ares

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"local/aurora/aurora/monitor"
	"local/util/cmd"
	alog "local/util/log"

	pb "local/ares/protos"

	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"google.golang.org/grpc"
)

const (
	version    = "0.1.0"
	configName = "config.json"
)

//
// Application Aurora application.
//
type Application struct {
	config Configuration

	server *Server

	lis        net.Listener
	grpcServer *grpc.Server

	logger  alog.Logger
	monitor *monitor.MonitorServer

	cmdList chan string
	cmdDic  *cmd.Dictionary
	quit    *quitInformation
}

type quitInformation struct {
	ch        chan bool
	waitGroup *sync.WaitGroup
}

func newQuitInformation() *quitInformation {
	return &quitInformation{
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{}}
}

func (q *quitInformation) quit() {
	close(q.ch)
}

func (q *quitInformation) wait() {
	q.waitGroup.Wait()
}

func (q *quitInformation) add() {
	q.waitGroup.Add(1)
}

func (q *quitInformation) done() {
	q.waitGroup.Done()
}

// NewApp create a new application.
func NewApp() *Application {
	app := Application{
		cmdList: make(chan string),
		quit:    newQuitInformation()}
	return &app
}

// Init the application.
func (app *Application) Init() {
	app.initLog()

	app.logger.Info("Starting Ares Manager %v ...", version)

	var err error
	app.config, err = LoadConfig(configName)
	if err != nil {
		app.logger.Error("Failed to load config file; %v", err)
		os.Exit(1)
	}

	app.setLogMode()

	app.lis, err = net.Listen("tcp", ":"+app.config.Port)
	if err != nil {
		app.logger.Error("Failed to listen: %v", err)
		os.Exit(1)
	}

	// Tracer
	tracer := alog.NewTracer(
		app.config.ZipkinAddr,
		"Ares",
		fmt.Sprintf("localhost:%s", app.config.Port))

	if tracer != nil {
		app.grpcServer = grpc.NewServer(
			grpc.StatsHandler(
				zipkingrpc.NewServerHandler(
					tracer.GetTracer(),
				),
			),
		)
	} else {
		app.grpcServer = grpc.NewServer()
	}

	server := NewServer(
		app.quit,
		app.logger,
		tracer)
	defer server.Close()

	pb.RegisterAresServiceServer(app.grpcServer, server)

	if app.monitor != nil {
		app.monitor.Run(app.config.Monitor)
	}

	app.initCommandDictionary()

	server.Init()
	app.server = server

	app.logger.Info("Ares Manager has been started successfully!")
}

// Run the application.
func (app *Application) Run() {

	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Recovered: %v", r)
		}
	}()

	app.server.start()
	go app.runService()
	go app.cmdInput()
	go app.cmdProc()

	time.Sleep(time.Second * 1)
	app.quit.wait()
}

// Close the application.
func (app *Application) Close() {
	app.logger.Close()
}

// GetServer returns aurora engine
func (app *Application) GetServer() *Server {
	return app.server
}

//
func (app *Application) initLog() {
	var err error
	//create your file with desired read/write permissions
	app.logger, err = alog.New(nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *Application) setLogMode() {
	logConf := app.config.Debug
	if logConf == "on" {
		app.logger.SetLevel(alog.Trace)
	}
}

func (app *Application) runService() {
	if err := app.grpcServer.Serve(app.lis); err != nil {
		app.logger.Warn("failed to serve: %v", err)
		app.quit.quit()
	}
}

func (app *Application) cmdInput() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please type a command to run.")

	for i := 0; ; i++ {
		text, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		text = text[0 : len(text)-2]
		app.cmdList <- text
	}
}

func (app *Application) cmdProc() {
	app.quit.add()
	defer app.quit.done()

	for {
		select {
		case cmd := <-app.cmdList:
			result := app.runCommand(cmd)
			fmt.Print(result)
		case <-app.quit.ch:
			return
		}
	}

}

func (app *Application) runCommand(cmd string) string {

	return app.cmdDic.Run(cmd)
}

func (app *Application) initCommandDictionary() {
	app.cmdDic = cmd.NewDictionary()
	app.cmdDic.RegisterCommand("quit",
		&cmd.ItemImpl{
			Help: "Quit the program.",
			RunFunc: func([]string) (string, bool) {
				ret := ""
				app.quit.quit()
				return ret, true
			}})
	app.cmdDic.RegisterCommand("ln",
		&cmd.ItemImpl{
			Help: "List nodes.",
			RunFunc: func([]string) (string, bool) {
				ret := ""
				ret = app.server.showNodes()
				return ret, true
			}})
	app.cmdDic.RegisterCommand("la",
		&cmd.ItemImpl{
			Help: "List apps.",
			RunFunc: func([]string) (string, bool) {
				ret := ""
				ret = app.server.showApps()
				return ret, true
			}})
	app.cmdDic.RegisterCommand("on",
		&cmd.ItemImpl{
			Help: "Turn on app.",
			RunFunc: func(args []string) (string, bool) {
				ret := ""
				if len(args) < 2 {
					return "input app name.", false
				}
				appName := args[1]
				ret = app.server.reqTurnOnApp(appName)
				return ret, true
			}})
}

type connectFuncType func(conn chan bool, failed chan bool, abort chan bool)

// reconnect using connectFunc.
// try to reconnect every 5 seconds for `reconnectLimit` count.
// returns true if connect success, false otherwise
func reconnect(connectFunc connectFuncType, reconnectLimit int) bool {

	var connectChan = make(chan bool)
	var failedChannel = make(chan bool)
	var abortChannel = make(chan bool)

	reconnectCount := 0

	for true {
		go connectFunc(connectChan, failedChannel, abortChannel)

		select {
		case <-connectChan:
			return true
		case <-abortChannel:
			return false
		case <-failedChannel:
			reconnectCount++
			if reconnectCount >= reconnectLimit {
				return false
			} else {
				time.Sleep(time.Second * 5)
			}
		}
	}
	return false
}
