package moment

import (
	"rela_recommend/algo/utils"
)

 type MomentAlgoCoarse struct {
	MomentAlgoBase
 }

func (self *MomentAlgoCoarse) Init() {
	self.Model = &utils.XgboostClassifier{}
	utils.LoadModel(self.FilePath, self)
}

func (self *MomentAlgoCoarse) Features(ctx *AlgoContext, data *DataInfo) *utils.Features {
	return GetMomentFeatures(self, ctx, data)
}