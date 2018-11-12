package redisCluster

import (
	redisCluster "github.com/chasex/redis-go-cluster"
)

type RedisCluster struct {
	Rc    *redisCluster.Cluster
	batch *redisCluster.Batch
	data  []interface{}
}

func (rc *RedisCluster) Close() error {
	//rc.Rc.Close()
	return nil
}

func (rc *RedisCluster) Err() error {
	return nil
}

func (rc *RedisCluster) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return rc.Rc.Do(commandName, args...)
}

func (rc *RedisCluster) Send(commandName string, args ...interface{}) error {
	rc.NewBatch()
	//调用send时清空 receiver
	var da []interface{}
	rc.data = da
	rc.batch.Put(commandName, args...)
	return nil
}
func (rc *RedisCluster) Flush() error {
	data, err := rc.Rc.RunBatch(rc.batch)
	rc.data = data
	rc.batch = nil
	return err
}

func (rc *RedisCluster) Receive() (reply interface{}, err error) {
	return rc.data, nil
}

func (rc *RedisCluster) NewBatch() *redisCluster.Batch {
	if rc.batch == nil {
		rc.batch = rc.Rc.NewBatch()
	}
	return rc.batch
}

func (rc *RedisCluster) GetCluster() *redisCluster.Cluster {
	return rc.Rc
}
