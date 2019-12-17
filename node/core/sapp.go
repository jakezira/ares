package ares

import (
	alog "local/util/log"
	"os/exec"
	"time"

	"github.com/shirou/gopsutil/process"
)

type SApp struct {
	name   string
	run    string
	dir    string
	logger alog.Logger

	cmd      *exec.Cmd
	startCh  chan error
	finishCh chan error

	status int
}

var (
	AppStatus_On   = 0
	AppStatus_Fail = 1
	AppStatus_Off  = 2
)

func newSApp(name, run, dir string, logger alog.Logger) *SApp {
	app := &SApp{
		name:   name,
		run:    run,
		dir:    dir,
		logger: logger,

		startCh:  make(chan error, 1),
		finishCh: make(chan error, 1),

		status: AppStatus_Off,
	}

	go app.monitorStart()
	go app.monitorFinish()

	return app
}

func (a *SApp) runApp() {
	cmd := exec.Command("cmd", "/C", "start", a.run)
	cmd.Dir = a.dir

	if err := cmd.Start(); err != nil {
		a.logger.Error("failed to run process: %s", a.name)
		a.status = AppStatus_Fail
		return
	}

	a.cmd = cmd

	a.startCh <- cmd.Wait()

	go func() {
		loop := false
		for {
			if loop {
				break
			}
			select {
			case <-time.After(1 * time.Second):
				processes, _ := process.Processes()
				for _, p := range processes {
					pid, _ := p.Ppid()
					_ = pid
					/*if pid == int32(a.cmd.Process.Pid) {
						_, err := p.Status()
						a.logger.Info("Process status: %s", err.Error())
						// loop = true
						// a.done <- fmt.Errorf("%v", "crashed")
					}*/

				}
				// isExist, _ := process.PidExists(int32(a.cmd.Process.Pid))
				// if !isExist {
				// }
			}
		}
	}()
}

func (a *SApp) killApp() {
	if a.cmd != nil {
		if err := a.cmd.Process.Kill(); err != nil {
			a.logger.Error("failed to kill process: %s", err.Error())
		}
	}
}

func (a *SApp) monitorStart() {
	for {
		select {
		case err := <-a.startCh:
			if err != nil {
				a.logger.Info("app %s failed to start with error = %s", a.name, err.Error())
				a.status = AppStatus_Fail
			} else {
				a.logger.Info("app %s started successfully", a.name)
				a.status = AppStatus_On
				pid := a.findAppProcessId()
				_ = pid

			}
		}
	}
}

func (a *SApp) monitorFinish() {
	for {
		select {
		case err := <-a.finishCh:
			if err != nil {
				a.logger.Info("app %s finished with error = %s", a.name, err.Error())
				a.status = AppStatus_Fail

			} else {
				a.logger.Info("app %s finished successfully", a.name)
				a.status = AppStatus_On

			}
		}
	}
}

func (a *SApp) findAppProcessId() int {
	processes, _ := process.Processes()
	for _, p := range processes {
		pid, _ := p.Ppid()
		//if
		//p.Name()
		/*if pid == int32(a.cmd.Process.Pid) {
			_, err := p.Status()
			a.logger.Info("Process status: %s", err.Error())
			// loop = true
			// a.done <- fmt.Errorf("%v", "crashed")
		}*/
		_ = pid
	}
	return 0
}

func (a *SApp) waitFinish(pid int) {
	// var waitmsg syscall.Waitmsg

	// if p.Pid == -1 {
	// 	return nil, ErrInvalid
	// }

	// err := syscall.WaitProcess(p.Pid, &waitmsg)
	// if err != nil {
	// 	return nil, NewSyscallError("wait", err)
	// }

	//a.finishCh <- err
}
