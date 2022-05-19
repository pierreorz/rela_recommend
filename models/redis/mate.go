package redis

import (
	"encoding/json"
	"rela_recommend/cache"
	"rela_recommend/factory"
	"rela_recommend/log"
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


// 读取假装情侣在线用户
func (self *PretendLoveUserModule) QueryPretendLoveList()  ([]PretendLoveUser,error) {
	keyFormatter := "chat:waiting_users"
	user_bytes, err := factory.AwsCluster.LRange(keyFormatter, 0, -1)
	log.Infof("user_bytes=====================%+v", user_bytes)
	users := []PretendLoveUser{}
	if len(user_bytes)>0 {
		log.Infof("user_bytes=====================%+v", user_bytes)
		for i := 0; i < len(user_bytes); i++ {
			user_byte := user_bytes[i]
			log.Infof("user_byte=====================%+v", user_byte)
			if user_byte != nil && len(user_byte) > 0 {
				user:=PretendLoveUser{}
				if err := json.Unmarshal(user_byte, &user); err != nil {
					log.Error(err.Error(), string(user_byte))
				}else{
					log.Infof("user=====================%+v", user)
					users = append(users, user)
				}
			}
		}
		log.Infof("pretend======%+v",users)
		return users, err
	}
	return users, err
}





