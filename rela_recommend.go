package main

import (
	"flag"
	"os"
	"os/signal"
	"rela_recommend/conf"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/routes"
	"rela_recommend/utils/routers"
	"runtime"
	"syscall"

	// "fmt"
	// "time"
	// "rela_recommend/algo/live"
	"rela_recommend/service/abtest"
	"rela_recommend/service/performs"
)

var (
	buildTime  string
	configFile = flag.String("conf", "", "param config file")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	flag.Parse()

	abtest.BeginWatching("127.0.0.1:8500")

	// features := []map[int]float32{map[int]float32{1:2019.000000,501:26.000000,502:170.000000,503:41.000000,504:891.000000,505:411.000000,507:14254.720000,506:1358.460000,509:916.000000,510:273.270000,10191:1.000000,10140:1.000000,10022:1.000000,10024:1.000000,10029:1.000000,10111:1.000000,10127:1.000000,10158:1.000000,10161:1.000000,10180:1.000000,10187:1.000000,10210:1.000000,10220:1.000000,10221:1.000000},
	// 	map[int]float32{1:100.000000,2:201.000000,3:80.000000,501:24.000000,502:201.000000,503:41.000000,504:825.000000,505:668.000000,507:383.470000,506:8.290000,509:3.000000,10191:1.000000,10140:1.000000,10018:1.000000,10024:1.000000,10109:1.000000,10120:1.000000,10127:1.000000,10152:1.000000,10161:1.000000,10180:1.000000,10187:1.000000,10210:1.000000,10220:1.000000,10230:1.000000},
	// 	map[int]float32{1:26.000000,2:168.000000,3:71.000000,501:27.000000,502:174.000000,503:57.000000,504:812.000000,505:595.000000,507:2281.170000,506:33.880000,509:171.000000,10190:1.000000,10142:1.000000,10006:1.000000,10027:1.000000,10106:1.000000,10123:1.000000,10128:1.000000,10153:1.000000,10173:1.000000,10177:1.000000,10187:1.000000,10206:1.000000,10211:1.000000,10221:1.000000},
	// 	map[int]float32{1:2019.000000,2:168.000000,501:25.000000,502:165.000000,503:45.000000,504:849.000000,505:510.000000,507:4760.000000,506:152.170000,509:84.000000,510:0.690000,10190:1.000000,10140:1.000000,10020:1.000000,10024:1.000000,10029:1.000000,10110:1.000000,10130:1.000000,10159:1.000000,10170:1.000000,10178:1.000000,10187:1.000000,10210:1.000000,10220:1.000000,10221:1.000000},
	// 	map[int]float32{1:20.000000,2:174.000000,3:62.000000,501:101.000000,502:140.000000,503:41.000000,504:898.000000,505:526.000000,509:93.000000,10195:1.000000,10140:1.000000,10018:1.000000,10024:1.000000,10104:1.000000,10122:1.000000,10130:1.000000,10159:1.000000,10161:1.000000,10177:1.000000,10187:1.000000,10206:1.000000,10211:1.000000,10221:1.000000 }}

	// model, _ := live.LiveAlgosMap["LiveModelV1_0"]
	// startTime := time.Now()
	// for j := 0; j < 10000; j++ {
	// 	for i, _ := range features {
	// 		ft := &utils.Features{}
	// 		ft.FromMap(features[i])
	// 		model.PredictSingle(ft)
	// 		// fmt.Printf("%d: %f\n", i, score)
	// 	}
	// }
	// endTime := time.Now()
	// fmt.Printf("%d: %f\n", 0, endTime.Sub(startTime).Seconds())
	// return

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

	performs.BeginWatching(cfg.Influxdb.Org, cfg.Influxdb.Bucket) // 监听写入influxdb

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
