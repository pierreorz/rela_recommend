package abtest

import (
	"time"
	"encoding/json"
	"strconv"
	"crypto/md5"
	"rela_recommend/utils"
	"rela_recommend/log"
)

func GetMd5Int64(userId int64) int64 {
	idString := utils.GetString(userId)

	md5New := md5.New()
	md5New.Write([]byte(idString))
	bytes := md5New.Sum(nil)
	
	return utils.BytesToInt64(bytes)
}

type TestingVersion struct {
	Name string                 `json:"name"` 
	Percentage float32          `json:"percentage"` 
	FactorMap map[string]string `json:"factor_map"`
}

type Testing struct {
	Name string                 `json:"name"` 
	Group string                `json:"group"` 
	Status int                  `json:"status"` 
	BeginTime time.Time         `json:"begin_time"` 
	EndTime time.Time           `json:"end_time"` 
	Versions []TestingVersion   `json:"versions"` 
}

type AbTest struct {
	DataId int64
	FactorMap map[string]string
}
func (self *AbTest) GetString(key string, defVal string) string {
	if val, ok := self.FactorMap[key]; ok {
		return val
	}
	return defVal
}
func (self *AbTest) GetInt64(key string, defVal int64) int64 {
	if val, ok := self.FactorMap[key]; ok {
		vali, err := strconv.ParseInt(val, 10, 8)
		if err == nil {
			return vali
		} else {
			log.Warnf("%s:%s can't parse to int", key, val)
		}
	}
	return defVal
}
func (self *AbTest) GetFloat64(key string, defVal float64) float64 {
	if val, ok := self.FactorMap[key]; ok {
		vali, err := strconv.ParseFloat(val, 8)
		if err == nil {
			return vali
		} else {
			log.Warnf("%s:%s can't parse to float", key, val)
		}
	}
	return defVal
}


var defaultFactorMap map[string]string
var testings []Testing = make([]Testing, 0)
var whiteList map[string]map[string]string
func init() {
	// 因子默认值
	defaultFactorMap = map[string]string{"match_model": "MatchAlgoV1_0"}
	// 测试记录
	ab_config := `[
		{"name": "测试模型版本", "app": "match", "group": "", "status": 1, "begin_time": "2018-01-01 09:00:00", "end_time": "2020-01-01 09:00:00", "versions": [
			{"name": "v1.0", "percentage": 20, "factor_map": {"model": "MatchAlgoV1_0"}},
			{"name": "v1.1", "percentage": 20, "factor_map": {"model": "MatchAlgoV1_1"}}
		]}, 
		{"name": "测试活跃加权", "app": "match", "group": "", "status": 1, "begin_time": "2018-01-01 09:00:00", "end_time": "2020-01-01 09:00:00", "versions": [
			{"name": "v1.0", "percentage": 20, "factor_map": {"24hour_upper": "1"}},
			{"name": "v1.1", "percentage": 20, "factor_map": {"24hour_upper": "1.1"}},
			{"name": "v1.2", "percentage": 20, "factor_map": {"24hour_upper": "1.2"}}
		]}
	]`
	if err := json.Unmarshal(([]byte)(ab_config), &testings); err != nil {
		log.Error(err.Error())
	}
	// 白名单
	whiteList = map[string]map[string]string{"0":map[string]string{}}
}


func GetAbTest(app string, dataId int64) AbTest {
	// todo 初始化默认因子
	// todo 以测试纪录测试因子
	// todo 添加白名单因子
	abtest := AbTest{DataId: dataId}
	return abtest
}