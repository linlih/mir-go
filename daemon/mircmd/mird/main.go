// Package main
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/24 9:13 AM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package main

import (
	"fmt"
	"github.com/takama/daemon"
	"log"
	common2 "minlib/common"
	"mir-go/daemon/common"
	mir "mir-go/daemon/mircmd"
	"os"
)

const (
	name                  = "mird"                           // 服务的名字
	description           = "Multi-Identifier Router"        // 服务描述
	defaultConfigFilePath = "/usr/local/etc/mir/mirconf.ini" // MIR配置文件路径
)

// dependencies that are NOT required by the service, but might be used
var dependencies = []string{"dummy.service"}

var stdlog, errlog *log.Logger

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: myservice install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	mirConfig, err := common.ParseConfig(defaultConfigFilePath)
	if err != nil {
		common2.LogFatal(err)
	}

	starter := mir.NewMIRStarter(mirConfig)
	return starter.Start("123456")
}
func init() {
	stdlog = log.New(os.Stdout, "", 0)
	errlog = log.New(os.Stderr, "", 0)
}

func main() {
	srv, err := daemon.New(name, description, daemon.GlobalDaemon, dependencies...)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
