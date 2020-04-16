package memory

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"rela_recommend/utils"
	"time"
)

func TestFreeCache(t *testing.T) {
	cache, err := NewMemoryCache(1024 * 1024 * 10)
	if err != nil {
		t.Error(err.Error())
	}

	key := "abcd"
	val := "efghijkl"
	err = cache.Set(key, val)
	if err != nil {
		t.Error("err should be nil")
	}
	value, err := cache.Get(key)
	if err != nil || utils.GetString(value) != val {
		t.Error("value not equal")
	}
	err = cache.Del(key)
	if err != nil {
		t.Error("del should return nil")
	}
	value, err = cache.Get(key)
	if err != ErrNotFound {
		t.Error("error should be ErrNotFound after being deleted")
	}
	err = cache.Del(key)
	if err != nil {
		t.Error("del should not return nil")
	}

	n := 500
	for i := 0; i < n; i++ {
		keyStr := fmt.Sprintf("key%v", i)
		valStr := strings.Repeat(keyStr, 10)
		err = cache.SetEx(keyStr, valStr, 0)
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Second)
	for i := 1; i < n; i += 2 {
		keyStr := fmt.Sprintf("key%v", i)
		cache.Get(keyStr)
	}

	for i := 1; i < n; i += 8 {
		keyStr := fmt.Sprintf("key%v", i)
		cache.Del(keyStr)
	}

	for i := 0; i < n; i += 2 {
		keyStr := fmt.Sprintf("key%v", i)
		valStr := strings.Repeat(keyStr, 10)
		err = cache.SetEx(keyStr, valStr, 0)
		if err != nil {
			t.Error(err)
		}
	}
	for i := 1; i < n; i += 2 {
		keyStr := fmt.Sprintf("key%v", i)
		expectedValStr := strings.Repeat(keyStr, 10)
		value, err = cache.Get(keyStr)
		if err == nil && utils.GetString(value) != expectedValStr {
			t.Errorf("value is %v, expected %v", value, expectedValStr)
		}
	}
}

func TestOverwrite(t *testing.T) {
	cache, err := NewMemoryCache(1024 * 1024 * 10)
	if err != nil {
		t.Error(err.Error())
	}
	conn := cache.GetConn()
	memcache, _ := conn.(*Cache)
	key := "abcd"
	var val []byte
	cache.Set(key, val)
	val = []byte("efgh")
	cache.Set(key, val)
	val = append(val, 'i')
	cache.Set(key, val)
	if count := memcache.OverwriteCount(); count != 0 {
		t.Error("overwrite count is", count, "expected ", 0)
	}
	res, _ := cache.Get(key)
	if utils.GetString(res) != string(val) {
		t.Error(res)
	}
	val = append(val, 'j')
	cache.Set(key, val)
	res, _ = cache.Get(key)
	if utils.GetString(res) != string(val) {
		t.Error(res, "aaa")
	}
	val = append(val, 'k')
	cache.Set(key, val)
	res, _ = cache.Get(key)
	if utils.GetString(res) != "efghijk" {
		t.Error(res)
	}
	val = append(val, 'l')
	cache.Set(key, val)
	res, _ = cache.Get(key)
	if utils.GetString(res) != "efghijkl" {
		t.Error(res)
	}
	val = append(val, 'm')
	cache.Set(key, val)
	if count := memcache.OverwriteCount(); count != 3 {
		t.Error("overwrite count is", count, "expected ", 3)
	}

}

func TestExpire(t *testing.T) {
	cache, err := NewMemoryCache(1024 * 1024 * 10)
	if err != nil {
		t.Error(err.Error())
	}
	key := "abcd"
	val := "efgh"
	err = cache.SetEx(key, val, 1)
	if err != nil {
		t.Error("err should be nil")
	}
	time.Sleep(time.Second)
	_, err = cache.Get(key)
	if err == nil {
		t.Fatal("key should be expired", val)
	}
}

func TestInt64Key(t *testing.T) {
	cache, err := NewMemoryCache(1024 * 1024)
	if err != nil {
		t.Error(err.Error())
	}
	conn := cache.GetConn()
	memcache, _ := conn.(*Cache)
	err = memcache.SetInt(1, []byte("abc"), 0)
	if err != nil {
		t.Error("err should be nil")
	}
	err = memcache.SetInt(2, []byte("cde"), 0)
	if err != nil {
		t.Error("err should be nil")
	}
	val, err := memcache.GetInt(1)
	if err != nil {
		t.Error("err should be nil")
	}
	if !bytes.Equal(val, []byte("abc")) {
		t.Error("value not equal")
	}
	affected := memcache.DelInt(1)
	if !affected {
		t.Error("del should return affected true")
	}
	_, err = memcache.GetInt(1)
	if err != ErrNotFound {
		t.Error("error should be ErrNotFound after being deleted")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache, err := NewMemoryCache(1024 * 1024 * 256)
	if err != nil {
		b.Error(err.Error())
	}
	var key [8]byte
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		cache.Set(string(key[:]), make([]byte, 8))
	}
}

func BenchmarkMapSet(b *testing.B) {
	m := make(map[string][]byte)
	var key [8]byte
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		m[string(key[:])] = make([]byte, 8)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数
	cache, err := NewMemoryCache(1024 * 1024 * 256)
	if err != nil {
		b.Error(err.Error())
	}
	var key [8]byte
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		cache.Set(string(key[:]), make([]byte, 8))
	}
	//做一些初始化的工作,例如读取文件数据,数据库连接之类的,
	//这样这些时间不影响我们测试函数本身的性能
	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		cache.Get(string(key[:]))
	}
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[string][]byte)
	var key [8]byte
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		m[string(key[:])] = make([]byte, 8)
	}
	b.StartTimer()
	var hitCount int64
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(key[:], uint64(i))
		if m[string(key[:])] != nil {
			hitCount++
		}
	}
}

func BenchmarkHashFunc(b *testing.B) {
	key := make([]byte, 8)
	rand.Read(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashFunc(key)
	}
}
