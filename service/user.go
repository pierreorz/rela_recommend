package service

import (
	"rela_recommend/models"
	"rela_recommend/log"
	"rela_recommend/factory"
)

type UserCardRes struct {
	User models.User `json:"user"`
}

func GetUserCard(userId int64) (UserCardRes, error) {
	UserModule := models.NewUserModule(factory.DbR, factory.CacheRds)
	var res UserCardRes
	var user models.User
	if err := UserModule.Query(userId, &user); err != nil {
		log.Error(err.Error())
		return res, err
	}
	res.User = user
	return res, nil
}