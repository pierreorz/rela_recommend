package redis

import (
	"rela_recommend/cache"
	"rela_recommend/log"
	"rela_recommend/service/abtest"
)

type PretendLoveUser struct{
	PretendLoveUserList []string `json:"pretend_love_user_list"` //假装情侣用户列表
}

type PretendLoveUserModule struct {
	CachePikaModule
}


func NewMateCacheModule(ctx abtest.IAbTestAble, cache *cache.Cache, store *cache.Cache) *PretendLoveUserModule {
	return &PretendLoveUserModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
}


// 读取假装情侣在线用户
func (self *PretendLoveUserModule) QueryPretendLoveList()  (PretendLoveUser,error) {
	keyFormatter := "chat:waiting_user"
	awsList, err := self.Get(keyFormatter)
	log.Infof("awsList=====================%+v",awsList)
	obj:=awsList.(PretendLoveUser)
	return obj, err

}





