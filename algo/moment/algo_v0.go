package moment

import (
	"rela_recommend/algo/utils"
)

 type MomentAlgoV0 struct {
	MomentAlgoBase
 }

func (self *MomentAlgoV0) Init() {
	self.Model = &utils.XgboostClassifier{}
	utils.LoadModel(self.FilePath, self)
}
