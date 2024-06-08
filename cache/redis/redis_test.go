// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	bm, err := NewRedisCache("127.0.0.1:6379", "", 0)
	if err != nil {
		t.Error("init err")
	}
	timeoutDuration := 10
	if err = bm.SetEx("astaxie", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.Exists("astaxie") {
		t.Error("check err")
	}

	time.Sleep(11 * time.Second)

	if bm.Exists("astaxie") {
		t.Error("check err")
	}
	if err = bm.SetEx("astaxie", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}

	if v, _ := redis.Int(bm.Get("astaxie")); v != 1 {
		t.Error("get err")
	}

	if _, err = bm.Incr("astaxie"); err != nil {
		t.Error("Incr Error", err)
	}

	if v, _ := redis.Int(bm.Get("astaxie")); v != 2 {
		t.Error("get err")
	}

	if _, err = bm.Decr("astaxie"); err != nil {
		t.Error("Decr Error", err)
	}

	if v, _ := redis.Int(bm.Get("astaxie")); v != 1 {
		t.Error("get err")
	}
	bm.Del("astaxie")
	if bm.Exists("astaxie") {
		t.Error("delete err")
	}

	//test string
	if err = bm.SetEx("astaxie", "author", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.Exists("astaxie") {
		t.Error("check err")
	}

	if v, _ := redis.String(bm.Get("astaxie")); v != "author" {
		t.Error("get err")
	}

	//test GetMulti
	if err = bm.SetEx("astaxie1", "author1", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.Exists("astaxie1") {
		t.Error("check err")
	}
}
