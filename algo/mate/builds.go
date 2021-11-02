package mate

import (
	"rela_recommend/algo"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()

	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}

	// 获取search的广告列表
	var searchResList []search.MateTextResDataItem
	pf.Run("search", func(*performs.Performs) interface{} {
		var searchErr error
		if searchResList, searchErr = search.CallMateTextList(params); searchErr == nil {
			return len(searchResList)
		} else {
			return searchErr
		}
	})

	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId: params.UserId,
		}

		// 组装被曝光者信息
		dataIds := make([]int64, 0)
		dataList := make([]algo.IDataInfo, 0)
		for i, searchRes := range searchResList {
			info := &DataInfo{
				DataId:     searchRes.Id,
				SearchData: &searchResList[i],
				RankInfo:   &algo.RankInfo{},
			}
			dataIds = append(dataIds, searchRes.Id)
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
