package factory

import (
	"rela_recommend/cache"
	"rela_recommend/cache/memory"
	cacheUtils "rela_recommend/cache/utils"
	"rela_recommend/conf"
	"rela_recommend/log"
	"rela_recommend/rpc"
	"rela_recommend/utils"
	"strings"
	"time"

	"gopkg.in/mgo.v2"

	"rela_recommend/service/segment"

	"github.com/gocql/gocql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// "rela_recommend/service/abtest"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// mysql slave
var DbR *gorm.DB

// mysql master
var DbW *gorm.DB

// redis cache
var CacheRds cache.Cache

var CacheLiveRds cache.Cache

// redis cache
var CacheCluster cache.Cache

// 新的阿里flink双写的redis
var CacheBehaviorRdsBackup cache.Cache

//本地cache
var CacheLoc cache.Cache

// pika
var PikaCluster cache.Cache

// cassandra client
var CassandraClient *gocql.Session

// match mongo
var MatchClusterMon *mgo.Session

var SearchRpcClient *rpc.HttpClient
var ApiRpcClient *rpc.HttpClient
var ChatRoomRpcClient *rpc.HttpClient
var LiveRpcClient *rpc.HttpClient
var AiSearchRpcClient *rpc.HttpClient

// influxdb
var InfluxdbClient influxdb2.Client

// 分词
var Segmenter segment.ISegmenter

//配置相关
var Cfg *conf.Config

var IsProduction bool

func Init(cfg *conf.Config) {
	Cfg = cfg
	//设置日志打印等级
	log.SetLevelByName(cfg.LogLevel)

	//只要本地时间不是东八区程序就退出, 后续需要动态根据时区进行修改
	if 8*60*60 != utils.GetLocalTimeZoneOffset() {
		panic("not utc+8, then going to quit")
	}
	initDB(cfg)
	initConsul(cfg)
	initCache(cfg)
	// initMongo(cfg)
	// initCassandraSession(cfg)
	initRpc(cfg)
	initSegmenter(cfg)
}

func initConsul(cfg *conf.Config) {
	// abtest.BeginWatching("127.0.0.1:8500")
}

func initDB(cfg *conf.Config) {
	// var err error
	// DbW, err = gorm.Open("mysql", cfg.Rdb.MasterAddr)
	// if err != nil {
	// 	log.Error(err.Error())
	// }
	// DbW.DB().SetMaxIdleConns(1)
	// DbW.DB().SetMaxOpenConns(10)
	// if cfg.LogLevel == "debug" {
	// 	DbW.LogMode(true)
	// }
	//sqlCreate(DbW)

	// DbR, err = gorm.Open("mysql", cfg.Rdb.SlaveAddr)
	// if err != nil {
	// 	log.Error(err.Error())
	// }
	// DbR.DB().SetMaxIdleConns(1)
	// DbR.DB().SetMaxOpenConns(10)
	// if cfg.LogLevel == "debug" {
	// 	DbR.LogMode(true)
	// }
	if len(cfg.Influxdb.Addr) > 0 && len(cfg.Influxdb.Bucket) > 0 {
		InfluxdbClient = influxdb2.NewClient(cfg.Influxdb.Addr, cfg.Influxdb.Token)
		log.Infof("INIT InfluxdbClient: %s ....", cfg.Influxdb.Addr)
	}
}

func initIsProduction(cfg *conf.Config) {
	IsProduction = cfg.IsProduction
}

func initCassandraSession(cfg *conf.Config) {
	server := cfg.Cassandra.Addresses
	var sevs = strings.Split(server, ",")
	cluster := gocql.NewCluster(sevs...)
	cluster.Keyspace = "rela_db"
	cluster.Consistency = gocql.Quorum
	cluster.NumConns = 100

	session, err := cluster.CreateSession()

	if err != nil {
		log.Error(err)
	}

	CassandraClient = session
}

func initCache(cfg *conf.Config) {
	var err error
	CacheRds, err = cacheUtils.NewRedisOrClusterCache(cfg.Rds.RedisAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	CacheLiveRds, err = cacheUtils.NewRedisOrClusterCache(cfg.Rds.RedisLiveAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT ClusterAddr: %s ....", cfg.Rds.ClusterAddr)
	CacheCluster, err = cacheUtils.NewRedisOrClusterCache(cfg.Rds.ClusterAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT PikaAddr: %s ....", cfg.Rds.PikaAddr)
	PikaCluster, err = cacheUtils.NewRedisOrClusterCache(cfg.Rds.PikaAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT ClusterAddr: %s ....", cfg.Rds.ClusterAddr)
	CacheBehaviorRdsBackup, err = cacheUtils.NewRedisOrClusterCache(cfg.Rds.BehaviorBackupAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	memoryCacheSize := 1024 * 1024 * 1024 // 1GB
	log.Infof("INIT CacheLocal: %d ....", memoryCacheSize)
	CacheLoc, err = memory.NewMemoryCache(memoryCacheSize)
	if err != nil {
		log.Error(err.Error())
	}
}

func initMongo(cfg *conf.Config) {
	var err error
	MatchClusterMon, err = mgo.Dial(cfg.MatchClusterMongoAddr)
	MatchClusterMon.SetPoolLimit(100)
	if err != nil {
		log.Error(err.Error())
	}
}

func initRpc(cfg *conf.Config) {
	SearchRpcClient = rpc.NewHttpClient(cfg.Rpc.SearchRpcAddr, time.Millisecond*500)
	ApiRpcClient = rpc.NewHttpClient(cfg.Rpc.ApiRpcAddr, time.Millisecond*100)
	ChatRoomRpcClient = rpc.NewHttpClient(cfg.Rpc.ChatRoomRpcAddr, time.Millisecond*1000)
	AiSearchRpcClient = rpc.NewHttpClient(cfg.Rpc.AiSearchRpcAddr, time.Millisecond*1000)
	LiveRpcClient = rpc.NewHttpClient(cfg.Rpc.LiveRpcAddr, time.Millisecond*700)
}

func initSegmenter(cfg *conf.Config) {
	Segmenter = segment.NewSegmenter()
	log.Infof("INIT Segmenter: %s", Segmenter.Cut("你好分词已经准备好了！"))
}

func Close() {
	//close db
	// DbW.Close()
	// DbR.Close()

	//close cassandra
	// CassandraClient.Close()

	//close mgo
	// MatchClusterMon.Close()

	//close cache
	CacheRds.Close()
	CacheLiveRds.Close()
	CacheCluster.Close()
	PikaCluster.Close()
	CacheBehaviorRdsBackup.Close()
}
