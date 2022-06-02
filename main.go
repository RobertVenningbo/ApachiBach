package main

import (
	"fmt"
	_ "log"
	_ "net/http"
	"os"
	_ "swag/backend"
	"swag/controller"
	"swag/db"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	DB := db.Init()
	h := controller.New(DB)
	/*
		Shared routes/end-points
	*/
	// router.LoadHTMLGlob("templates/*")
	v1 := router.Group("/v1/api")
	{
		// set&get for the log
		v1.POST("/logmsg", h.CreateMessage)
		v1.GET("/logmsg", h.GetMessages)
		v1.GET("/logmsg/:id", h.GetMessage)
	}

	router.GET("/log", controller.LogHandler)

	var ispctaken bool
	serverport := os.Args[2]
	if os.Args[1] == "submitter" {
		router.GET("/", controller.SubmissionHandler)
		router.Run(":" + serverport)
	} else if os.Args[1] == "reviewer" {
		router.GET("/", controller.SubmissionHandler) //fix, give a reviewer its own
		router.Run(":" + serverport)
	} else if os.Args[1] == "pc" {
		if !ispctaken {
			ispctaken = true //TODO: make this work
			router.GET("/", controller.PCHandler)
			router.Run(":" + serverport)
		} else {
			fmt.Println("PC is already running")
			os.Exit(1)
		}
	} else {
		fmt.Println("Wrong arguments given")
		os.Exit(1)
	}

}
