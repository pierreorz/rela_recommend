package abtest

import (
	"time"
	"strconv"
	"strings"
	"hash/fnv"
	"encoding/json"
	"rela_recommend/utils"
	"rela_recommend/log"
)

// 测试版本，每个测试版本可以修改多个因子
type TestingVersion struct {
	Name string                 `json:"name"` 
	Desc string					`json:"desc"`
	Percentage int          	`json:"percentage"`  // 概率 0-100
	FactorMap map[string]string `json:"factor_map"`
}

// 测试，每个测试可以包含多个测试版本，每次只会命中其中一个
type Testing struct {
	Name string                 `json:"name"` 
	Desc string					`json:"desc"`
	App string					`json:"app"`
	Group string                `json:"group"` 
	Status int                  `json:"status"`		 //状态,0:新建，1:上线，2:下线
	DalyChange int				`json:"daly_change"` //测试名单每天变化,0:不变，1:变
	BeginTime time.Time         `json:"begin_time"` 
	EndTime time.Time           `json:"end_time"` 
	Versions []TestingVersion   `json:"versions"` 
}

// 白名单，可以设置某些人的某些因子为特定值
type WhiteName struct {
	Name string					`json:"name"`
	Desc string					`json:"desc"`
	App string					`json:"app"`
	Ids []int64					`json:"ids"`
	FactorMap map[string]string `json:"factor_map"`
}


type AbTest struct {
	App string
	DataId int64
	CurrentTime time.Time
	FactorMap map[string]string
	HitTestingMap map[string]TestingVersion
	HitWriteList []WhiteName
}

// 生成用户AB码
func (self *AbTest) generateInt(preString string, dalyChange int) int {
	idString := preString + utils.GetString(self.DataId)
	if(dalyChange == 1) {
		idString = self.CurrentTime.Format("2000-01-01") + idString
	}
	hash32 := fnv.New32a()
	hash32.Write([]byte(idString))
	res := hash32.Sum32()
	
	return int(res % 100)
}
// 更新因子
func (self *AbTest) update(newMap map[string]string) {
	for key, val := range newMap {
		self.FactorMap[key] = val
	}
}
// 命中测试
func (self *AbTest) updateTesting(test Testing) {
	var perVal = self.generateInt(test.Name, test.DalyChange)  // 随机出AB分组因子
	var perSum int = 0
	for _, version := range test.Versions {
		if perSum <= perVal && perVal < perSum + version.Percentage {
			self.update(version.FactorMap)
			self.HitTestingMap[test.Name] = version
			break
		}
		perSum = perSum + version.Percentage
	}
	if perSum > 100 {
		log.Warn("percentage sum > 100:", test.Name)
	}
}
// 命中测试
func (self *AbTest) updateWhite(white WhiteName) {
	for _, id := range white.Ids {
		if id == self.DataId {
			self.update(white.FactorMap)
			self.HitWriteList = append(self.HitWriteList, white)
			break
		}
	}
}
// 初始化内容
func (self *AbTest) Init(defMap map[string]map[string]string, testingMap map[string][]Testing, whiteListMap map[string][]WhiteName) {
	self.CurrentTime = time.Now()
	self.FactorMap = map[string]string{}
	self.HitTestingMap = map[string]TestingVersion{}
	self.HitWriteList = []WhiteName{}
	// 初始化因子
	if defVal, ok := defMap[self.App]; ok {
		self.update(defVal)
	}
	// 选择测试组
	if testList, ok := testingMap[self.App]; ok {
		for _, test := range testList {
			if test.App == self.App && test.Status == 1 && test.BeginTime.Before(self.CurrentTime) && test.EndTime.After(self.CurrentTime) {
				self.updateTesting(test)
			}
		}
	}
	// 增加白名单
	if whiteList, ok := whiteListMap[self.App]; ok {
		for _, white := range whiteList {
			if white.App == self.App {
				self.updateWhite(white)
			}
		}
	}
}

func (self *AbTest) GetString(key string, defVal string) string {
	if val, ok := self.FactorMap[key]; ok {
		return val
	}
	return defVal
}
func (self *AbTest) GetBool(key string, defVal bool) bool {
	if val, ok := self.FactorMap[key]; ok {
		val = strings.ToLower(val)
		return val != "0" && val != "f" && val != "false" && val != "no"
	}
	return defVal
}
func (self *AbTest) GetInt64(key string, defVal int64) int64 {
	if val, ok := self.FactorMap[key]; ok {
		vali, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			return vali
		} else {
			log.Warnf("%s:%s can't parse to int", key, val)
		}
	}
	return defVal
}
func (self *AbTest) GetInt(key string, defVal int) int {
	return int(self.GetInt64(key, int64(defVal)))
}

func (self *AbTest) GetFloat64(key string, defVal float64) float64 {
	if val, ok := self.FactorMap[key]; ok {
		vali, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return vali
		} else {
			log.Warnf("%s:%s can't parse to float", key, val)
		}
	}
	return defVal
}
func (self *AbTest) GetFloat(key string, defVal float32) float32 {
	return float32(self.GetFloat64(key, float64(defVal)))
}
func (self *AbTest) GetTestings() string {
	var strMap map[string]string = map[string]string{}
	for key, val := range self.HitTestingMap {
		strMap[key] = val.Name
	}
	jsonStr, err := json.Marshal(strMap)
	if err == nil {
		res := string(jsonStr)
		return strings.Replace(res, " ", "", -1)
	} else {
		log.Warn(err.Error())
		return ""
	}
}
func (self *AbTest) IsSwitchOn(key string, defVal bool) bool {
	val := self.GetInt(key, -1)
	if val >= 0 {
		randomVal := self.generateInt(key, 0)
		return randomVal < val
	}
	return defVal
}

func GetAbTest(app string, dataId int64) *AbTest {
	abtest := AbTest{App: app, DataId: dataId}
	abtest.Init(defaultFactorMap, testingMap, whiteListMap)
	return &abtest
}
