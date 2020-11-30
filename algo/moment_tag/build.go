package moment_tag

import (
	"rela_recommend/algo"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	abtest := ctx.GetAbTest()

	// 确定候选用户
	dataIds := params.DataIds
	if abtest.GetBool("always_use_search", false) { // 是否一直使用search
		pf.Run("search", func(*performs.Performs) interface{} {
			var searchErr error
			if dataIds, searchErr = search.CallSearchMomentTagIdList(params.UserId, params.Lat, params.Lng,
				params.Offset, params.Limit, params.Params["query"]); searchErr == nil {
				return len(dataIds)
			} else {
				return searchErr
			}
		})
	}

	// 组装用户信息
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId: params.UserId,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for _, dataId := range dataIds {
			info := &DataInfo{
				DataId:   dataId,
				RankInfo: &algo.RankInfo{},
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}
