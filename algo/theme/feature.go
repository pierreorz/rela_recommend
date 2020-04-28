package theme

import (
	"math"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/factory"
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

	fs.AddCategory(525, 5, 0, ctx.GetPlatform(), 0) //用户操作系统

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


func GetThemeFeaturesv0(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo)*utils.Features {
	fs := &utils.Features{}

	data := idata.(*DataInfo)
	mem := data.MomentCache
	wordVec := model.GetWords()
	memu := data.UserCache
	memex := data.MomentExtendCache
	if (memu != nil) {
		fs.Add(1, float32(memu.Age))
		fs.Add(2, float32(memu.Height))
		fs.Add(3, float32(memu.Weight))
		fs.Add(5, float32(memex.AndroidFlag))
	}
	if ctx.GetUserInfo() != nil {
		userData := ctx.GetUserInfo().(*UserInfo)
		reqUser := userData.UserCache
		if (reqUser != nil) {
			fs.Add(10, float32(reqUser.Age))
			fs.Add(11, float32(reqUser.Height))
			fs.Add(12, float32(reqUser.Weight))
		}
		if (userData.ThemeUser != nil) {
			UserAls := userData.ThemeUser
			userAls_line := UserAls.UserEmbedding
			if len(userAls_line) > 0 {
				fs.AddArray(200, 100, userAls_line)
			}
			//增加词特征
			userWordMap := UserAls.UserWordProfile
			wordsCount := len(mem.MomentsText)
			words := factory.Segmenter.Cut(mem.MomentsText)
			if wordsCount > 0 {
				min_num := math.Min(float64(len(words)), 10.0)
				for i := 0; i < int(min_num); i++ {
					if _, ok := userWordMap[words[i]]; ok {
						fs.Add(i+50, userWordMap[words[i]])
					}
				}
			}
			//增加topword词偏好 800-1100
			for i := 0; i < len(words); i++ {
				if index_num, ok := wordVec[words[i]]; ok {
						if value,ok :=userWordMap[words[i]]; ok {
							fs.Add(int(index_num[0])+800,value)
						}
					}
				}
			}

		}


	//ALS话题向量
	if (data.ThemeProfile != nil) {
		ThemeAls := data.ThemeProfile
		themeAls_line := ThemeAls.ThemeEmbedding
		if len(themeAls_line) > 0 {
			fs.AddArray(400, 100, themeAls_line)
		}
	}

	imageUrl := mem.ImageUrl
	if len(imageUrl) > 0 {
		fs.Add(8, float32(1.0))
	} else {
		fs.Add(8, float32(0.0))
	}


	return fs

}