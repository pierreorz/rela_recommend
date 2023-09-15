package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"rela_recommend/cache"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/service/abtest"
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

type TextTypeCategText struct{
	TextTypeid int64  `json:"text_type_id"`
	TextTypeName string  `json:"text_type_name"`
	CategTypeid int64 `json:"categ_type_id"`
	CategTypeName string `json:"categ_type_name"`
	TextLine string `json:"text_line"`
}

type BehaviorMate struct {
	ReqUserID int64 `json:"reqUserId"`
	Data      []struct {
		ID       int64 `json:"id"`
		TagType  int64 `json:"tagType"`
		TextType int64 `json:"textType"`
		UserID   int64 `json:"userId"`
	} `json:"data"`
}


type MataCategTextModule struct{
	CachePikaModule
}

func NewMateCacheModule(cache *cache.Cache, store *cache.Cache) *PretendLoveUserModule {
	return &PretendLoveUserModule{cacheCluster: *cache, storeCluster: *store}
}


func NewMateCaegtCacheModule(ctx abtest.IAbTestAble,cache *cache.Cache, store *cache.Cache) *MataCategTextModule {
	return &MataCategTextModule{CachePikaModule{ctx: ctx, cache: *cache, store: *store}}
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
	if err==nil{
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
func(this *UserCacheModule) QueryUserBaseMap(userId int64,userIds []int64) (*UserProfile,map[int64]*UserProfile, error){
	var user *UserProfile
	var userMap map[int64]*UserProfile
	user,userMap,userCacheErr:=this.QueryByUserAndUsersMap(userId,userIds)
	if userCacheErr==nil{
		return user,userMap,nil
	}
	return user,userMap,nil
}
//获取文案信息
func (this *MataCategTextModule) QueryMateUserCategTextList(textType int64,categType int64) (TextTypeCategText,error){
	keyFormatter := fmt.Sprintf("mate_text:text_type:%d:categ_type:%d", textType,categType)
	var categText TextTypeCategText
	err := this.GetSetStruct(keyFormatter,&categText,6*60*60, 1*60*60)
	return categText, err
}
//获取redis
func (this *MataCategTextModule) QueryMatebehaviorMap(userId int64)(BehaviorMate,error){
	keyFormatter := fmt.Sprintf("mate_behavior:%d", userId)
	var behaviorMap BehaviorMate
	err := this.GetSetStruct(keyFormatter,&behaviorMap,6*60*60, 1*60*60)
	return behaviorMap,err
}



//请求写入pika数据，
func (this *MataCategTextModule) SetMataUserPikaMap(userId int64,res []byte,cacheTime2 int)(int,error) {
	keyFormatter := fmt.Sprintf("mate_result:pika_user:%d", userId)
	if res!=nil {
		err := this.store.SetEx(keyFormatter, res, cacheTime2)
		if err != nil {
			log.Infof("setcache warn: %s\n",err)
		}
	}
	return len(res),nil
}
//获取10分钟内pika数据
func (this *MataCategTextModule) QueryMataUserPikaMap(userId int64)(string,error) {
	keyFormatter := fmt.Sprintf("mate_result:pika_user:%d", userId)
	var matePikaString string
	a, _ :=this.store.Get(keyFormatter)

	log.Info("",a)
	matePikaString,err := redis.String(this.store.Get(keyFormatter))

	log.Info("get pika matePikaString================",matePikaString)
	log.Info("get pika err================",err)
	return matePikaString,err
}
