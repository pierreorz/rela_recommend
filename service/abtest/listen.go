package abtest

import (
	"encoding/json"
	"rela_recommend/log"
	"sync"
)

var defaultFactorMap map[string]map[string]Factor = map[string]map[string]Factor{}
var testingMap map[string][]Testing = make(map[string][]Testing, 0)
var whiteListMap map[string][]WhiteName = make(map[string][]WhiteName, 0)
var formulaListMap map[string][]string = make(map[string][]string, 0)

var configLocker = &sync.RWMutex{}
var testLocker = &sync.RWMutex{}
var whiteLocker = &sync.RWMutex{}
var formulaLocker = &sync.RWMutex{}

func getDefaultFactorMap(key string) map[string]Factor {
	configLocker.RLock()
	defer configLocker.RUnlock()
	return defaultFactorMap[key]
}
func setDefaultFactorMap(key string, val map[string]Factor) {
	configLocker.Lock()
	defer configLocker.Unlock()
	defaultFactorMap[key] = val
}

func getTestingMap(key string) []Testing {
	testLocker.RLock()
	defer testLocker.RUnlock()
	return testingMap[key]
}
func setTestingMap(key string, val []Testing) {
	testLocker.Lock()
	defer testLocker.Unlock()
	testingMap[key] = val
}

func getWhiteListMap(key string) []WhiteName {
	whiteLocker.RLock()
	defer whiteLocker.RUnlock()
	return whiteListMap[key]
}
func setWhiteListMap(key string, val []WhiteName) {
	whiteLocker.Lock()
	defer whiteLocker.Unlock()
	whiteListMap[key] = val
}

func getFormulaListMap(key string) []string {
	formulaLocker.RLock()
	defer formulaLocker.RUnlock()
	return formulaListMap[key]
}
func setFormulaListMap(key string, val []string) {
	formulaLocker.Lock()
	defer formulaLocker.Unlock()
	formulaListMap[key] = val
}

func init() {
	// 因子默认值
	default_config := `{
		"match": {"match_model": "QuickMatchTreeV1_0"},
		"live": {"model": "xgb_1.0", "new_score": "0.5", 

				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"strategy:top_recommend_level:weight": "1", 
				"strategy:old_score:weight": "2"},
		"moment": {"strategies": "time_level", "radius_range":"300km", 
				"backend_recommend_switched": "1", "new_moment_len":"0", 
				"new_moment_offset_second": "600", 
				"recommend_list_key": "moment_recommend_list:%d", 

				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"strategy:time_level:weight": "0", 

				"rich_strategy:paged:weight": "1", 
				"rich_strategy:behavior:weight": "1"},
		"moment.near": {"strategies": "time_level", "radius_range":"50km", 

				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"strategy:time_level:weight": "1", 

				"rich_strategy:paged:weight": "1", 
				"rich_strategy:behavior:weight": "1"},
		"theme": {"recommend_new": "1", "backend_recommend_switched": "1", 

				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"rich_strategy:paged:weight": "1", 
				"rich_strategy:behavior:weight": "1"},
		"theme.new": {"recommend_new": "1", "backend_recommend_switched": "0", 
				"recommend_list_key": "",

				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"rich_strategy:paged:weight": "1", 
				"rich_strategy:behavior:weight": "1"},
		"theme.hotweek": {"recommend_new": "0", "backend_recommend_switched": "0", 
				"recommend_list_key": "theme_hotweek_recommend_list:%d",
				
				"logger:features:weight": "1", 
				"logger:performs:weight": "2", 

				"rich_strategy:paged:weight": "1", 
				"rich_strategy:behavior:weight": "1", 
				"rich_strategy:self_upper:weight": "1"}
	}`
	if err := json.Unmarshal(([]byte)(default_config), &defaultFactorMap); err != nil {
		log.Error(err.Error())
		panic(err.Error())
	}
	// 测试记录. 测试名称为0-9a-zA-z_组成
	ab_config := `{
		"match": [
			{"name": "testing_model", "desc": "测试模型版本", "app": "match", "group": "", "status": 2, "daily_change": 0,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0", "desc": "", "percentage": 40, "factor_map": {"match_model": "QuickMatchTreeV1_0"}},
					{"name": "v1.3", "desc": "", "percentage": 40, "factor_map": {"match_model": "QuickMatchTreeV1_3"}},
					{"name": "v1.4", "desc": "", "percentage": 20, "factor_map": {"match_model": "QuickMatchTreeV1_4"}}
			]}, 
			{"name": "testing_upper", "desc": "测试活跃加权", "app": "match", "group": "", "status": 2, "daily_change": 1,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0", "percentage": 10, "factor_map": {"match_active_user_upper": "0.1"}},
					{"name": "v1.2", "percentage": 10, "factor_map": {"match_active_user_upper": "0.2"}}
			]} ],
		"live": [
			{"name": "testing_model_v1_6", "desc": "测试直播模型版本", "app": "live", "group": "", "status": 2, "daily_change": 0,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0.0", "percentage": 20, "factor_map": {"new_score": "0.0", "model": "base"}},
					{"name": "v1.3.5", "percentage": 20, "factor_map": {"new_score": "0.5", "model": "base"}},
					{"name": "v1.4.5", "percentage": 20, "factor_map": {"new_score": "0.5", "model": "xgb_1.0"}}
			]} ],
		"theme": [
			{"name": "testing_user_behavior_lower_v1_0", "desc": "测试用户行为降低", "app": "theme", "group": "", "status": 2, "daily_change": 0,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0", "percentage": 50, "factor_map": {"user_behavior_upper_switch": "true"}},
					{"name": "v1.1", "percentage": 50, "factor_map": {"user_behavior_upper_switch": "false"}}
			]} ]
	}`
	if err := json.Unmarshal(([]byte)(ab_config), &testingMap); err != nil {
		log.Error(err.Error())
		panic(err.Error())
	}
	// 白名单
	white_config := `{
		"match": [
			{"name": "match_model", "desc": "匹配模型", "app": "match", "ids":[104708381],"factor_map":{"match_model":"QuickMatchTreeV1_4"}}
		],
		"live": [
			{"name": "live_model", "desc": "直播模型", "app": "live", "ids":[104708381],"factor_map":{"model":"xgb_1.0", "new_score": "0.5"}}
		],
		"theme": [
			{"name": "theme_behavior", "desc": "话题行为", "app": "theme", "ids":[104708381, 524],"factor_map":{"user_behavior_upper_switch":"false"}}
		]
	}`
	if err := json.Unmarshal(([]byte)(white_config), &whiteListMap); err != nil {
		log.Error(err.Error())
		panic(err.Error())
	}
}
