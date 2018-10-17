package factory

import (
	"rela_recommend/cache"
	"strings"
	"rela_recommend/conf"
	"rela_recommend/log"
	"rela_recommend/utils"
	"rela_recommend/cache/redis"

	"github.com/gocql/gocql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// mysql slave
var DbR *gorm.DB
// mysql master
var DbW *gorm.DB

// redis cache
var CacheRds cache.Cache

// cassandra client
var CassandraClient *gocql.Session

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
	initCache(cfg)
	initCassandraSession(cfg)
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

	DbR, err = gorm.Open("mysql", cfg.Rdb.SlaveAddr)
	if err != nil {
		log.Error(err.Error())
	}
	DbR.DB().SetMaxIdleConns(1)
	DbR.DB().SetMaxOpenConns(10)
	if cfg.LogLevel == "debug" {
		DbR.LogMode(true)
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
	CacheRds, err = redis.NewRedisCache(cfg.Rds.RedisAddr, "", 0)
	if err != nil {
		log.Error(err.Error())
	}
}

func Close() {
	DbW.Close()
	DbR.Close()
	CassandraClient.Close()
}