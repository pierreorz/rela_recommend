package theme

import (
	"rela_recommend/factory"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

func GetThemeFeatures(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo)*utils.Features {
	fs := &utils.Features{}

	data := idata.(*DataInfo)
	mem := data.MomentCache
	wordVec := model.GetWords()
	memu := data.UserCache //用户缓存信息

	//用户连续特征 1 - 500
	if (memu!= nil) {
		fs.Add(1, float32(memu.Age))
		fs.Add(2, float32(memu.Height))
		fs.Add(3, float32(memu.Weight))

		//用户类别特征 501-599
		fs.AddCategory(501, 12, 0, rutils.GetInt(memu.Horoscope), 0) //星座
		fs.AddCategory(514, 10, 0, rutils.GetInt(memu.RoleName), 0)  //自我认同
	}
	if (memu != nil) {
		fs.AddCategory(525, 5, 0, ctx.GetPlatform(), 0) //用户操作系统
	}
	//词向量 600-856
	wordsCount := len(mem.MomentsText)

	if wordsCount > 0 {
		words := factory.Segmenter.Cut(mem.MomentsText)
		wordNum := make(map[int]float32)
		count := 0
		for i := 0; i < len(words); i++ {
			if dictValue, ok := wordVec[words[i]]; ok {
				count += 1
				for j, num := range dictValue {
					if _, value := wordNum[j]; value {
						wordNum[j] += num
					} else {
						wordNum[j] = num
					}
				}
			}
		}
		if count>0 {
			for k := 0; k < 256; k++ {
				fs.Add(k+600, wordNum[k]/float32(count))
			}
		}
	}
	return fs
}