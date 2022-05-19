package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

const (
	DefaultProxyAddr       = "127.0.0.1:80"
	DefaultMatchAddr       = "127.0.0.1:80"
	DefaultNsqAddr         = "127.0.0.1:80"
	DefaultMemoryCacheSize = 1024 * 1024 * 128
	DefaultLogLevel        = "debug"
	DefaultLiveRpcAddr     = "http://live:3500"
)

var (
	port         = flag.Int("port", 3100, "监听端口")
	isProduction = flag.Bool("isProduction", false, "是否是生产环境")
	workerId     = flag.Int("workerId", 0, " appid 表示无状态app的唯一性,每个无状态app的值应该不相同")
)

type rdbConfig struct {
	MasterAddr string `toml:"master_address"`
	SlaveAddr  string `toml:"slave_address"`
	WinkAddr   string `toml:"wink_address"`
	LiveAddr   string `toml:"live_address"`
}

type rdsConfig struct {
	RedisAddr          string `toml:"redis_addr"`
	RedisLiveAddr      string `toml:"redis_live_addr"`
	RedisInternalAddr  string `toml:"redis_internal_addr"`
	ClusterAddr        string `toml:"redis_cluster_addr"`
	LedisAddr          string `toml:"ledis_addr"`
	PikaAddr           string `toml:"pika_addr"`
	LedisViewAddr      string `toml:"ledis_view_addr"`
	LedisDataAddr      string `toml:"ledis_data_addr"`
	RedisComAddr       string `toml:"redis_com_addr"`
	RedisPushAddr      string `toml:"redis_push_addr"`
	BehaviorAddr       string `toml:"behavior_addr"`
	BehaviorBackupAddr string `toml:"behavior_backup_addr"`
	AwsRedisAddr       string `toml:"aws_redis_addr"`
}

type cassandraConfig struct {
	Addresses string `toml:"addresses"`
}

type rpcConfig struct {
	SearchRpcAddr   string `toml:"search_rpc_addr"`
	ApiRpcAddr      string `toml:"api_rpc_addr"`
	ChatRoomRpcAddr string `toml:"chatroom_rpc_addr"`
	LiveRpcAddr     string `toml:"live_rpc_addr"`
	AiSearchRpcAddr string `toml:"ai_search_rpc_addr"`
}

type influxdbConfig struct {
	Addr   string `toml:"addr"`
	Token  string `toml:"token"`
	Org    string `toml:"org"`
	Bucket string `toml:"bucket"`
}

type Config struct {
	FileName string `toml:"-"`

	WorkerId        int8   `toml:"-"`
	IsProduction    bool   `toml:"-"`
	Port            int    `toml:"port"`
	LogLevel        string `toml:"log_level"`
	WebApiAddr      string `toml:"web_api_addr"`
	NewWebApiAddr   string `toml:"new_web_api_addr"`
	PorxyAddr       string `toml:"porxy_addr"`
	FmProxyAddress  string `toml:"fm_proxy_addr"`
	GrpcPort        int    `toml:"grpc_port"`
	MemoryCacheSize int    `toml:"memory_cache_size"`

	ElasticAddr           string `toml:"elastic_addr"`
	LiveClusterMongoAddr  string `toml:"live_cluster_mongo_addr"`
	ClusterMongoAddr      string `toml:"cluster_mongo_addr"`
	MatchClusterMongoAddr string `toml:"match_cluster_mongo_addr"`
	MatchAddr             string `toml:"match_addr"`
	NsqAddr               string `toml:"nsq_addr"`
	InternalAddr          string `toml:"internal_addr"`

	WebBase string `toml:"web_base"`

	Rdb       rdbConfig       `toml:"rdb"`
	Rds       rdsConfig       `toml:"rds"`
	Cassandra cassandraConfig `toml:"cassandra"`
	Rpc       rpcConfig       `toml:"rpc"`
	Influxdb  influxdbConfig  `toml:"influxdb"`
}

func NewConfigWithFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if cfg, err := NewConfigWithData(data); err != nil {
		return nil, err
	} else {
		cfg.FileName = fileName
		return cfg, nil
	}
}

func NewConfigWithData(data []byte) (*Config, error) {
	cfg := NewConfigDefault()

	_, err := toml.Decode(string(data), cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewConfigDefault() *Config {
	cfg := new(Config)

	cfg.Port = *port
	cfg.IsProduction = *isProduction
	cfg.WorkerId = int8(*workerId)

	cfg.LogLevel = DefaultLogLevel
	cfg.PorxyAddr = DefaultProxyAddr
	cfg.MemoryCacheSize = DefaultMemoryCacheSize

	cfg.MatchAddr = DefaultMatchAddr
	cfg.NsqAddr = DefaultNsqAddr

	// rpc
	cfg.Rpc.LiveRpcAddr = DefaultLiveRpcAddr
	return cfg
}
