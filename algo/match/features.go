package match

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/models/redis"
	"rela_recommend/service"
	rutils "rela_recommend/utils"
)

// func GetMatchFeatures(userInfo *redis.UserProfile, userInfo2 *redis.UserProfile, dataMatch *redis.MatchProfile, dataMatch2 *redis.MatchProfile) *utils.Features {
// 	memu := userInfo
// 	matp := dataMatch
func GetMatchFeatures(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}
	data := idata.(*DataInfo)
	currTime := ctx.GetCreateTime().Unix()

	// 用户
	var role, wantRoles = 0, make([]int, 0)
	var memu *redis.UserProfile
	if ctx.GetUserInfo() != nil {
		user := ctx.GetUserInfo().(*UserInfo)
		if user.UserCache != nil {
			memu = user.UserCache
			fs.Add(1, float32(memu.Age))
			fs.Add(2, float32(memu.Height))
			fs.Add(3, float32(memu.Weight))
			fs.Add(4, float32(currTime-memu.LastUpdateTime))
			fs.AddCategory(10, 13, -1, rutils.GetInt(memu.Horoscope), -1)
			fs.AddCategory(30, 10, -1, memu.Affection, -1)
			role, wantRoles = rutils.GetInt(memu.RoleName), rutils.GetInts(memu.WantRole)
			fs.AddCategory(50, 10, -1, role, -1)
			fs.AddCategories(70, 10, -1, wantRoles, -1)
			fs.AddCategory(2115, 2, 0, rutils.GetInt(memu.IsVip), 0)
		}
		// 用户画像
		if user.MatchProfile != nil {
			matp := user.MatchProfile
			if matp.AgeMap != nil {
				fs.Add(2000, matp.AgeMap["age_18_20"])
				fs.Add(2001, matp.AgeMap["age_21_22"])
				fs.Add(2002, matp.AgeMap["age_23_24"])
				fs.Add(2003, matp.AgeMap["age_25_26"])
				fs.Add(2004, matp.AgeMap["age_27_29"])
				fs.Add(2005, matp.AgeMap["age_above_30"])
				fs.Add(2006, matp.AgeMap["age_unknown"])
			}
			if matp.RoleNameMap != nil {
				fs.Add(2007, matp.RoleNameMap["role_name_t"])
				fs.Add(2008, matp.RoleNameMap["role_name_p"])
				fs.Add(2009, matp.RoleNameMap["role_name_h"])
				fs.Add(2010, matp.RoleNameMap["role_name_bi"])
				fs.Add(2011, matp.RoleNameMap["role_name_other"])
				fs.Add(2012, matp.RoleNameMap["role_name_str"])
				fs.Add(2013, matp.RoleNameMap["role_name_fu"])
				fs.Add(2014, matp.RoleNameMap["role_name_unknown"])
			}
			if matp.HoroscopeMap != nil {
				fs.Add(2015, matp.HoroscopeMap["horoscope_cap"])
				fs.Add(2016, matp.HoroscopeMap["horoscope_aqua"])
				fs.Add(2017, matp.HoroscopeMap["horoscope_pis"])
				fs.Add(2018, matp.HoroscopeMap["horoscope_ar"])
				fs.Add(2019, matp.HoroscopeMap["horoscope_tau"])
				fs.Add(2020, matp.HoroscopeMap["horoscope_gemini"])
				fs.Add(2021, matp.HoroscopeMap["horoscope_cancer"])
				fs.Add(2022, matp.HoroscopeMap["horoscope_leo"])
				fs.Add(2023, matp.HoroscopeMap["horoscope_virgo"])
				fs.Add(2024, matp.HoroscopeMap["horoscope_libra"])
				fs.Add(2025, matp.HoroscopeMap["horoscope_scor"])
				fs.Add(2026, matp.HoroscopeMap["horoscope_sagi"])
				fs.Add(2027, matp.HoroscopeMap["horoscope_unknown"])
			}
			if matp.HeightMap != nil {
				fs.Add(2028, matp.HeightMap["height_under_155"])
				fs.Add(2029, matp.HeightMap["height_156_160"])
				fs.Add(2030, matp.HeightMap["height_161_163"])
				fs.Add(2031, matp.HeightMap["height_164_166"])
				fs.Add(2032, matp.HeightMap["height_167_170"])
				fs.Add(2033, matp.HeightMap["height_171_180"])
				fs.Add(2034, matp.HeightMap["height_above_180"])
				fs.Add(2035, matp.HeightMap["height_unknown"])
			}
			if matp.WeightMap != nil {
				fs.Add(2036, matp.WeightMap["weight_under_41"])
				fs.Add(2037, matp.WeightMap["weight_42_45"])
				fs.Add(2038, matp.WeightMap["weight_46_49"])
				fs.Add(2039, matp.WeightMap["weight_50_52"])
				fs.Add(2040, matp.WeightMap["weight_53_57"])
				fs.Add(2041, matp.WeightMap["weight_above_58"])
				fs.Add(2042, matp.WeightMap["weight_unknown"])
			}
			if matp.DistanceMap != nil {
				fs.Add(2043, matp.DistanceMap["dis_under_20"])
				fs.Add(2044, matp.DistanceMap["dis_21_40"])
				fs.Add(2045, matp.DistanceMap["dis_41_60"])
				fs.Add(2046, matp.DistanceMap["dis_61_80"])
				fs.Add(2047, matp.DistanceMap["dis_81_100"])
				fs.Add(2048, matp.DistanceMap["dis_101_200"])
				fs.Add(2049, matp.DistanceMap["dis_201_300"])
				fs.Add(2050, matp.DistanceMap["dis_301_400"])
				fs.Add(2051, matp.DistanceMap["dis_401_500"])
				fs.Add(2052, matp.DistanceMap["dis_above_500"])
			}
			if matp.DistanceMap != nil {
				fs.Add(2053, matp.LikeTypeMap["like_type_like"])
				fs.Add(2054, matp.LikeTypeMap["like_type_dislike"])
				fs.Add(2055, matp.LikeTypeMap["like_type_superlike"])
			}
			if matp.AffectionMap != nil {
				fs.Add(2056, matp.AffectionMap["affection_single"])
				fs.Add(2057, matp.AffectionMap["affection_dating"])
				fs.Add(2058, matp.AffectionMap["affection_stable"])
				fs.Add(2059, matp.AffectionMap["affection_married"])
				fs.Add(2060, matp.AffectionMap["affection_open_re"])
				fs.Add(2061, matp.AffectionMap["affection_relationship"])
				fs.Add(2062, matp.AffectionMap["affection_waiting"])
				fs.Add(2063, matp.AffectionMap["affection_secret"])
			}
			if matp.MobileSysMap != nil {
				fs.Add(2064, matp.MobileSysMap["mobile_sys_ios"])
				fs.Add(2065, matp.MobileSysMap["mobile_sys_android"])
			}
			if matp.TotalCount >= 0 {
				fs.Add(2066, float32(matp.TotalCount))
			}
			if matp.FreqWeekMap != nil {
				fs.Add(2067, matp.FreqWeekMap["monday"])
				fs.Add(2068, matp.FreqWeekMap["tuesday"])
				fs.Add(2069, matp.FreqWeekMap["wednesday"])
				fs.Add(2070, matp.FreqWeekMap["thursday"])
				fs.Add(2071, matp.FreqWeekMap["friday"])
				fs.Add(2072, matp.FreqWeekMap["saturday"])
				fs.Add(2073, matp.FreqWeekMap["sunday"])
			}
			if matp.FreqTimeMap != nil {
				fs.Add(2074, matp.FreqTimeMap["time_0_2"])
				fs.Add(2075, matp.FreqTimeMap["time_2_4"])
				fs.Add(2076, matp.FreqTimeMap["time_4_6"])
				fs.Add(2077, matp.FreqTimeMap["time_6_8"])
				fs.Add(2078, matp.FreqTimeMap["time_8_10"])
				fs.Add(2079, matp.FreqTimeMap["time_10_12"])
				fs.Add(2080, matp.FreqTimeMap["time_12_14"])
				fs.Add(2081, matp.FreqTimeMap["time_14_16"])
				fs.Add(2082, matp.FreqTimeMap["time_16_18"])
				fs.Add(2083, matp.FreqTimeMap["time_18_20"])
				fs.Add(2084, matp.FreqTimeMap["time_20_22"])
				fs.Add(2085, matp.FreqTimeMap["time_22_24"])
			}
			if matp.ContinuesUse >= 0 {
				fs.Add(2086, float32(matp.ContinuesUse))
			}
			if matp.ImageMap != nil {
				fs.AddCategory(2087, 2, 0, rutils.GetInt(matp.ImageMap["cover_has_face"]), 0)
				fs.AddCategory(2090, 2, 0, rutils.GetInt(matp.ImageMap["head_has_face"]), 0)
				fs.AddCategory(2095, 2, 0, rutils.GetInt(matp.ImageMap["imagewall_has_face"]), 0)
				fs.AddCategory(2100, 2, 0, rutils.GetInt(matp.ImageMap["has_cover"]), 0)
				fs.Add(2110, matp.ImageMap["imagewall_count"])
			}
			if matp.MomentMap != nil {
				fs.Add(2111, matp.MomentMap["moments_count"])
			}
			if matp.UserInfoMap != nil {
				fs.AddCategory(40, 10, -1, rutils.GetInt(matp.UserInfoMap["want_affection"]), -1)
			}

		}
	}

	curr := data.UserCache
	currMatch := data.MatchProfile

	if curr != nil {
		// if userInfo2 != nil {
		// 	curr := userInfo2
		fs.Add(4000, float32(curr.Age))
		fs.Add(4001, float32(curr.Height))
		fs.Add(4002, float32(curr.Weight))
		fs.Add(4003, float32(currTime-curr.LastUpdateTime))
		if memu != nil {
			fs.Add(4004, float32(rutils.EarthDistance(memu.Location.Lon, memu.Location.Lat, curr.Location.Lon, curr.Location.Lat)/1000.0))
		}
		fs.AddCategory(4010, 13, -1, rutils.GetInt(curr.Horoscope), -1)
		fs.AddCategory(4030, 10, -1, curr.Affection, -1)
		uRole, uWantRoles := rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
		fs.AddCategory(4050, 10, -1, uRole, -1) // 自我认同
		fs.AddCategories(4070, 10, -1, uWantRoles, -1)
		fs.AddCategory(5115, 2, 0, rutils.GetInt(curr.IsVip), 0)

		// 交叉
		fs.AddCategory(6000, 2, 0, rutils.GetInt(role > 0 && rutils.IsInInts(role, uWantRoles)), 0)
		fs.AddCategory(6002, 2, 0, rutils.GetInt(uRole > 0 && rutils.IsInInts(uRole, wantRoles)), 0)
	}
	// if dataMatch2 != nil {
	if currMatch != nil {
		// currMatch := dataMatch2
		if currMatch.AgeMap != nil {
			fs.Add(5000, currMatch.AgeMap["age_18_20"])
			fs.Add(5001, currMatch.AgeMap["age_21_22"])
			fs.Add(5002, currMatch.AgeMap["age_23_24"])
			fs.Add(5003, currMatch.AgeMap["age_25_26"])
			fs.Add(5004, currMatch.AgeMap["age_27_29"])
			fs.Add(5005, currMatch.AgeMap["age_above_30"])
			fs.Add(5006, currMatch.AgeMap["age_unknown"])
		}
		if currMatch.RoleNameMap != nil {
			fs.Add(5007, currMatch.RoleNameMap["role_name_t"])
			fs.Add(5008, currMatch.RoleNameMap["role_name_p"])
			fs.Add(5009, currMatch.RoleNameMap["role_name_h"])
			fs.Add(5010, currMatch.RoleNameMap["role_name_bi"])
			fs.Add(5011, currMatch.RoleNameMap["role_name_other"])
			fs.Add(5012, currMatch.RoleNameMap["role_name_str"])
			fs.Add(5013, currMatch.RoleNameMap["role_name_fu"])
			fs.Add(5014, currMatch.RoleNameMap["role_name_unknown"])
		}
		if currMatch.HoroscopeMap != nil {
			fs.Add(5015, currMatch.HoroscopeMap["horoscope_cap"])
			fs.Add(5016, currMatch.HoroscopeMap["horoscope_aqua"])
			fs.Add(5017, currMatch.HoroscopeMap["horoscope_pis"])
			fs.Add(5018, currMatch.HoroscopeMap["horoscope_ar"])
			fs.Add(5019, currMatch.HoroscopeMap["horoscope_tau"])
			fs.Add(5020, currMatch.HoroscopeMap["horoscope_gemini"])
			fs.Add(5021, currMatch.HoroscopeMap["horoscope_cancer"])
			fs.Add(5022, currMatch.HoroscopeMap["horoscope_leo"])
			fs.Add(5023, currMatch.HoroscopeMap["horoscope_virgo"])
			fs.Add(5024, currMatch.HoroscopeMap["horoscope_libra"])
			fs.Add(5025, currMatch.HoroscopeMap["horoscope_scor"])
			fs.Add(5026, currMatch.HoroscopeMap["horoscope_sagi"])
			fs.Add(5027, currMatch.HoroscopeMap["horoscope_unknown"])
		}
		if currMatch.HeightMap != nil {
			fs.Add(5028, currMatch.HeightMap["height_under_155"])
			fs.Add(5029, currMatch.HeightMap["height_156_160"])
			fs.Add(5030, currMatch.HeightMap["height_161_163"])
			fs.Add(5031, currMatch.HeightMap["height_164_166"])
			fs.Add(5032, currMatch.HeightMap["height_167_170"])
			fs.Add(5033, currMatch.HeightMap["height_171_180"])
			fs.Add(5034, currMatch.HeightMap["height_above_180"])
			fs.Add(5035, currMatch.HeightMap["height_unknown"])
		}
		if currMatch.WeightMap != nil {
			fs.Add(5036, currMatch.WeightMap["weight_under_41"])
			fs.Add(5037, currMatch.WeightMap["weight_42_45"])
			fs.Add(5038, currMatch.WeightMap["weight_46_49"])
			fs.Add(5039, currMatch.WeightMap["weight_50_52"])
			fs.Add(5040, currMatch.WeightMap["weight_53_57"])
			fs.Add(5041, currMatch.WeightMap["weight_above_58"])
			fs.Add(5042, currMatch.WeightMap["weight_unknown"])
		}
		if currMatch.DistanceMap != nil {
			fs.Add(5043, currMatch.DistanceMap["dis_under_20"])
			fs.Add(5044, currMatch.DistanceMap["dis_21_40"])
			fs.Add(5045, currMatch.DistanceMap["dis_41_60"])
			fs.Add(5046, currMatch.DistanceMap["dis_61_80"])
			fs.Add(5047, currMatch.DistanceMap["dis_81_100"])
			fs.Add(5048, currMatch.DistanceMap["dis_101_200"])
			fs.Add(5049, currMatch.DistanceMap["dis_201_300"])
			fs.Add(5050, currMatch.DistanceMap["dis_301_400"])
			fs.Add(5051, currMatch.DistanceMap["dis_401_500"])
			fs.Add(5052, currMatch.DistanceMap["dis_above_500"])
		}
		if currMatch.LikeTypeMap != nil {
			fs.Add(5053, currMatch.LikeTypeMap["like_type_like"])
			fs.Add(5054, currMatch.LikeTypeMap["like_type_dislike"])
			fs.Add(5055, currMatch.LikeTypeMap["like_type_superlike"])
		}
		if currMatch.AffectionMap != nil {
			fs.Add(5056, currMatch.AffectionMap["affection_single"])
			fs.Add(5057, currMatch.AffectionMap["affection_dating"])
			fs.Add(5058, currMatch.AffectionMap["affection_stable"])
			fs.Add(5059, currMatch.AffectionMap["affection_married"])
			fs.Add(5060, currMatch.AffectionMap["affection_open_re"])
			fs.Add(5061, currMatch.AffectionMap["affection_relationship"])
			fs.Add(5062, currMatch.AffectionMap["affection_waiting"])
			fs.Add(5063, currMatch.AffectionMap["affection_secret"])
		}
		if currMatch.MobileSysMap != nil {
			fs.Add(5064, currMatch.MobileSysMap["mobile_sys_ios"])
			fs.Add(5065, currMatch.MobileSysMap["mobile_sys_android"])
		}
		if currMatch.TotalCount >= 0 {
			fs.Add(5066, float32(currMatch.TotalCount))
		}
		if currMatch.FreqWeekMap != nil {
			fs.Add(5067, currMatch.FreqWeekMap["monday"])
			fs.Add(5068, currMatch.FreqWeekMap["tuesday"])
			fs.Add(5069, currMatch.FreqWeekMap["wednesday"])
			fs.Add(5070, currMatch.FreqWeekMap["thursday"])
			fs.Add(5071, currMatch.FreqWeekMap["friday"])
			fs.Add(5072, currMatch.FreqWeekMap["saturday"])
			fs.Add(5073, currMatch.FreqWeekMap["sunday"])
		}
		if currMatch.FreqTimeMap != nil {
			fs.Add(5074, currMatch.FreqTimeMap["time_0_2"])
			fs.Add(5075, currMatch.FreqTimeMap["time_2_4"])
			fs.Add(5076, currMatch.FreqTimeMap["time_4_6"])
			fs.Add(5077, currMatch.FreqTimeMap["time_6_8"])
			fs.Add(5078, currMatch.FreqTimeMap["time_8_10"])
			fs.Add(5079, currMatch.FreqTimeMap["time_10_12"])
			fs.Add(5080, currMatch.FreqTimeMap["time_12_14"])
			fs.Add(5081, currMatch.FreqTimeMap["time_14_16"])
			fs.Add(5082, currMatch.FreqTimeMap["time_16_18"])
			fs.Add(5083, currMatch.FreqTimeMap["time_18_20"])
			fs.Add(5084, currMatch.FreqTimeMap["time_20_22"])
			fs.Add(5085, currMatch.FreqTimeMap["time_22_24"])
		}
		if currMatch.ContinuesUse >= 0 {
			fs.Add(5086, float32(currMatch.ContinuesUse))
		}
		if currMatch.ImageMap != nil {
			fs.AddCategory(5087, 2, 0, rutils.GetInt(currMatch.ImageMap["cover_has_face"]), 0)
			fs.AddCategory(5090, 2, 0, rutils.GetInt(currMatch.ImageMap["head_has_face"]), 0)
			fs.AddCategory(5095, 2, 0, rutils.GetInt(currMatch.ImageMap["imagewall_has_face"]), 0)
			fs.AddCategory(5100, 2, 0, rutils.GetInt(currMatch.ImageMap["has_cover"]), 0)
			fs.Add(5110, currMatch.ImageMap["imagewall_count"])
		}
		if currMatch.MomentMap != nil {
			fs.Add(5111, currMatch.MomentMap["moments_count"])
		}
		if currMatch.UserInfoMap != nil {
			fs.AddCategory(4040, 10, -1, rutils.GetInt(currMatch.UserInfoMap["want_affection"]), -1)
		}

	}
	return fs
}

func GetMatchFeaturesv1(ctx algo.IContext, model algo.IAlgo, idata algo.IDataInfo) *utils.Features {
	fs := &utils.Features{}
	data := idata.(*DataInfo)
	currTime := ctx.GetCreateTime().Unix()

	// 用户
	var role, wantRoles = 0, make([]int, 0)
	var memu *redis.UserProfile
	if ctx.GetUserInfo() != nil {
		user := ctx.GetUserInfo().(*UserInfo)
		if user.UserCache != nil {
			memu = user.UserCache
			fs.Add(1, float32(memu.Age))
			fs.Add(2, float32(memu.Height))
			fs.Add(3, float32(memu.Weight))
			fs.Add(4, float32(currTime-memu.LastUpdateTime))
			fs.AddCategory(10, 13, -1, rutils.GetInt(memu.Horoscope), -1)
			fs.AddCategory(30, 10, -1, memu.Affection, -1)
			role, wantRoles = rutils.GetInt(memu.RoleName), rutils.GetInts(memu.WantRole)
			fs.AddCategory(50, 10, -1, role, -1)
			fs.AddCategories(70, 10, -1, wantRoles, -1)
			fs.AddCategory(80, 2, 0, rutils.GetInt(memu.IsVip), 0)
		}
		// 用户画像
		if user.MatchProfile != nil {
			matp := user.MatchProfile
			if matp.UserInfoMap != nil {
				fs.AddCategory(40, 10, -1, rutils.GetInt(matp.UserInfoMap["want_affection"]), -1)
			}
			if matp.AgeMap != nil {
				fs.Add(2000, matp.AgeMap["age_18_20"])
				fs.Add(2001, matp.AgeMap["age_21_22"])
				fs.Add(2002, matp.AgeMap["age_23_24"])
				fs.Add(2003, matp.AgeMap["age_25_26"])
				fs.Add(2004, matp.AgeMap["age_27_29"])
				fs.Add(2005, matp.AgeMap["age_above_30"])
				fs.Add(2006, matp.AgeMap["age_unknown"])
			}
			if matp.RoleNameMap != nil {
				fs.Add(2007, matp.RoleNameMap["role_name_t"])
				fs.Add(2008, matp.RoleNameMap["role_name_p"])
				fs.Add(2009, matp.RoleNameMap["role_name_h"])
				fs.Add(2010, matp.RoleNameMap["role_name_bi"])
				fs.Add(2011, matp.RoleNameMap["role_name_other"])
				fs.Add(2012, matp.RoleNameMap["role_name_str"])
				fs.Add(2013, matp.RoleNameMap["role_name_fu"])
				fs.Add(2014, matp.RoleNameMap["role_name_unknown"])
			}
			if matp.HoroscopeMap != nil {
				fs.Add(2015, matp.HoroscopeMap["horoscope_cap"])
				fs.Add(2016, matp.HoroscopeMap["horoscope_aqua"])
				fs.Add(2017, matp.HoroscopeMap["horoscope_pis"])
				fs.Add(2018, matp.HoroscopeMap["horoscope_ar"])
				fs.Add(2019, matp.HoroscopeMap["horoscope_tau"])
				fs.Add(2020, matp.HoroscopeMap["horoscope_gemini"])
				fs.Add(2021, matp.HoroscopeMap["horoscope_cancer"])
				fs.Add(2022, matp.HoroscopeMap["horoscope_leo"])
				fs.Add(2023, matp.HoroscopeMap["horoscope_virgo"])
				fs.Add(2024, matp.HoroscopeMap["horoscope_libra"])
				fs.Add(2025, matp.HoroscopeMap["horoscope_scor"])
				fs.Add(2026, matp.HoroscopeMap["horoscope_sagi"])
				fs.Add(2027, matp.HoroscopeMap["horoscope_unknown"])
			}
			if matp.HeightMap != nil {
				fs.Add(2028, matp.HeightMap["height_under_155"])
				fs.Add(2029, matp.HeightMap["height_156_160"])
				fs.Add(2030, matp.HeightMap["height_161_163"])
				fs.Add(2031, matp.HeightMap["height_164_166"])
				fs.Add(2032, matp.HeightMap["height_167_170"])
				fs.Add(2033, matp.HeightMap["height_171_180"])
				fs.Add(2034, matp.HeightMap["height_above_180"])
				fs.Add(2035, matp.HeightMap["height_unknown"])
			}
			if matp.WeightMap != nil {
				fs.Add(2036, matp.WeightMap["weight_under_41"])
				fs.Add(2037, matp.WeightMap["weight_42_45"])
				fs.Add(2038, matp.WeightMap["weight_46_49"])
				fs.Add(2039, matp.WeightMap["weight_50_52"])
				fs.Add(2040, matp.WeightMap["weight_53_57"])
				fs.Add(2041, matp.WeightMap["weight_above_58"])
				fs.Add(2042, matp.WeightMap["weight_unknown"])
			}
			if matp.DistanceMap != nil {
				fs.Add(2043, matp.DistanceMap["dis_under_20"])
				fs.Add(2044, matp.DistanceMap["dis_21_40"])
				fs.Add(2045, matp.DistanceMap["dis_41_60"])
				fs.Add(2046, matp.DistanceMap["dis_61_80"])
				fs.Add(2047, matp.DistanceMap["dis_81_100"])
				fs.Add(2048, matp.DistanceMap["dis_101_200"])
				fs.Add(2049, matp.DistanceMap["dis_201_300"])
				fs.Add(2050, matp.DistanceMap["dis_301_400"])
				fs.Add(2051, matp.DistanceMap["dis_401_500"])
				fs.Add(2052, matp.DistanceMap["dis_above_500"])
			}
			if matp.DistanceMap != nil {
				fs.Add(2053, matp.LikeTypeMap["like_type_like"])
				fs.Add(2054, matp.LikeTypeMap["like_type_dislike"])
				fs.Add(2055, matp.LikeTypeMap["like_type_superlike"])
			}
			if matp.AffectionMap != nil {
				fs.Add(2056, matp.AffectionMap["affection_single"])
				fs.Add(2057, matp.AffectionMap["affection_dating"])
				fs.Add(2058, matp.AffectionMap["affection_stable"])
				fs.Add(2059, matp.AffectionMap["affection_married"])
				fs.Add(2060, matp.AffectionMap["affection_open_re"])
				fs.Add(2061, matp.AffectionMap["affection_relationship"])
				fs.Add(2062, matp.AffectionMap["affection_waiting"])
				fs.Add(2063, matp.AffectionMap["affection_secret"])
			}
			if matp.MobileSysMap != nil {
				fs.Add(2064, matp.MobileSysMap["mobile_sys_ios"])
				fs.Add(2065, matp.MobileSysMap["mobile_sys_android"])
			}
			if matp.TotalCount >= 0 {
				fs.Add(2066, float32(matp.TotalCount))
			}
			if matp.FreqWeekMap != nil {
				fs.Add(2067, matp.FreqWeekMap["monday"])
				fs.Add(2068, matp.FreqWeekMap["tuesday"])
				fs.Add(2069, matp.FreqWeekMap["wednesday"])
				fs.Add(2070, matp.FreqWeekMap["thursday"])
				fs.Add(2071, matp.FreqWeekMap["friday"])
				fs.Add(2072, matp.FreqWeekMap["saturday"])
				fs.Add(2073, matp.FreqWeekMap["sunday"])
			}
			if matp.FreqTimeMap != nil {
				fs.Add(2074, matp.FreqTimeMap["time_0_2"])
				fs.Add(2075, matp.FreqTimeMap["time_2_4"])
				fs.Add(2076, matp.FreqTimeMap["time_4_6"])
				fs.Add(2077, matp.FreqTimeMap["time_6_8"])
				fs.Add(2078, matp.FreqTimeMap["time_8_10"])
				fs.Add(2079, matp.FreqTimeMap["time_10_12"])
				fs.Add(2080, matp.FreqTimeMap["time_12_14"])
				fs.Add(2081, matp.FreqTimeMap["time_14_16"])
				fs.Add(2082, matp.FreqTimeMap["time_16_18"])
				fs.Add(2083, matp.FreqTimeMap["time_18_20"])
				fs.Add(2084, matp.FreqTimeMap["time_20_22"])
				fs.Add(2085, matp.FreqTimeMap["time_22_24"])
			}
			if matp.ContinuesUse >= 0 {
				fs.Add(2086, float32(matp.ContinuesUse))
			}
			if matp.ImageMap != nil {
				fs.AddCategory(2087, 2, 0, rutils.GetInt(matp.ImageMap["cover_has_face"]), 0)
				fs.AddCategory(2090, 2, 0, rutils.GetInt(matp.ImageMap["head_has_face"]), 0)
				fs.AddCategory(2095, 2, 0, rutils.GetInt(matp.ImageMap["imagewall_has_face"]), 0)
				fs.AddCategory(2100, 2, 0, rutils.GetInt(matp.ImageMap["has_cover"]), 0)
				fs.Add(2110, matp.ImageMap["imagewall_count"])
				fs.Add(2120, matp.ImageMap["cover_image_quality"])
				fs.Add(2121, matp.ImageMap["wall_image_quality"])
				fs.Add(2122, matp.ImageMap["head_image_quality"])
			}
			if matp.MomentMap != nil {
				fs.Add(2150, matp.MomentMap["moments_count"])
			}
			if matp.UserEmbedding != nil {
				fs.AddArray(3000, 128, matp.UserEmbedding["graph_embedding"])
			}
		}
	}

	curr := data.UserCache
	currMatch := data.MatchProfile

	if curr != nil {
		// if userInfo2 != nil {
		// 	curr := userInfo2
		fs.Add(4000, float32(curr.Age))
		fs.Add(4001, float32(curr.Height))
		fs.Add(4002, float32(curr.Weight))
		fs.Add(4003, float32(currTime-curr.LastUpdateTime))
		if memu != nil {
			fs.Add(4004, float32(rutils.EarthDistance(memu.Location.Lon, memu.Location.Lat, curr.Location.Lon, curr.Location.Lat)/1000.0))
		}
		fs.AddCategory(4010, 13, -1, rutils.GetInt(curr.Horoscope), -1)
		fs.AddCategory(4030, 10, -1, curr.Affection, -1)
		uRole, uWantRoles := rutils.GetInt(curr.RoleName), rutils.GetInts(curr.WantRole)
		fs.AddCategory(4050, 10, -1, uRole, -1) // 自我认同
		fs.AddCategories(4070, 10, -1, uWantRoles, -1)
		fs.AddCategory(4080, 2, 0, rutils.GetInt(curr.IsVip), 0)

		// 交叉
		fs.AddCategory(6000, 2, 0, rutils.GetInt(role > 0 && rutils.IsInInts(role, uWantRoles)), 0)
		fs.AddCategory(6002, 2, 0, rutils.GetInt(uRole > 0 && rutils.IsInInts(uRole, wantRoles)), 0)
	}
	// if dataMatch2 != nil {
	if currMatch != nil {
		// currMatch := dataMatch2
		if currMatch.UserInfoMap != nil {
			fs.AddCategory(4040, 10, -1, rutils.GetInt(currMatch.UserInfoMap["want_affection"]), -1)
		}
		if currMatch.AgeMap != nil {
			fs.Add(5000, currMatch.AgeMap["age_18_20"])
			fs.Add(5001, currMatch.AgeMap["age_21_22"])
			fs.Add(5002, currMatch.AgeMap["age_23_24"])
			fs.Add(5003, currMatch.AgeMap["age_25_26"])
			fs.Add(5004, currMatch.AgeMap["age_27_29"])
			fs.Add(5005, currMatch.AgeMap["age_above_30"])
			fs.Add(5006, currMatch.AgeMap["age_unknown"])
		}
		if currMatch.RoleNameMap != nil {
			fs.Add(5007, currMatch.RoleNameMap["role_name_t"])
			fs.Add(5008, currMatch.RoleNameMap["role_name_p"])
			fs.Add(5009, currMatch.RoleNameMap["role_name_h"])
			fs.Add(5010, currMatch.RoleNameMap["role_name_bi"])
			fs.Add(5011, currMatch.RoleNameMap["role_name_other"])
			fs.Add(5012, currMatch.RoleNameMap["role_name_str"])
			fs.Add(5013, currMatch.RoleNameMap["role_name_fu"])
			fs.Add(5014, currMatch.RoleNameMap["role_name_unknown"])
		}
		if currMatch.HoroscopeMap != nil {
			fs.Add(5015, currMatch.HoroscopeMap["horoscope_cap"])
			fs.Add(5016, currMatch.HoroscopeMap["horoscope_aqua"])
			fs.Add(5017, currMatch.HoroscopeMap["horoscope_pis"])
			fs.Add(5018, currMatch.HoroscopeMap["horoscope_ar"])
			fs.Add(5019, currMatch.HoroscopeMap["horoscope_tau"])
			fs.Add(5020, currMatch.HoroscopeMap["horoscope_gemini"])
			fs.Add(5021, currMatch.HoroscopeMap["horoscope_cancer"])
			fs.Add(5022, currMatch.HoroscopeMap["horoscope_leo"])
			fs.Add(5023, currMatch.HoroscopeMap["horoscope_virgo"])
			fs.Add(5024, currMatch.HoroscopeMap["horoscope_libra"])
			fs.Add(5025, currMatch.HoroscopeMap["horoscope_scor"])
			fs.Add(5026, currMatch.HoroscopeMap["horoscope_sagi"])
			fs.Add(5027, currMatch.HoroscopeMap["horoscope_unknown"])
		}
		if currMatch.HeightMap != nil {
			fs.Add(5028, currMatch.HeightMap["height_under_155"])
			fs.Add(5029, currMatch.HeightMap["height_156_160"])
			fs.Add(5030, currMatch.HeightMap["height_161_163"])
			fs.Add(5031, currMatch.HeightMap["height_164_166"])
			fs.Add(5032, currMatch.HeightMap["height_167_170"])
			fs.Add(5033, currMatch.HeightMap["height_171_180"])
			fs.Add(5034, currMatch.HeightMap["height_above_180"])
			fs.Add(5035, currMatch.HeightMap["height_unknown"])
		}
		if currMatch.WeightMap != nil {
			fs.Add(5036, currMatch.WeightMap["weight_under_41"])
			fs.Add(5037, currMatch.WeightMap["weight_42_45"])
			fs.Add(5038, currMatch.WeightMap["weight_46_49"])
			fs.Add(5039, currMatch.WeightMap["weight_50_52"])
			fs.Add(5040, currMatch.WeightMap["weight_53_57"])
			fs.Add(5041, currMatch.WeightMap["weight_above_58"])
			fs.Add(5042, currMatch.WeightMap["weight_unknown"])
		}
		if currMatch.DistanceMap != nil {
			fs.Add(5043, currMatch.DistanceMap["dis_under_20"])
			fs.Add(5044, currMatch.DistanceMap["dis_21_40"])
			fs.Add(5045, currMatch.DistanceMap["dis_41_60"])
			fs.Add(5046, currMatch.DistanceMap["dis_61_80"])
			fs.Add(5047, currMatch.DistanceMap["dis_81_100"])
			fs.Add(5048, currMatch.DistanceMap["dis_101_200"])
			fs.Add(5049, currMatch.DistanceMap["dis_201_300"])
			fs.Add(5050, currMatch.DistanceMap["dis_301_400"])
			fs.Add(5051, currMatch.DistanceMap["dis_401_500"])
			fs.Add(5052, currMatch.DistanceMap["dis_above_500"])
		}
		if currMatch.LikeTypeMap != nil {
			fs.Add(5053, currMatch.LikeTypeMap["like_type_like"])
			fs.Add(5054, currMatch.LikeTypeMap["like_type_dislike"])
			fs.Add(5055, currMatch.LikeTypeMap["like_type_superlike"])
		}
		if currMatch.AffectionMap != nil {
			fs.Add(5056, currMatch.AffectionMap["affection_single"])
			fs.Add(5057, currMatch.AffectionMap["affection_dating"])
			fs.Add(5058, currMatch.AffectionMap["affection_stable"])
			fs.Add(5059, currMatch.AffectionMap["affection_married"])
			fs.Add(5060, currMatch.AffectionMap["affection_open_re"])
			fs.Add(5061, currMatch.AffectionMap["affection_relationship"])
			fs.Add(5062, currMatch.AffectionMap["affection_waiting"])
			fs.Add(5063, currMatch.AffectionMap["affection_secret"])
		}
		if currMatch.MobileSysMap != nil {
			fs.Add(5064, currMatch.MobileSysMap["mobile_sys_ios"])
			fs.Add(5065, currMatch.MobileSysMap["mobile_sys_android"])
		}
		if currMatch.TotalCount >= 0 {
			fs.Add(5066, float32(currMatch.TotalCount))
		}
		if currMatch.FreqWeekMap != nil {
			fs.Add(5067, currMatch.FreqWeekMap["monday"])
			fs.Add(5068, currMatch.FreqWeekMap["tuesday"])
			fs.Add(5069, currMatch.FreqWeekMap["wednesday"])
			fs.Add(5070, currMatch.FreqWeekMap["thursday"])
			fs.Add(5071, currMatch.FreqWeekMap["friday"])
			fs.Add(5072, currMatch.FreqWeekMap["saturday"])
			fs.Add(5073, currMatch.FreqWeekMap["sunday"])
		}
		if currMatch.FreqTimeMap != nil {
			fs.Add(5074, currMatch.FreqTimeMap["time_0_2"])
			fs.Add(5075, currMatch.FreqTimeMap["time_2_4"])
			fs.Add(5076, currMatch.FreqTimeMap["time_4_6"])
			fs.Add(5077, currMatch.FreqTimeMap["time_6_8"])
			fs.Add(5078, currMatch.FreqTimeMap["time_8_10"])
			fs.Add(5079, currMatch.FreqTimeMap["time_10_12"])
			fs.Add(5080, currMatch.FreqTimeMap["time_12_14"])
			fs.Add(5081, currMatch.FreqTimeMap["time_14_16"])
			fs.Add(5082, currMatch.FreqTimeMap["time_16_18"])
			fs.Add(5083, currMatch.FreqTimeMap["time_18_20"])
			fs.Add(5084, currMatch.FreqTimeMap["time_20_22"])
			fs.Add(5085, currMatch.FreqTimeMap["time_22_24"])
		}
		if currMatch.ContinuesUse >= 0 {
			fs.Add(5086, float32(currMatch.ContinuesUse))
		}
		if currMatch.ImageMap != nil {
			fs.AddCategory(5087, 2, 0, rutils.GetInt(currMatch.ImageMap["cover_has_face"]), 0)
			fs.AddCategory(5090, 2, 0, rutils.GetInt(currMatch.ImageMap["head_has_face"]), 0)
			fs.AddCategory(5095, 2, 0, rutils.GetInt(currMatch.ImageMap["imagewall_has_face"]), 0)
			fs.AddCategory(5100, 2, 0, rutils.GetInt(currMatch.ImageMap["has_cover"]), 0)
			fs.Add(5110, currMatch.ImageMap["imagewall_count"])
			fs.Add(5120, currMatch.ImageMap["cover_image_quality"])
			fs.Add(5121, currMatch.ImageMap["wall_image_quality"])
			fs.Add(5122, currMatch.ImageMap["head_image_quality"])
		}
		if currMatch.MomentMap != nil {
			fs.Add(5150, currMatch.MomentMap["moments_count"])
		}
		if currMatch.UserEmbedding != nil {
			fs.AddArray(7000, 128, currMatch.UserEmbedding["graph_embedding"])
		}
	}

	if data.ItemBehavior != nil {
		// 点击互动
		listInteract := data.ItemBehavior.GetMatchListInteract()
		fs.Add(9000, float32(listInteract.Count))
		if listInteract.LastTime > 0 {
			fs.Add(9001, float32(float64(currTime)-listInteract.LastTime))
		}
		// 曝光
		listExposure := data.ItemBehavior.GetMatchListExposure()
		fs.Add(9002, float32(listExposure.Count))
		if listExposure.LastTime > 0 {
			fs.Add(9003, float32(float64(currTime)-listExposure.LastTime))
			fs.Add(9004, float32(listInteract.Count/listExposure.Count)) // 互动率
		}
	}
	return fs
}

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
