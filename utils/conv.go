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

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// GetString convert interface to string.
func GetString(v interface{}) string {
	switch result := v.(type) {
	case string:
		return result
	case []byte:
		return string(result)
	default:
		if v != nil {
			return fmt.Sprintf("%v", result)
		}
	}
	return ""
}

// GetBytes convert interface to []byte.
func GetBytes(v interface{}) []byte {
	switch result := v.(type) {
	case string:
		return []byte(result)
	case []byte:
		return result
	default:
		if v != nil {
			return []byte(fmt.Sprintf("%v", result))
		}
	}
	return nil
}

// GetInt convert interface to int.
func GetInt(v interface{}) int {
	switch result := v.(type) {
	case int:
		return result
	case int32:
		return int(result)
	case int64:
		return int(result)
	case bool:
		if bool(result) {
			return 1
		} else {
			return 0
		}
	case float32:
		return int(result)
	case float64:
		return int(result)
	default:
		if d := GetString(v); d != "" {
			value, _ := strconv.Atoi(d)
			return value
		}
	}
	return 0
}

// GetInt64 convert interface to int64.
func GetInt64(v interface{}) int64 {
	switch result := v.(type) {
	case int:
		return int64(result)
	case int32:
		return int64(result)
	case int64:
		return result
	case float32:
		return int64(result)
	case float64:
		return int64(result)
	default:
		if d := GetString(v); d != "" {
			value, _ := strconv.ParseInt(d, 10, 64)
			return value
		}
	}
	return 0
}

// GetFloat64 convert interface to float64.
func GetFloat64(v interface{}) float64 {
	switch result := v.(type) {
	case float64:
		return result
	default:
		if d := GetString(v); d != "" {
			value, _ := strconv.ParseFloat(d, 64)
			return value
		}
	}
	return 0
}

// GetBool convert interface to bool.
func GetBool(v interface{}) bool {
	switch result := v.(type) {
	case bool:
		return result
	default:
		if d := GetString(v); d != "" {
			value, _ := strconv.ParseBool(d)
			return value
		}
	}
	return false
}

func GetInterfaces(v interface{}) []interface{} {
	res := make([]interface{}, 0)
	if v != nil {
		switch result := v.(type) {
		case []interface{}:
			return result
		case []int64:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []int32:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []int:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []float32:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []float64:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []bool:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []string:
			for _, uid := range result {
				res = append(res, uid)
			}
		case []byte:
			for _, uid := range result {
				res = append(res, uid)
			}
		case [][]byte:
			for _, uid := range result {
				res = append(res, uid)
			}
		}
	}
	return res
}

func GetIntsWithStrings(s []string) []int {
	var ids = make([]int, 0)
	for _, uid := range s {
		ids = append(ids, GetInt(uid))
	}
	return ids
}

func GetInts(v interface{}) []int {
	if v != nil {
		switch result := v.(type) {
		case []int:
			return result
		case []string:
			return GetIntsWithStrings(result)
		case string:
			var ids = make([]int, 0)
			for _, uid := range strings.Split(GetString(v), ",") {
				ids = append(ids, GetInt(uid))
			}
			return ids
		}
	}
	return make([]int, 0)
}

func GetInt64sWithStrings(s []string) []int64 {
	var ids = make([]int64, 0)
	for _, uid := range s {
		ids = append(ids, GetInt64(uid))
	}
	return ids
}
func GetInt64s(v interface{}) []int64 {
	if v != nil {
		switch result := v.(type) {
		case []int64:
			return result
		case []string:
			return GetInt64sWithStrings(result)
		case string:
			var ids = make([]int64, 0)
			for _, uid := range strings.Split(GetString(v), ",") {
				if len(uid) > 0 {
					ids = append(ids, GetInt64(uid))
				}
			}
			return ids
		}
	}
	return make([]int64, 0)
}

