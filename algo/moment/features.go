package moment

import (
	"rela_recommend/factory"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
	"time"
	"strings"
	"strconv"
)


func GetMomentFeatures(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}
	shareToList := map[string]int{"all": 0, "friends": 1, "self": 2}

	data := idata.(*DataInfo)
	// 发布内容
	mem, meme := data.MomentCache, data.MomentExtendCache
	memprofile := data.MomentProfile
	fs.AddCategory(1, 24, 0, mem.InsertTime.Hour(), 0)
	fs.AddCategory(30, 7, 0, int(mem.InsertTime.Weekday()), 0)
	if shareTo, ok := shareToList[mem.ShareTo]; ok {
		fs.AddCategory(40, 3, 0, shareTo, 0)
	}
	// 		是否有 内容，图片，视频
	wordsCount := len(mem.MomentsText)
	fs.AddCategory(44, 2, 0, rutils.GetInt(wordsCount > 0), 0)
	fs.AddCategory(46, 2, 0, rutils.GetInt(len(mem.ImageUrl) > 0), 0)
	fs.AddCategory(48, 2, 0, rutils.GetInt(len(mem.VoiceUrl) > 0), 0)
	if(len(mem.VoiceUrl) > 0){
		fs.Add(54,float32(len(strings.Split(mem.ImageUrl,","))))
	}
	if memprofile != nil {
		fs.Add(50, float32(memprofile.TextCnt))
		fs.Add(51, float32(memprofile.LikeCnt))
	}
	fs.Add(52, float32(len(mem.MomentsText)))
	fs.Add(53, float32(ctx.GetCreateTime().Sub(mem.InsertTime).Minutes()))
	fs.Add(55,float32(time.Now().Hour()))

	// 发布者
	memu := data.UserCache
	var role, wantRoles = 0, make([]int, 0)
	if(memu != nil) {
		fs.Add(2000, float32(ctx.GetCreateTime().Sub(memu.CreateTime.Time).Seconds() / 60 / 60 / 24))
		fs.Add(2001, float32(memu.Age))
		fs.Add(2002, float32(memu.Height))
		fs.Add(2003, float32(memu.Weight))

		fs.AddCategory(2010, 13, -1, rutils.GetInt(memu.Horoscope), -1)	// 星座
		fs.AddCategory(2030, 10, -1, memu.Affection, -1)				// 单身情况
		role, wantRoles = rutils.GetInt(memu.RoleName), rutils.GetInts(memu.WantRole)
		fs.AddCategory(2040, 10, -1, role, -1)				// 自我认同
		fs.AddCategories(2050, 10, -1, wantRoles, -1)	// 想要寻找
	}
	memuEmbedding:=data.UserEmbedding
	if(memuEmbedding!=nil){
		Embedding:=strings.Split(memuEmbedding.Embedding,",")
		for i:=0;i<128;i++{
			v,err:=strconv.ParseFloat(Embedding[i], 32)
			if err==nil{
				fs.Add(3000+i,float32(v))
			}
		}
	}

	// 观看者
	if ctx.GetUserInfo() != nil {
		user := ctx.GetUserInfo().(*UserInfo)
		if user.UserCache != nil {
			curr := user.UserCache
			fs.Add(4000, float32(ctx.GetCreateTime().Sub(curr.CreateTime.Time).Seconds() / 60 / 60 / 24))
			fs.Add(4001, float32(curr.Age))
			fs.Add(4002, float32(curr.Height))
			fs.Add(4003, float32(curr.Weight))
			fs.Add(4004, float32(rutils.EarthDistance(float64(ctx.GetRequest().Lng), float64(ctx.GetRequest().Lat), meme.Lng, meme.Lat)))

			fs.AddCategory(4010, 13, -1, rutils.GetInt(curr.Horoscope), -1)	// 星座
			fs.AddCategory(4030, 10, -1, curr.Affection, -1)				// 单身情况
			uRole, uWantRoles := rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
			fs.AddCategory(4040, 10, -1, uRole, -1)				// 自我认同
			fs.AddCategories(4050, 10, -1, uWantRoles, -1)	// 想要寻找
			// 交叉
			fs.AddCategory(6000, 2, 0, rutils.GetInt(rutils.IsInInts(role, uWantRoles)), 0)
			fs.AddCategory(6002, 2, 0, rutils.GetInt(rutils.IsInInts(uRole, wantRoles)), 0)
		}
		if user.UserEmbedding!=nil{
			wmem:=user.UserEmbedding
			Embedding:=strings.Split(wmem.Embedding,",")
			for i:=0;i<128;i++{
				v,err:=strconv.ParseFloat(Embedding[i], 32)
				if err==nil{
					fs.Add(5100+i,float32(v))
				}
			}
		}
		if user.MomentProfile!=nil{
			matp:=user.MomentProfile
			if matp.AgeMap != nil {
				fs.Add(5000, matp.AgeMap["age_18_20"])
				fs.Add(5001, matp.AgeMap["age_21_22"])
				fs.Add(5002, matp.AgeMap["age_23_24"])
				fs.Add(5003, matp.AgeMap["age_25_26"])
				fs.Add(5004, matp.AgeMap["age_27_29"])
				fs.Add(5005, matp.AgeMap["age_above_30"])
				fs.Add(5006, matp.AgeMap["age_unknown"])
			}
			if matp.RoleNameMap != nil {
				fs.Add(5007, matp.RoleNameMap["role_name_t"])
				fs.Add(5008, matp.RoleNameMap["role_name_p"])
				fs.Add(5009, matp.RoleNameMap["role_name_h"])
				fs.Add(5010, matp.RoleNameMap["role_name_bi"])
				fs.Add(5011, matp.RoleNameMap["role_name_other"])
				fs.Add(5012, matp.RoleNameMap["role_name_str"])
				fs.Add(5013, matp.RoleNameMap["role_name_fu"])
				fs.Add(5014, matp.RoleNameMap["role_name_unknown"])
			}
			if matp.HoroscopeMap != nil {
				fs.Add(5015, matp.HoroscopeMap["horoscope_cap"])
				fs.Add(5016, matp.HoroscopeMap["horoscope_aqua"])
				fs.Add(5017, matp.HoroscopeMap["horoscope_pis"])
				fs.Add(5018, matp.HoroscopeMap["horoscope_ar"])
				fs.Add(5019, matp.HoroscopeMap["horoscope_tau"])
				fs.Add(5020, matp.HoroscopeMap["horoscope_gemini"])
				fs.Add(5021, matp.HoroscopeMap["horoscope_cancer"])
				fs.Add(5022, matp.HoroscopeMap["horoscope_leo"])
				fs.Add(5023, matp.HoroscopeMap["horoscope_virgo"])
				fs.Add(5024, matp.HoroscopeMap["horoscope_libra"])
				fs.Add(5025, matp.HoroscopeMap["horoscope_scor"])
				fs.Add(5026, matp.HoroscopeMap["horoscope_sagi"])
				fs.Add(5027, matp.HoroscopeMap["horoscope_unknown"])
			}
			if matp.HeightMap != nil {
				fs.Add(5028, matp.HeightMap["height_under_155"])
				fs.Add(5029, matp.HeightMap["height_156_160"])
				fs.Add(5030, matp.HeightMap["height_161_163"])
				fs.Add(5031, matp.HeightMap["height_164_166"])
				fs.Add(5032, matp.HeightMap["height_167_170"])
				fs.Add(5033, matp.HeightMap["height_171_180"])
				fs.Add(5034, matp.HeightMap["height_above_180"])
				fs.Add(5035, matp.HeightMap["height_unknown"])
			}
			if matp.WeightMap != nil {
				fs.Add(5036, matp.WeightMap["weight_under_41"])
				fs.Add(5037, matp.WeightMap["weight_42_45"])
				fs.Add(5038, matp.WeightMap["weight_46_49"])
				fs.Add(5039, matp.WeightMap["weight_50_52"])
				fs.Add(5040, matp.WeightMap["weight_53_57"])
				fs.Add(5041, matp.WeightMap["weight_above_58"])
				fs.Add(5042, matp.WeightMap["weight_unknown"])
			}
		}

	}
	
	// 分词结果
	if wordsCount > 0 {
		var words = make([]string, 0)
		if memprofile != nil {
			words = memprofile.MomentsTextWords
		} else {
			words = factory.Segmenter.Cut(mem.MomentsText)
		}
		words = model.CheckWords(words)
		fs.AddHashStrings(100000, 100000, words)
	}
	return fs
}
