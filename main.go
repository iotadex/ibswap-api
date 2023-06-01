package main

import (
	"ibswap/api"
	"ibswap/daemon"
	"ibswap/gl"
	"ibswap/model"
	"ibswap/service"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-d")
	}

	daemon.Background("./out.log", true)

	gl.CreateLogFiles()

	model.ConnectToMysql()

	api.StartHttpServer()

	service.Start()

	daemon.WaitForKill()

	api.StopHttpServer()
}
