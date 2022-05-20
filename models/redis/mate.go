package redis

import (
	"rela_recommend/cache"
	"rela_recommend/factory"
	"strings"
)

type PretendLoveUser struct{
	Userid string `json:"user_id"` //假装情侣用户id
	Socketid string `json:"socket_id"` //
	Roleid string `json:"role_id"` //角色id
}

type PretendLoveUserModule struct {
	cacheCluster cache.Cache
	storeCluster cache.Cache
}

func NewMateCacheModule(cache *cache.Cache, store *cache.Cache) *PretendLoveUserModule {
	return &PretendLoveUserModule{cacheCluster: *cache, storeCluster: *store}
}
func SetPretendLoveUser(Userid string,Socketid string,Roleid string ) PretendLoveUser{
	return PretendLoveUser{
		Userid ,Socketid,Roleid,
	}
}

// 读取假装情侣在线用户
func (self *PretendLoveUserModule) QueryPretendLoveList()  ([]PretendLoveUser,error) {
	keyFormatter := "chat:waiting_users"
	user_bytes, err := factory.AwsCluster.LRange(keyFormatter, 0, -1)
	users := []PretendLoveUser{}
	if len(user_bytes)>0 {
		for i := 0; i < len(user_bytes); i++ {
			user_byte := user_bytes[i]
			if user_byte != nil && len(user_byte) > 0 {
				userLine:=string(user_byte)
				userList:=strings.Split(userLine, ",")
				if len(userList)==3{
					userPretend:=SetPretendLoveUser(userList[0],userList[1],userList[2])
					users=append(users, userPretend)
				}
			}
		}
		//log.Infof("pretend======%+v",users)
		return users, err
	}
	return users, err
}
//获取假装情侣在线用户信息
func(this *UserCacheModule) QuserUserBaseMap(userId int64,userIds []int64) (*UserProfile,map[int64]*UserProfile, error){
	var user *UserProfile
	var userMap map[int64]*UserProfile
	user,userMap,userCacheErr:=this.QueryByUserAndUsersMap(userId,userIds)
	if userCacheErr==nil{
		return user,userMap,nil
	}
	return user,userMap,nil
}
//获取假装情侣在线用户话题偏好




