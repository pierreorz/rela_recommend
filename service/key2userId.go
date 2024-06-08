package service

import (
	"rela_recommend/factory"
	"rela_recommend/models"
	"rela_recommend/utils"
)

const (
	PREFIX_KEY_TO_USER       = "key:user:"
	EXPIRE_KEY_TO_USER       = 3600 * 24 * 3
	EXPIRE_LOCAL_KEY_TO_USER = 600
)

type defaultKey2UserId struct{}

func (*defaultKey2UserId) GetUserIdByKey(key string) (int64, error) {
	rdsKey := PREFIX_KEY_TO_USER + key

	data, _ := factory.CacheRds.Get(rdsKey)
	userId := utils.GetInt64(data)
	if userId != 0 {
		return userId, nil
	}

	var login models.Login
	if err := models.NewLoginModule(factory.DbW, factory.CacheRds).QueryByKey(key, &login); err != nil {
		return 0, err
	}
	//目前redis写操作不由这边控制，所以只写本地cache
	// factory.CacheLoc.SetEx(rdsKey, userId, EXPIRE_LOCAL_KEY_TO_USER)
	factory.CacheRds.SetEx(rdsKey, userId, EXPIRE_KEY_TO_USER)
	return login.UserId, nil
}

func (*defaultKey2UserId) DelUserIdByKey(key string) error {
	if key == "" {
		return nil
	}
	return factory.CacheRds.Del(PREFIX_KEY_TO_USER + key)
}

func DefaultKeyService() *defaultKey2UserId {
	return &defaultKey2UserId{}
}
