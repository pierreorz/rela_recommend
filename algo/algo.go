package algo

import (
	"sync"
	"rela_recommend/algo/utils"
)

// ************************************************** 算法
type IAlgo interface {
	Name() string
	Init()
	Features(IContext, IDataInfo) *utils.Features
	PredictSingle(*utils.Features) float32
	Predict(IContext) error
}

type AlgoBase struct {
	FilePath string
	AlgoName string
	Model utils.IModelAlgo		`json:"model"`
	FeaturesFunc func(IContext, IDataInfo) *utils.Features
}

func (self *AlgoBase) Name() string {
	return self.AlgoName
}

func (self *AlgoBase) Init() {
	// self.Model.Init(self.FilePath)
	utils.LoadModel(self.FilePath, self)
}

func (self *AlgoBase) PredictSingle(features *utils.Features) float32 {
	new_features := self.Model.TransformSingle(features)
	return self.Model.PredictSingle(new_features)
}

// 使用简单计算单个
func (self *AlgoBase) doPredictSingle(ctx IContext, index int) {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	features := self.Features(ctx, dataInfo)
	rankInfo.Features = features
	rankInfo.AlgoScore = self.PredictSingle(features)
	rankInfo.Score = rankInfo.AlgoScore
	rankInfo.AlgoName = self.Name()
}

// 使用简单计算
func (self *AlgoBase) doPredict(ctx IContext) {
	for i := 0; i < ctx.GetDataLength(); i++ {
		self.doPredictSingle(ctx, i)
	}
}
// 使用goroutine多线程并行计算
func (self *AlgoBase) goPredict(ctx IContext, batch int) {
	parts := utils.SplitIndexs(ctx.GetDataLength(), batch)
	wg := new(sync.WaitGroup)
	for _, part := range parts {
		wg.Add(1)
		go func(part []int) {
			defer wg.Done()
			for _, indx := range part {
				self.doPredictSingle(ctx, indx)
			}
        }(part)
	}
	wg.Wait()
}


func (self *AlgoBase) Predict(ctx IContext) error {
	if ctx.GetDataLength() < 100 {
		self.doPredict(ctx)
	} else {
		self.goPredict(ctx, 6)
	}
	return nil
}

func (self *AlgoBase) Features(ctx IContext, data IDataInfo) *utils.Features {
	if self.FeaturesFunc != nil {
		return self.FeaturesFunc(ctx, data)
	}
	return &utils.Features{}
}
