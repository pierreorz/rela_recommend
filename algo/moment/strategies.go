package moment

import(
	"rela_recommend/algo"
)

// 按照6小时优先策略
func DoTimeLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 6
	rankInfo.Level = -hours
	return nil
}
