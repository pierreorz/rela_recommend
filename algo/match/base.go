package match

import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)


type MatchAlgo struct {
	algo.AlgoBase
}

func (self *MatchAlgo) Features(ctx algo.IContext, data algo.IDataInfo) *utils.Features {
	// todo features
	_ = ctx.(*algo.ContextBase)
	_ = ctx.GetUserInfo().(*UserInfo)
	_ = data.(*DataInfo)
	return &utils.Features{}
}
