package base

import (
	"time"
	"errors"
	"strings"
	"rela_recommend/algo"
	"rela_recommend/log"
	uutils "rela_recommend/utils"
	"rela_recommend/service/abtest"
	"rela_recommend/service/performs"
)

type ContextBase struct {
	RankId string
	CreateTime time.Time
	Platform int
	App *algo.AppInfo
	Request *algo.RecommendRequest
	AbTest *abtest.AbTest
	DataIds []int64
	Algo algo.IAlgo

	User algo.IUserInfo
	DataList []algo.IDataInfo

	Performs *performs.Performs
	Response *algo.RecommendResponse
}

func(self *ContextBase) GetRankId() string {
	return self.RankId
}

func(self *ContextBase) GetAppInfo() *algo.AppInfo {
	return self.App
}

func(self *ContextBase) GetCreateTime() time.Time {
	return self.CreateTime
}

func(self *ContextBase) GetPlatform() int {
	return self.Platform
}

func(self *ContextBase) GetRequest() *algo.RecommendRequest {
	return self.Request
}

func(self *ContextBase) GetAbTest() *abtest.AbTest {
	return self.AbTest
}

func(self *ContextBase) GetUserInfo() algo.IUserInfo {
	return self.User
}

func(self *ContextBase) SetUserInfo(user algo.IUserInfo) {
	self.User = user
}

func(self *ContextBase) GetDataIds() []int64 {
	return self.DataIds
}

func(self *ContextBase) SetDataIds(dataIds []int64) {
	self.DataIds = dataIds
}

func(self *ContextBase) GetDataList() []algo.IDataInfo {
	return self.DataList
}

func(self *ContextBase) GetDataLength() int {
	return len(self.DataList)
}

func(self *ContextBase) SetDataList(dataList []algo.IDataInfo) {
	self.DataList = dataList
}

func(self *ContextBase) GetDataByIndex(index int) algo.IDataInfo {
	return self.DataList[index]
}

func(self *ContextBase) GetPerforms() *performs.Performs {
	return self.Performs
}

func(self *ContextBase) GetResponse() *algo.RecommendResponse {
	return self.Response
}

func(self *ContextBase) SetResponse(response *algo.RecommendResponse) {
	self.Response = response
}

func(self *ContextBase) DoNew(app *algo.AppInfo, params *algo.RecommendRequest) error {
	if app == nil {
		return errors.New("app is nil")
	}
	self.App = app
	if params == nil {
		return errors.New("params is nil")
	}
	self.Request = params
	self.Performs = &performs.Performs{}
	return nil
}

func (self *ContextBase) DoInit() error {
	self.RankId = uutils.UniqueId()
	self.AbTest = abtest.GetAbTestWithSetting(self.App.Name, self.Request.UserId, self.Request.AbMap)
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
func(self *ContextBase) DoBuildData() error {
	var err error
	app := self.GetAppInfo()
	if app.BuilderMap != nil {
		var name = self.AbTest.GetString(app.BuilderKey, app.BuilderDefault)
		if len(name) > 0 {
			if builder, ok := app.BuilderMap[name]; ok {
				err = builder.Do(self)
			} else {
				err = errors.New("builder not found:" + name)
			}
		}
	}
	return err
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

	var strategies = []algo.IStrategy{}
	if app.StrategyMap != nil {
		var namestr = self.AbTest.GetString(app.StrategyKey, app.StrategyDefault)
		var names = strings.Split(namestr, ",")
		for _, name := range names {
			if len(name) > 0 {
				if strategy, ok := app.StrategyMap[name]; ok {
					strategies = append(strategies, strategy)
				} else {
					log.Warnf("%s can't find strategy %s", app.Name, name)
				}
			}
		}
	}
	// 添加默认策略，计算推荐分数，计算推荐理由
	strategies = append(strategies, &algo.StrategyBase{ DoSingle: algo.StrategyScoreFunc })
	for _, strategy := range strategies {
		err = strategy.Do(self)
		if err != nil {
			break
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

func(self *ContextBase) Do(app *algo.AppInfo, params *algo.RecommendRequest) error {
	var err = self.DoNew(app, params)
	pfm := self.GetPerforms()
	if err == nil {
		pfm.Begin("init")
		err = self.DoInit()
		pfm.End("init")
	}
	if err == nil {
		pfm.Begin("buildData")
		err = self.DoBuildData()
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
