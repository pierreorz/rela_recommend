package match

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/service"
)

func GetFeaturesV0(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}

	var userInfo = &UserInfo{}
	if ctx.GetUserInfo() != nil {
		userInfo = ctx.GetUserInfo().(*UserInfo)
	}
	dataInfo := idata.(*DataInfo)

	fsIndex := service.UserRow2(userInfo.UserCache, dataInfo.UserCache)

	for i, v := range fsIndex {
		fs.Add(i, v)
	}

	return fs
}
