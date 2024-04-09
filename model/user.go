package model

import (
	"errors"
	"fmt"
	"konzek_assg/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null;" json:"password"`
	Tasks    []Task `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (user *User) SaveInTransaction(tx *gorm.DB) (*User, error) {
	if user.ID == 0 {
		err := tx.Create(user).Error
		if err != nil {
			return nil, errors.New("error in creating user")
		}
	} else {
		err := tx.Save(user).Error
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func (user *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

}

func FindUserByUsername(username string) (User, error) {
	var user User
	err := database.Database.Where("username=?", username).Find(&user).Error
	if err != nil {
		return User{}, fmt.Errorf("there is no such a user")
	}
	return user, nil
}

func FindUserById(id uint) (User, error) {
	var user User
	err := database.Database.Preload("Tasks").Where("ID=?", id).Find(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}
