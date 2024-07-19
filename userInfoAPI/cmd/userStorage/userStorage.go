package userstorage

import (
	"fmt"
	"sync"

	model "main.go/user_model"
)

var UserStorage map[int]model.UserInfo
var UserStorageMutex *sync.RWMutex

func InitUserStorage() map[int]model.UserInfo {
	UserStorage = make(map[int]model.UserInfo)
	UserStorageMutex = &sync.RWMutex{} // Инициализация мьютекса
	return UserStorage
}
func UserData(UserStorage map[int]model.UserInfo) {
	UserStorage[1234567890] = model.UserInfo{
		Name:       "Vadim",
		Surname:    "Vadimov",
		Patronymic: "Vadimovich",
		Address:    "г. Москва, ул. Ленина, д. 5, кв. 1",
	}
	UserStorage[1111111111] = model.UserInfo{
		Name:       "Sergey",
		Surname:    "Sergeev",
		Patronymic: "Sergeevich",
		Address:    "г. Москва, ул. Ленина, д. 5, кв. 1",
	}
	UserStorage[1234565432] = model.UserInfo{
		Name:       "Ivan",
		Surname:    "Ivanovov",
		Patronymic: "Ivanovich",
		Address:    "г. Москва, ул. Ленина, д. 5, кв. 1",
	}
}

// GetAccount возвращает аккаунт из кэша по ID.
func UserInfo(passportSerie, passportNumber int) (model.UserInfo, error) {
	UserStorageMutex.RLock()
	defer UserStorageMutex.RUnlock()
	// Количество цифр в номере паспорта
	// Объединяем passportSerie и passportNumber в одно число
	combinedPassport := passportNumber*1000000 + passportSerie
	userInfo, exists := UserStorage[combinedPassport]
	if !exists {
		err := fmt.Errorf("нет пользователя с указанными данными")
		return userInfo, err
	}
	return userInfo, nil
}
