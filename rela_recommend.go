package main

import (
	"flag"
	"os"
	"os/signal"
	"rela_recommend/conf"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/routers"
	"rela_recommend/routes"
	"runtime"
	"syscall"
)

var (
	buildTime  string
	configFile = flag.String("conf", "", "param config file")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	flag.Parse()
	var err error
	var cfg *conf.Config
	if len(*configFile) == 0 {
		log.Error("no config set, please set the config")
		return
	}

	cfg, err = conf.NewConfigWithFile(*configFile)
	if err != nil {
		log.Error(err.Error())
		return
	}

	log.Infof("built on:%s", buildTime)
	log.Infof("%+v", *cfg)

	factory.Init(cfg)

	router := routers.Default()
	routes.RegisterRouters(router)
	routes.RegisterHandler(router)

	apiServer := routes.NewApiServer(cfg.Port, router)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go apiServer.Run()
	<-sc

	apiServer.Close()
	factory.Close()
	log.Info("rela_recommend is closed.")
}
