package moment

import (
	"sort"
	"errors"
)

// ************************************** 用户排序
type DataListSorter struct {
	List []DataInfo
	Context *AlgoContext
} 

func (self DataListSorter) Swap(i, j int) {
	self.List[i], self.List[j] = self.List[j], self.List[i]
}
func (self DataListSorter) Len() int {
	return len(self.List)
}
// 以此按照：打分，最后登陆时间
func (self DataListSorter) Less(i, j int) bool {
	ranki, rankj := self.List[i].RankInfo, self.List[j].RankInfo

	if ranki.IsTop != rankj.IsTop {
		return ranki.IsTop > rankj.IsTop		// IsTop ： 倒序， 是否置顶
	} else {
		if ranki.Level != rankj.Level {
			return ranki.Level > rankj.Level		// Level : 倒序， 推荐星数
		} else {
			hoursI := int(self.Context.CreateTime.Sub(self.List[i].MomentCache.InsertTime).Hours()) / 12
			hoursJ := int(self.Context.CreateTime.Sub(self.List[j].MomentCache.InsertTime).Hours()) / 12
			if hoursI != hoursJ {
				return hoursI < hoursJ					// 每24小时优先: 正序
			} else {
				if ranki.Score != rankj.Score {
					return ranki.Score > rankj.Score		// Score : 倒序， 推荐分数
				} else {
					return self.List[i].DataId < self.List[j].DataId	// UserId : 正序
				}
			}
		}
	}
}
func (self *DataListSorter) Sort() {
	sort.Sort(self)
}
func (self *DataListSorter) DoStrategies() *DataListSorter {
	strategies := []Strategy{}
	for i, _ := range strategies {
		strategies[i].Do(self.Context, self.List)
	}
	return self
}
func (self *DataListSorter) DoAlgo() error {
	var err error
	var modelName = self.Context.AbTest.GetString("moment_model", "MomentModelV1_0")
	model, ok := AlgosMap[modelName]
	if ok {
		model.Predict(self.Context)
	} else {
		err = errors.New("algo not found:" + modelName)
	}

	return err
}

