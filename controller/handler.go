package controller

import "gorm.io/gorm"

//this is dependency injection
type handler struct{
	DB *gorm.DB
}

func New(db *gorm.DB) handler{
	return handler{db}
}