package algo

import (
	"os"
	"rela_recommend/algo/utils"
	utils2 "rela_recommend/utils"
	"strings"
	"sync"
)

// ************************************************** 算法
type IAlgo interface {
	Name() string
	Init()
	FeaturesSingle(IContext, IDataInfo) *utils.Features
	DoFeatures(IContext) error
	PredictSingle(*utils.Features) float32
	Predict(IContext) error
	GetWords() map[string][]float32
	CheckWords([]string) []string
}

type AlgoBase struct {
	FilePath string
	AlgoName string
	Model utils.IModelAlgo			`json:"model"`
	Words map[string][]float32		`json:"words"`
	FeaturesFunc func(IContext, IAlgo, IDataInfo) *utils.Features
}




func (self *AlgoBase) Name() string {
	return self.AlgoName
}

func (self *AlgoBase) Init() {
	// self.Model.Init(self.FilePath)
	utils.LoadModel(self.FilePath, self)
}

func (self *AlgoBase) GetWords() map[string][]float32 {
	return self.Words
}


func (self *AlgoBase) PredictSingle(features *utils.Features) float32 {
	new_features := self.Model.TransformSingle(features)
	return self.Model.PredictSingle(new_features)
}

// 使用简单计算单个
func (self *AlgoBase) doPredictSingle(ctx IContext, index int) {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	if rankInfo.Features == nil {
		rankInfo.Features = self.FeaturesSingle(ctx, dataInfo)
	}
	if rankInfo.ExpId!=""&&!strings.Contains(rankInfo.ExpId,utils2.OffTime)&&!strings.Contains(rankInfo.ExpId,utils2.RequestErr){
		rankInfo.AlgoScore = float32(rankInfo.PaiScore)
		rankInfo.AlgoName="pai"
		rankInfo.Score=rankInfo.AlgoScore
	}else{
		rankInfo.AlgoScore=self.PredictSingle(rankInfo.Features)
		rankInfo.Score = rankInfo.AlgoScore
		rankInfo.AlgoName = self.Name()
	}
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

func (self *AlgoBase) FeaturesSingle(ctx IContext, data IDataInfo) *utils.Features {
	return self.FeaturesFunc(ctx, self, data)
}
func (self *AlgoBase) doFeaturesSingle(ctx IContext, index int) {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	rankInfo.Features = self.FeaturesSingle(ctx, dataInfo)
}

func (self *AlgoBase) doFeatures(ctx IContext) error {
	for i := 0; i < ctx.GetDataLength(); i++ {
		self.doFeaturesSingle(ctx, i)
	}
	return nil
}

func (self *AlgoBase) goFeatures(ctx IContext, batch int) error {
	parts := utils.SplitIndexs(ctx.GetDataLength(), batch)
	wg := new(sync.WaitGroup)
	for _, part := range parts {
		wg.Add(1)
		go func(part []int) {
			defer wg.Done()
			for _, indx := range part {
				self.doFeaturesSingle(ctx, indx)
			}
        }(part)
	}
	wg.Wait()
	return nil
}
func (self *AlgoBase) DoFeatures(ctx IContext) error {
	abtest := ctx.GetAbTest()
	threshold := abtest.GetInt("algo.thread.threshold", 100)
	if ctx.GetDataLength() < threshold {
		return self.doFeatures(ctx)
	} else {
		jobs := abtest.GetInt("algo.thread.jobs", 6)
		return self.goFeatures(ctx, jobs)
	}
}

// 检查词是否被允许
func (self *AlgoBase) CheckWords(words []string) []string {
	res := make([]string, 0)
	for _, word := range words {
		if _, ok := self.Words[word]; ok {
			res = append(res, word)
		}
	}
	return res
}

// 获取当前工作目录
func GetWorkDir(path string) string {
	work_dir, _ := os.Getwd()
	return work_dir + path
}

// 将算法列表初始化，并且生成到指定Map
func AlgoListInitToMap(algoList []IAlgo) map[string]IAlgo {
	algoMap := map[string]IAlgo{}
	for index, _ := range algoList {
		algoList[index].Init()
		algoMap[algoList[index].Name()] = algoList[index]
	}
	return algoMap
}
