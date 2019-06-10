package match

import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

func FeaturesBase(ctx algo.IContext, model algo.IAlgo, data algo.IDataInfo) *utils.Features {
	// todo features
	_ = ctx.(*algo.ContextBase)
	_ = ctx.GetUserInfo().(*UserInfo)
	_ = data.(*DataInfo)
	return &utils.Features{}
}
