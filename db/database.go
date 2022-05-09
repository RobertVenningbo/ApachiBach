package db

import (
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"fmt"
)

var Conn *gorm.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "q"
	dbname   = "apachi"
)

func ConnectDB(){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	Conn, err := gorm.Open(postgres.Open(psqlInfo), gorm.DB{})
	if err != nil {
		panic(err)
	}
	sqldb, err :=  Conn.DB()
	defer sqldb.Close()

	err = sqldb.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to database")
}