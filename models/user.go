package models

import (
	"github.com/jinzhu/gorm"
	"rela_recommend/cache"
)

type User struct {
	Id int64 `gorm:"column:id" json:"id"`
	UserName string `gorm:"userName" json:"userName"`
}

type IUserModule interface {
	Query(int64, *User) error
}

func (this *User) TableName() string {
	return "app_user"
}

type UserModule struct {
	db *gorm.DB
	cach cache.Cache
}

func NewUserModule(db *gorm.DB, cach cache.Cache) IUserModule {
	return &UserModule{db: db, cach: cach}
}

func (this *UserModule) Query(id int64, user *User) error {
	return this.db.Where("id = ?", id).Find(user).Error
}