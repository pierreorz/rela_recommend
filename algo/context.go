package algo

import (
	"time"
	"errors"
	"strings"
	"rela_recommend/log"
	uutils "rela_recommend/utils"
	"rela_recommend/service/abtest"
	"rela_recommend/service/performs"
)

// ************************************************** 上下文
type IUserInfo interface {
}

type IDataInfo interface {
	GetDataId() int64
	GetRankInfo() *RankInfo
	SetRankInfo(*RankInfo)
}

type IContext interface {
	GetRankId() string
	GetCreateTime() time.Time
	GetPlatform() int
	GetAppInfo() *AppInfo
	GetRequest() *RecommendRequest
	GetAbTest() *abtest.AbTest

	GetUserInfo() IUserInfo
	SetUserInfo(IUserInfo)
	GetDataIds() []int64
	SetDataIds([]int64)
	GetDataList() []IDataInfo
	GetDataLength() int
	SetDataList([]IDataInfo)
	GetDataByIndex(int) IDataInfo

	SetResponse(*RecommendResponse)
	GetResponse() *RecommendResponse

	GetPerforms() *performs.Performs

	DoNew(*AppInfo, *RecommendRequest) error
	DoInit() error
	DoBuildData(func(IContext) error) error
	DoFeatures() error
	DoAlgo() error
	DoStrategies() error
	DoSort() error
	DoPage() error
	DoLog() error
	Do(*AppInfo, *RecommendRequest, func(IContext) error) error
}

type ContextBase struct {
	RankId string
	CreateTime time.Time
	Platform int
	App *AppInfo
	Request *RecommendRequest
	AbTest *abtest.AbTest
	DataIds []int64
	Algo IAlgo

	User IUserInfo
	DataList []IDataInfo

	Performs *performs.Performs
	Response *RecommendResponse
}

func(self *ContextBase) GetRankId() string {
	return self.RankId
}

func(self *ContextBase) GetAppInfo() *AppInfo {
	return self.App
}

func(self *ContextBase) GetCreateTime() time.Time {
	return self.CreateTime
}

func(self *ContextBase) GetPlatform() int {
	return self.Platform
}

func(self *ContextBase) GetRequest() *RecommendRequest {
	return self.Request
}

func(self *ContextBase) GetAbTest() *abtest.AbTest {
	return self.AbTest
}

func(self *ContextBase) GetUserInfo() IUserInfo {
	return self.User
}

func(self *ContextBase) SetUserInfo(user IUserInfo) {
	self.User = user
}

func(self *ContextBase) GetDataIds() []int64 {
	return self.DataIds
}

func(self *ContextBase) SetDataIds(dataIds []int64) {
	self.DataIds = dataIds
}

func(self *ContextBase) GetDataList() []IDataInfo {
	return self.DataList
}

func(self *ContextBase) GetDataLength() int {
	return len(self.DataList)
}

func(self *ContextBase) SetDataList(dataList []IDataInfo) {
	self.DataList = dataList
}

func(self *ContextBase) GetDataByIndex(index int) IDataInfo {
	return self.DataList[index]
}

func(self *ContextBase) GetPerforms() *performs.Performs {
	return self.Performs
}

func(self *ContextBase) GetResponse() *RecommendResponse {
	return self.Response
}

func(self *ContextBase) SetResponse(response *RecommendResponse) {
	self.Response = response
}

func(self *ContextBase) DoNew(app *AppInfo, params *RecommendRequest) error {
	self.App = app
	self.Request = params
	self.Performs = &performs.Performs{}
	return nil
}

func (self *ContextBase) DoInit() error {
	self.RankId = uutils.UniqueId()
	self.AbTest = abtest.GetAbTest(self.App.Name, self.Request.UserId)
	self.Platform = uutils.GetPlatform(self.Request.Ua)
	self.CreateTime = time.Now()

	var err error
	var modelName = self.AbTest.GetString(self.App.AlgoKey, self.App.AlgoDefault)
	if self.App.AlgoMap != nil {
		if model, ok := self.App.AlgoMap[modelName]; ok {
			self.Algo = model
		} else {
			err = errors.New("algo not found:" + modelName)
		}
	}
	return err
}

// 构建数据
func(self *ContextBase) DoBuildData(buildFunc func(IContext) error) error {
	return buildFunc(self)
}

// 执行特征工程
func(self *ContextBase) DoFeatures() error {
	if self.Algo != nil {
		return self.Algo.DoFeatures(self)
	}
	return nil
}
// 执行算法
func(self *ContextBase) DoAlgo() error {
	if self.Algo != nil {
		return self.Algo.Predict(self)
	}
	return nil
}

// 执行策略
func(self *ContextBase) DoStrategies() error {
	var err error
	app := self.GetAppInfo()
	if app.StrategyMap != nil {
		var namestr = self.AbTest.GetString(app.StrategyKey, app.StrategyDefault)
		var names = strings.Split(namestr, ",")
		for _, name := range names {
			if len(name) > 0 {
				if strategy, ok := app.StrategyMap[name]; ok {
					err = strategy.Do(self)
					if err != nil {
						break
					}
				} else {
					log.Warnf("%s can't find strategy %s", app.Name, name)
				}
			}
		}
	}
	return err
}

// 执行排序
func(self *ContextBase) DoSort() error {
	var err error
	app := self.GetAppInfo()
	if app.SorterMap != nil {
		var name = self.AbTest.GetString(app.SorterKey, app.SorterDefault)
		if len(name) > 0 {
			if sorter, ok := app.SorterMap[name]; ok {
				err = sorter.Do(self)
			} else {
				err = errors.New("sorter not found:" + name)
			}
		}
	}
	return err
}

// 执行分页
func(self *ContextBase) DoPage() error {
	var err error
	appInfo := self.GetAppInfo()
	pages := self.GetAppInfo().PagerMap
	if pages != nil {
		var name = self.AbTest.GetString(appInfo.PagerKey, appInfo.PagerDefault)
		if len(name) > 0 {
			if pager, ok := pages[name]; ok {
				return pager.Do(self)
			} else {
				err = errors.New("pager not found:" + name)
			}
		}
	}
	return err
}
// 执行打日志
func(self *ContextBase) DoLog() error {
	var err error
	app := self.GetAppInfo()
	if app.LoggerMap != nil {
		var namestr = self.AbTest.GetString(app.LoggerKey, app.LoggerDefault)
		var names = strings.Split(namestr, ",")
		for _, name := range names {
			if len(name) > 0 {
				if logger, ok := app.LoggerMap[name]; ok {
					err = logger.Do(self)
					if err != nil {
						break
					}
				} else {
					log.Warnf("%s can't find logger %s", app.Name, name)
				}
			}
		}
	}

	return err
}

func(self *ContextBase) Do(app *AppInfo, params *RecommendRequest, buildFunc func(IContext) error) error {
	var err error
	err = self.DoNew(app, params)
	pfm := self.GetPerforms()
	if err == nil {
		pfm.Begin("init")
		err = self.DoInit()
		pfm.End("init")
	}
	if err == nil {
		pfm.Begin("buildData")
		err = self.DoBuildData(buildFunc)
		pfm.End("buildData")
	}
	if err == nil {
		pfm.Begin("features")
		err = self.DoFeatures()
		pfm.End("features")
	}
	if err == nil {
		pfm.Begin("algo")
		err = self.DoAlgo()
		pfm.End("algo")
	}
	if err == nil {
		pfm.Begin("strategies")
		err = self.DoStrategies()
		pfm.End("strategies")
	}
	if err == nil {
		pfm.Begin("sort")
		err = self.DoSort()
		pfm.End("sort")
	}
	if err == nil {
		pfm.Begin("page")
		err = self.DoPage()
		pfm.End("page")
	}

	pfm.Begin("log")
	err1 := self.DoLog()
	if err == nil {
		err = err1
	}
	pfm.End("log")

	return err
}