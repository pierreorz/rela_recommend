package algo

import (
	"time"
	"errors"
	"strings"
	// "rela_recommend/log"
	uutils "rela_recommend/utils"
	"rela_recommend/service/abtest"
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

	DoInit(*AppInfo, *RecommendRequest) error
	DoBuildData(func(IContext) error) error
	DoAlgo() error
	DoStrategies() error
	DoSort() error
	DoPage() (*RecommendResponse, error)
	Do(*AppInfo, *RecommendRequest, func(IContext) error) (*RecommendResponse, error)
}

type ContextBase struct {
	RankId string
	CreateTime time.Time
	Platform int
	App *AppInfo
	Request *RecommendRequest
	AbTest *abtest.AbTest
	DataIds []int64

	User IUserInfo
	DataList []IDataInfo
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

func (self *ContextBase) DoInit(app *AppInfo, params *RecommendRequest) error {
	self.RankId = uutils.UniqueId()
	self.App = app
	self.AbTest = abtest.GetAbTest(app.Name, params.UserId)
	self.Platform = uutils.GetPlatform(params.Ua)
	self.CreateTime = time.Now()
	self.Request = params
	return nil
}

// 构建数据
func(self *ContextBase) DoBuildData(buildFunc func(IContext) error) error {
	return buildFunc(self)
}

// 执行算法
func(self *ContextBase) DoAlgo() error {
	var err error
	app := self.GetAppInfo()
	var modelName = self.AbTest.GetString(app.AlgoKey, "model")
	model, ok := app.AlgoMap[modelName]
	if ok {
		model.Predict(self)
	} else {
		err = errors.New("algo not found:" + modelName)
	}

	return err
}

// 执行策略
func(self *ContextBase) DoStrategies() error {
	var err error
	app := self.GetAppInfo()
	if app.StrategyMap != nil {
		var names = strings.Split(self.AbTest.GetString(app.StrategyKey, "strategies"), ",")
		for _, name := range names {
			if strategy, ok := app.StrategyMap[name]; ok {
				strategy.Do(self)
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
	if sorts == nil || len(sorts) == 0 {
		sorts = map[string]ISorter{"sorter_base": &SorterBase{}}
	}
	var name = self.AbTest.GetString(abKey, "sorter_base")
	if sorter, ok := sorts[name]; ok {
		err = sorter.Do(self)
	} else {
		err = errors.New("sorter not found:" + name)
	}
	return err
}

// 执行排序
func(self *ContextBase) DoPage() (*RecommendResponse, error) {
	var err error
	appInfo := self.GetAppInfo()
	pages := self.GetAppInfo().PagerMap
	if pages == nil || len(pages) == 0 {
		pages = map[string]IPager{"pager_base": &PagerBase{}}
	}
	var name = self.AbTest.GetString(appInfo.PagerKey, "pager_base")
	if pager, ok := pages[name]; ok {
		return pager.Do(self)
	} else {
		err = errors.New("pager not found:" + name)
	}
	return nil, err
}

func(self *ContextBase) Do(app *AppInfo, params *RecommendRequest, buildFunc func(IContext) error) (*RecommendResponse, error) {
	if err := self.DoInit(app, params); err != nil {
		return nil, err
	}
	if err := self.DoBuildData(buildFunc); err != nil {
		return nil, err
	}
	if err := self.DoAlgo(); err != nil {
		return nil, err
	}
	if err := self.DoStrategies(); err != nil {
		return nil, err
	}
	if err := self.DoSort(); err != nil {
		return nil, err
	}
	return self.DoPage()
}