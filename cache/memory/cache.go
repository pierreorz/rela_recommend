package memory

import (
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"rela_recommend/cache"
	"rela_recommend/utils"
)

// github.com/coocood/freecache
type Cache struct {
	locks    [256]sync.Mutex
	segments [256]segment
}

func hashFunc(data []byte) uint64 {
	return utils.Md5Uint64(data)
}

var errNotByteArray = errors.New("not []byte")

var DefaultSize = 128 * 1024 * 1024

// The cache size will be set to 512KB at minimum.
// If the size is set relatively large, you should call
// `debug.SetGCPercent()`, set it to a much smaller value
// to limit the memory consumption and GC pause time.
func NewMemoryCache(size int) (cache.Cache, error) {
	var cache Cache
	for i := 0; i < 256; i++ {
		cache.segments[i] = newSegment(size/256, i)
	}
	return &cache, nil
}

func (cache *Cache) Set(key string, value interface{}) error {
	return cache.setEx(key, value, -1)
}

func (cache *Cache) SetEx(key string, value interface{}, expireSeconds int) error {
	return cache.setEx(key, value, expireSeconds)
}

// If the key is larger than 65535 or value is larger than 1/1024 of the cache size,
// the entry will not be written to the cache. expireSeconds <= 0 means no expire,
// but it can be evicted when cache is full.
func (cache *Cache) setEx(key string, value interface{}, expireSeconds int) (err error) {
	byteKey := []byte(key)
	var byteValue []byte
	switch result := value.(type) {
	case []byte:
		byteValue = result
	case string:
		byteValue = []byte(result)
	default:
		return errNotByteArray
	}
	hashVal := hashFunc(byteKey)
	segId := hashVal & 255
	cache.locks[segId].Lock()
	err = cache.segments[segId].set(byteKey, byteValue, hashVal, expireSeconds)
	cache.locks[segId].Unlock()
	return
}

// Get the value or not found error.
func (cache *Cache) Get(key string) (interface{}, error) {
	byteKey := []byte(key)
	return cache.get(byteKey)
}

// Get the value or not found error.
func (cache *Cache) get(key []byte) ([]byte, error) {
	hashVal := hashFunc(key)
	segId := hashVal & 255
	cache.locks[segId].Lock()
	value, err := cache.segments[segId].get(key, hashVal)
	cache.locks[segId].Unlock()
	return value, err
}

func (cache *Cache) Del(key string) error {
	byteKey := []byte(key)
	cache.delete(byteKey)
	return nil
}

func (cache *Cache) delete(key []byte) (affected bool) {
	hashVal := hashFunc(key)
	segId := hashVal & 255
	cache.locks[segId].Lock()
	affected = cache.segments[segId].del(key, hashVal)
	cache.locks[segId].Unlock()
	return
}

// GetMulti is a batch version of Get.
func (cache *Cache) Mget(keys []string) ([]interface{}, error) {
	return nil, nil
}

// increase cached int value by key, as a counter.
func (cache *Cache) Incr(key string) (int64, error) {
	return 0, nil
}

// IncrBy increase counter in redis.
func (cache *Cache) IncrBy(key string, count int) (int64, error) {
	return 0, nil
}

// decrease cached int value by key, as a counter.
func (cache *Cache) Decr(key string) (int64, error) {
	return 0, nil
}

// check if cached value exists or not.
func (cache *Cache) Exists(key string) bool {
	return false
}

// get conn
func (cache *Cache) GetConn() interface{} {
	return nil
}

func (cache *Cache) SetInt(key int64, value []byte, expireSeconds int) (err error) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return cache.setEx(string(bKey[:]), value, expireSeconds)
}

func (cache *Cache) GetInt(key int64) (value []byte, err error) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return cache.get(bKey[:])
}

func (cache *Cache) DelInt(key int64) (affected bool) {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], uint64(key))
	return cache.delete(bKey[:])
}

func (cache *Cache) EvacuateCount() (count int64) {
	for i := 0; i < 256; i++ {
		count += atomic.LoadInt64(&cache.segments[i].totalEvacuate)
	}
	return
}

func (cache *Cache) ExpiredCount() (count int64) {
	for i := 0; i < 256; i++ {
		count += atomic.LoadInt64(&cache.segments[i].totalExpired)
	}
	return
}

func (cache *Cache) EntryCount() (entryCount int64) {
	for i := 0; i < 256; i++ {
		entryCount += atomic.LoadInt64(&cache.segments[i].entryCount)
	}
	return
}

// The average unix timestamp when a entry being accessed.
// Entries have greater access time will be evacuated when it
// is about to be overwritten by new value.
func (cache *Cache) AverageAccessTime() int64 {
	var entryCount, totalTime int64
	for i := 0; i < 256; i++ {
		totalTime += atomic.LoadInt64(&cache.segments[i].totalTime)
		entryCount += atomic.LoadInt64(&cache.segments[i].totalCount)
	}
	if entryCount == 0 {
		return 0
	} else {
		return totalTime / entryCount
	}
}

func (cache *Cache) OverwriteCount() (overwriteCount int64) {
	for i := 0; i < 256; i++ {
		overwriteCount += atomic.LoadInt64(&cache.segments[i].overwrites)
	}
	return
}

func (cache *Cache) LRange(key string, start int, end int) ([][]byte, error) {
	return nil, nil
 }

func (cache *Cache) MsetEx(keyValMap map[string]interface{}, expire int64) error {
	return nil
 }

 func (cache *Cache) SMembers(key string) ([]interface{}, error) {
	return nil, nil
 }

func (cache *Cache) Close() { }
