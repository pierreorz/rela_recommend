package cache

type Cache interface {
	// get cached value by key.
	Get(key string) (interface{}, error)
	// GetMulti is a batch version of Get.
	Mget(keys []string) ([]interface{}, error)
	// lrange key start end
	LRange(key string, start int, end int) ([][]byte, error)
	ZRange(key string, start int, end int) ([]interface{}, error)

	SMembers(key string) ([]interface{}, error)
	// SetMulti is a batch version of Get.
	MsetEx(keyValMap map[string]interface{}, expire int64) error
	// set cached value with key
	Set(key string, val interface{}) error
	// set cached value with key and expire time.
	SetEx(key string, val interface{}, expireSeconds int) error
	// delete cached value by key.
	Del(key string) error
	// increase cached int value by key, as a counter.
	Incr(key string) (int64, error)

	IncrBy(key string, count int) (int64, error)
	// decrease cached int value by key, as a counter.
	Decr(key string) (int64, error)
	// check if cached value exists or not.
	Exists(key string) bool
	// get conn
	GetConn() interface{}
	//close
	Close()
}
