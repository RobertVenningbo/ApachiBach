package controller

import "gorm.io/gorm"

//this is dependency injection
type handler struct {
	DB *gorm.DB // this is deprecated as per 06/06/2022, waiting with deleting till 100% sure.
}

// this is deprecated as per 06/06/2022, waiting with deleting till 100% sure.
func New(db *gorm.DB) handler {
	return handler{db}
}
