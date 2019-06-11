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

	DoInit(*AppInfo, *RecommendRequest) error
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


func (self *ContextBase) DoInit(app *AppInfo, params *RecommendRequest) error {
	self.RankId = uutils.UniqueId()
	self.App = app
	self.AbTest = abtest.GetAbTest(app.Name, params.UserId)
	self.Platform = uutils.GetPlatform(params.Ua)
	self.CreateTime = time.Now()
	self.Request = params
	self.Performs = &performs.Performs{}

	var err error
	var modelName = self.AbTest.GetString(app.AlgoKey, "model_base")
	if model, ok := app.AlgoMap[modelName]; ok {
		self.Algo = model
	} else {
		err = errors.New("algo not found:" + modelName)
	}
	return err
}

// 构建数据
func(self *ContextBase) DoBuildData(buildFunc func(IContext) error) error {
	return buildFunc(self)
}

// 执行特征工程
func(self *ContextBase) DoFeatures() error {
	return self.Algo.DoFeatures(self)
}
// 执行算法
func(self *ContextBase) DoAlgo() error {
	return self.Algo.Predict(self)
}

// 执行策略
func(self *ContextBase) DoStrategies() error {
	var err error
	app := self.GetAppInfo()
	if app.StrategyMap != nil {
		var names = strings.Split(self.AbTest.GetString(app.StrategyKey, ""), ",")
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
	abKey := self.GetAppInfo().SorterKey
	sorts := self.GetAppInfo().SorterMap
	if sorts != nil {
		var name = self.AbTest.GetString(abKey, "base")
		if sorter, ok := sorts[name]; ok {
			err = sorter.Do(self)
		} else {
			err = errors.New("sorter not found:" + name)
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
		var name = self.AbTest.GetString(appInfo.PagerKey, "base")
		if pager, ok := pages[name]; ok {
			return pager.Do(self)
		} else {
			err = errors.New("pager not found:" + name)
		}
	}
	return err
}
// 执行打日志
func(self *ContextBase) DoLog() error {
	var err error
	app := self.GetAppInfo()
	if app.LoggerMap != nil {
		var names = strings.Split(self.AbTest.GetString(app.LoggerKey, "features,performs"), ",")
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
	if err := self.DoInit(app, params); err != nil {
		return err
	}
	pfm := self.GetPerforms()
	pfm.Begin("build_data")
	if err := self.DoBuildData(buildFunc); err != nil {
		return err
	}
	pfm.EndAndBegin("build_data", "features")
	if err := self.DoFeatures(); err != nil {
		return err
	}
	pfm.EndAndBegin("features", "algo")
	if err := self.DoAlgo(); err != nil {
		return err
	}
	pfm.EndAndBegin("algo", "strategies")
	if err := self.DoStrategies(); err != nil {
		return err
	}
	pfm.EndAndBegin("strategies", "sort")
	if err := self.DoSort(); err != nil {
		return err
	}
	pfm.EndAndBegin("sort", "page")
	if err := self.DoPage(); err != nil {
		return err
	}
	pfm.EndAndBegin("page", "log")
	if err := self.DoLog(); err != nil {
		return err
	}
	pfm.End("log")
	return nil
}