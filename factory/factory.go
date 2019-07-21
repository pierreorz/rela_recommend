package factory

import (
	"gopkg.in/mgo.v2"
	"rela_recommend/cache"
	"rela_recommend/cache/redis"
	"rela_recommend/rpc"
	"rela_recommend/conf"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"rela_recommend/cache/redisCluster"
	"rela_recommend/service/segment"
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
// Behavior cache
var CacheBehaviorRds cache.Cache

// pika
var PikaCluster cache.Cache

// cassandra client
var CassandraClient *gocql.Session

// match mongo
var MatchClusterMon *mgo.Session

var SearchRpcClient *rpc.HttpClient
var ApiRpcClient *rpc.HttpClient

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
	// initDB(cfg)
	initCache(cfg)
	// initMongo(cfg)
	// initCassandraSession(cfg)
	initRpc(cfg)
	initSegmenter(cfg)
}

func initDB(cfg *conf.Config) {
	var err error
	DbW, err = gorm.Open("mysql", cfg.Rdb.MasterAddr)
	if err != nil {
		log.Error(err.Error())
	}
	DbW.DB().SetMaxIdleConns(1)
	DbW.DB().SetMaxOpenConns(10)
	if cfg.LogLevel == "debug" {
		DbW.LogMode(true)
	}
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
	CacheRds, err = redisCluster.NewRedisCache(cfg.Rds.RedisAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	CacheLiveRds, err = redis.NewRedisCache(cfg.Rds.RedisLiveAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT ClusterAddr: %s ....", cfg.Rds.ClusterAddr)
	CacheCluster, err = redis.NewRedisCache(cfg.Rds.ClusterAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT PikaAddr: %s ....", cfg.Rds.PikaAddr)
	PikaCluster, err = redis.NewRedisCache(cfg.Rds.PikaAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}

	log.Infof("INIT ClusterAddr: %s ....", cfg.Rds.ClusterAddr)
	CacheBehaviorRds, err = redis.NewRedisCache(cfg.Rds.BehaviorAddr, "", 0)
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

func initRpc(cfg *conf.Config){
	SearchRpcClient = rpc.NewHttpClient(cfg.Rpc.SearchRpcAddr, time.Millisecond * 500)
	ApiRpcClient = rpc.NewHttpClient(cfg.Rpc.ApiRpcAddr, time.Millisecond * 100)
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
	CacheBehaviorRds.Close()
}
