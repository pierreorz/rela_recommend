package abtest

import (
	"rela_recommend/log"
	"encoding/json"
)

var defaultFactorMap map[string]map[string]string = map[string]map[string]string{}
var testingList []Testing = make([]Testing, 0)
var whiteList []WhiteName = make([]WhiteName, 0)
func init() {
	// 因子默认值
	default_config := `{
		"match": {"match_model": "QuickMatchTreeV1_0", "default_key": "test"}
	}`
	if err := json.Unmarshal(([]byte)(default_config), &defaultFactorMap); err != nil {
		log.Error(err.Error())
	}
	// 测试记录
	ab_config := `[
		{"name": "testing_model", "desc": "测试模型版本", "app": "match", "group": "", "status": 1, "daly_change": 0,
			"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
				{"name": "v1.0", "desc": "", "percentage": 10, "factor_map": {"match_model": "QuickMatchTreeV1_0"}},
				{"name": "v1.2", "desc": "", "percentage": 10, "factor_map": {"match_model": "QuickMatchTreeV1_2"}},
				{"name": "v1.3", "desc": "", "percentage": 10, "factor_map": {"match_model": "QuickMatchTreeV1_3"}}
		]}, 
		{"name": "testing_upper", "desc": "测试活跃加权", "app": "match", "group": "", "status": 1, "daly_change": 1,
			"begin_time": "2018-01-01T09:00:00Z", "end_time": "2020-01-01T09:00:00Z", "versions": [
				{"name": "v1.0", "percentage": 10, "factor_map": {"match_active_user_upper": "0.1"}},
				{"name": "v1.2", "percentage": 10, "factor_map": {"match_active_user_upper": "0.2"}}
		]}
	]`
	if err := json.Unmarshal(([]byte)(ab_config), &testingList); err != nil {
		log.Error(err.Error())
	}
	// 白名单
	white_config := `[
		{"name": "test_test", "desc": "测试", "app": "match", "ids":[1,2,3],"factor_map":{"white_test":"1"}},
		{"name": "match_model", "desc": "匹配模型", "app": "match", "ids":[104708381],"factor_map":{"match_model":"QuickMatchTreeV1_3"}}
	]`
	if err := json.Unmarshal(([]byte)(white_config), &whiteList); err != nil {
		log.Error(err.Error())
	}
}

