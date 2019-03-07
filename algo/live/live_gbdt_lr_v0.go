package live

import (
	"rela_recommend/algo/utils"
)

type LiveGbdtLrV0 struct {
	LiveAlgoBase
}

func (self *LiveGbdtLrV0) Init() {
	model := &utils.GradientBoostingLRClassifier{}
	model.Init(self.FilePath)
	self.model = model
}
