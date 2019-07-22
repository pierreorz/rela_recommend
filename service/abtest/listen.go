package abtest

import (
	"rela_recommend/log"
	"encoding/json"
)

var defaultFactorMap map[string]map[string]string = map[string]map[string]string{}
var testingMap map[string][]Testing = make(map[string][]Testing, 0)
var whiteListMap map[string][]WhiteName = make(map[string][]WhiteName, 0)
func init() {
	// 因子默认值
	default_config := `{
		"match": {"match_model": "QuickMatchTreeV1_0", "default_key": "test"},
		"live": {"live_model": "LiveModelV1_3", "new_score": "0.5"},
		"moment": {"strategies": "time_level", "radius_range":"300km"},
		"moment.near": {"strategies": "time_level", "radius_range":"50km"}
	}`
	if err := json.Unmarshal(([]byte)(default_config), &defaultFactorMap); err != nil {
		log.Error(err.Error())
	}
	// 测试记录. 测试名称为0-9a-zA-z_组成
	ab_config := `{
		"match": [
			{"name": "testing_model", "desc": "测试模型版本", "app": "match", "group": "", "status": 1, "daily_change": 0,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0", "desc": "", "percentage": 40, "factor_map": {"match_model": "QuickMatchTreeV1_0"}},
					{"name": "v1.3", "desc": "", "percentage": 40, "factor_map": {"match_model": "QuickMatchTreeV1_3"}},
					{"name": "v1.4", "desc": "", "percentage": 20, "factor_map": {"match_model": "QuickMatchTreeV1_4"}}
			]}, 
			{"name": "testing_upper", "desc": "测试活跃加权", "app": "match", "group": "", "status": 1, "daily_change": 1,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0", "percentage": 10, "factor_map": {"match_active_user_upper": "0.1"}},
					{"name": "v1.2", "percentage": 10, "factor_map": {"match_active_user_upper": "0.2"}}
			]} ],
		"live": [
			{"name": "testing_model_v1_4", "desc": "测试直播模型版本", "app": "live", "group": "", "status": 1, "daily_change": 1,
				"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
					{"name": "v1.0.0", "percentage": 15, "factor_map": {"new_score": "0.0", "model": "base"}},
					{"name": "v1.3.3", "percentage": 15, "factor_map": {"new_score": "0.3", "model": "base"}},
					{"name": "v1.3.5", "percentage": 15, "factor_map": {"new_score": "0.5", "model": "base"}},
					{"name": "v1.4.3", "percentage": 15, "factor_map": {"new_score": "0.3", "model": "xgb_1.0"}},
					{"name": "v1.4.5", "percentage": 15, "factor_map": {"new_score": "0.5", "model": "xgb_1.0"}}
			]} ],
		"moment.near": [
			
		]
	}`
	if err := json.Unmarshal(([]byte)(ab_config), &testingMap); err != nil {
		log.Error(err.Error())
	}
	// 白名单
	white_config := `{
		"match": [
			{"name": "test_test", "desc": "测试", "app": "match", "ids":[1,2,3],"factor_map":{"white_test":"1"}},
			{"name": "match_model", "desc": "匹配模型", "app": "match", "ids":[104708381],"factor_map":{"match_model":"QuickMatchTreeV1_4"}}
		],
		"live": [
			{"name": "live_model", "desc": "直播模型", "app": "live", "ids":[104708381],"factor_map":{"model":"xgb_1.0", "new_score": "0.6"}}
		],
		"moment.near": [
			{"name": "moment_model", "desc": "日志白名单", "app": "moment", "ids":[104708381],"factor_map":{"strategies": "time_level", "radius_range":"50km"}}
		],
		"theme": [
			{"name": "theme_model", "desc": "话题白名单", "app": "theme", "ids":[104708381],"factor_map":{"strategies": "hot,user_behavior"}}
		]
	}`
	if err := json.Unmarshal(([]byte)(white_config), &whiteListMap); err != nil {
		log.Error(err.Error())
	}
}

