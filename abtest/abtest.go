package abtest

import (
	"crypto/md5"
	"rela_recommend/utils"
)

func GetMd5Int64(userId int64) int64 {
	idString := utils.GetString(userId)

	md5New := md5.New()
	md5New.Write([]byte(idString))
	bytes := md5New.Sum(nil)
	
	return utils.BytesToInt64(bytes)
}

func IsSwitched(userId int64, proba int64) bool {
	md5Val := GetMd5Int64(userId)
	return md5Val % 100 < proba
}