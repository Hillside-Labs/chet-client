package models

import (
	"errors"
	"log"

	"gorm.io/gorm"
)

func NewUser(email string, db *gorm.DB) (*User, error) {
	u := User{Email: email}
	user, err := FindUser(u, db)

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		db.Save(&u)
		log.Println("Created user: ", u)
	}

	return &u, nil
}

func FindUser(conditions any, db *gorm.DB) (*User, error) {
	var user User
	err := db.Where(conditions).First(&user).Error

	return &user, err
}
