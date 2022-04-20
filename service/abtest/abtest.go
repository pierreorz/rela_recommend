package abtest

import (
	"encoding/json"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strconv"
	"strings"
	"time"
)

// 测试版本，每个测试版本可以修改多个因子
type TestingVersion struct {
	Name       string            `json:"name"`
	Desc       string            `json:"desc"`
	Percentage int               `json:"percentage"` // 概率 0-100
	FactorMap  map[string]Factor `json:"factor_map"`
}

func (self *TestingVersion) GetFormulaKeys() []string {
	keys := []string{}
	for _, factor := range self.FactorMap {
		keys = append(keys, factor.GetFormulaKeys()...)
	}
	return keys
}

// 测试，每个测试可以包含多个测试版本，每次只会命中其中一个
type Testing struct {
	Name        string           `json:"name"`
	Desc        string           `json:"desc"`
	App         string           `json:"app"`
	Group       string           `json:"group"`
	Status      int              `json:"status"`       //状态,0:新建，1:上线，2:下线
	DailyChange int              `json:"daily_change"` //测试名单每天变化,0:不变，1:变
	BeginTime   time.Time        `json:"begin_time"`
	EndTime     time.Time        `json:"end_time"`
	Versions    []TestingVersion `json:"versions"`
}

func (self *Testing) GetFormulaKeys() []string {
	keys := []string{}
	for _, version := range self.Versions {
		keys = append(keys, version.GetFormulaKeys()...)
	}
	return keys
}

// 白名单，可以设置某些人的某些因子为特定值
type WhiteName struct {
	Name      string            `json:"name"`
	Desc      string            `json:"desc"`
	App       string            `json:"app"`
	Ids       []int64           `json:"ids"`
	FactorMap map[string]Factor `json:"factor_map"`
}

func (self *WhiteName) GetFormulaKeys() []string {
	keys := []string{}
	for _, factor := range self.FactorMap {
		keys = append(keys, factor.GetFormulaKeys()...)
	}
	return keys
}

type AbTest struct {
	App           string                    `json:"app"`     // 服务名称
	DataId        int64                     `json:"data_id"` // userid
	Ua            string                    `json:"ua"`
	Lat           float32                   `json:"lat"`
	Lng           float32                   `json:"lng"`
	DataAttr      map[string]interface{}    `json:"data_attr"`       // user 属性
	RankId        string                    `json:"ds-rank_id"`         // 唯一请求id
	CurrentTime   time.Time                 `json:"create_time"`     // abtest时间
	SettingMap    map[string]string         `json:"setting_map"`     // 用户自定义配置
	FactorMap     map[string]string         `json:"factor_map"`      // 返回的配置对
	HitTestingMap map[string]TestingVersion `json:"hit_testing_map"` // 命中的test
	HitWriteMap   map[string]WhiteName      `json:"hit_write_list"`  // 命中的白名单
}

// 生成用户AB码
func (self *AbTest) generateInt(preString string, dailyChange int) int {
	idString := preString + utils.GetString(self.DataId)
	if dailyChange == 1 {
		idString = self.CurrentTime.Format("2006-01-02") + idString
	}
	// hash32 := fnv.New32a()
	// hash32.Write([]byte(idString))
	// res := hash32.Sum32()

	res := utils.Md5Uint32([]byte(idString))

	return int(res % 100)
}

// 更新因子
func (self *AbTest) updateFactor(newMap map[string]Factor) {
	for key, val := range newMap {
		self.FactorMap[key] = val.GetValue(self.DataAttr)
	}
}
func (self *AbTest) update(newMap map[string]string) {
	for key, val := range newMap {
		self.FactorMap[key] = val
	}
}

// 命中测试
func (self *AbTest) updateTesting(test Testing) {
	var perVal = self.generateInt(test.Name, test.DailyChange) // 随机出AB分组因子
	var perSum int = 0
	for _, version := range test.Versions {
		if perSum <= perVal && perVal < perSum+version.Percentage {
			self.updateFactor(version.FactorMap)
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
			self.updateFactor(white.FactorMap)
			self.HitWriteMap[white.Name] = white
			break
		}
	}
}

func GetFormulaKeys(defMap map[string]Factor, testings []Testing, whiteLists []WhiteName) []string {
	keys := &utils.SetString{}
	for _, factor := range defMap {
		keys.AppendArray(factor.GetFormulaKeys())
	}

	for _, test := range testings {
		keys.AppendArray(test.GetFormulaKeys())
	}

	for _, white := range whiteLists {
		keys.AppendArray(white.GetFormulaKeys())
	}

	return keys.ToList()
}

// 初始化内容
func (self *AbTest) init(settingMap map[string]string) {
	self.CurrentTime = time.Now()
	self.RankId = utils.UniqueId()
	self.SettingMap = settingMap
	self.FactorMap = map[string]string{}
	self.HitTestingMap = map[string]TestingVersion{}
	self.HitWriteMap = map[string]WhiteName{}
	self.DataAttr = map[string]interface{}{}

	// 初始化用户信息
	if keyList := getFormulaListMap(self.App); len(keyList) > 0 {
		self.DataAttr = self.GetUserAttr(keyList)
	}

	// 初始化因子
	if defVal := getDefaultFactorMap(self.App); len(defVal) > 0 {
		self.updateFactor(defVal)
	}
	// 选择测试组, status状态： （-1: 已取消 0: 已删除 1: 待执行 2：运行中 3：已结束 ）
	if testList := getTestingMap(self.App); len(testList) > 0 {
		for _, test := range testList {
			if test.App == self.App && test.Status == 2 && test.BeginTime.Before(self.CurrentTime) && test.EndTime.After(self.CurrentTime) {
				self.updateTesting(test)
			}
		}
	}
	// 增加白名单
	if whiteList := getWhiteListMap(self.App); len(whiteList) > 0 {
		for _, white := range whiteList {
			if white.App == self.App {
				self.updateWhite(white)
			}
		}
	}
	// 配置设置值
	if len(settingMap) > 0 {
		self.update(settingMap)
	}

	// 记录abtest日志
	if logJson, logErr := json.Marshal(self); logErr == nil {
		log.Infof("abtest %s", logJson) // 此日志格式会有实时任务解析，谨慎更改
	}
}

func (self *AbTest) GetString(key string, defVal string) string {
	if val, ok := self.FactorMap[key]; ok {
		return val
	}
	return defVal
}

func (self *AbTest) GetStrings(key string, defVals string) []string {
	strs := self.GetString(key, defVals)
	res := make([]string, 0)
	for _, str := range strings.Split(strs, ",") {
		if len(str) > 0 {
			res = append(res, str)
		}
	}
	return res
}

// 返回字符串集合
func (self *AbTest) GetStringSet(key string, defVals string) *utils.SetString {
	strs := self.GetStrings(key, defVals)
	return utils.NewSetStringFromArray(strs)
}

func (self *AbTest) GetBool(key string, defVal bool) bool {
	if val, ok := self.FactorMap[key]; ok {
		val = strings.ToLower(val)
		return val != "" && val != "0" && val != "f" && val != "false" && val != "no"
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

func (self *AbTest) GetInt64s(key string, defVals string) []int64 {
	strs := self.GetStrings(key, defVals)
	res := []int64{}
	for _, str := range strs {
		if vali, err := strconv.ParseInt(str, 10, 64); err == nil {
			res = append(res, vali)
		}
	}
	return res
}

func (self *AbTest) GetInt64Set(key string, defVals string) *utils.SetInt64 {
	int64s := self.GetInt64s(key, defVals)
	return utils.NewSetInt64FromArray(int64s)
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
func (self *AbTest) GetTestings(expId string,requestId string) string {
	var strMap map[string]string = map[string]string{}
	for key, val := range self.HitTestingMap {
		strMap[key] = val.Name
	}
	strMap["expId"]=expId
	strMap["requestId"]=requestId
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
	return GetAbTestWithSetting(app, dataId, nil)
}

func GetAbTestWithSetting(app string, dataId int64, settingMap map[string]string) *AbTest {
	abtest := AbTest{App: app, DataId: dataId}
	abtest.init(settingMap)
	return &abtest
}

func GetAbTestWithUaLocSetting(app string, dataId int64, ua string, lat float32, lng float32, settingMap map[string]string) *AbTest {
	abtest := AbTest{App: app, DataId: dataId, Ua: ua, Lat: lat, Lng: lng}
	abtest.init(settingMap)
	return &abtest
}
