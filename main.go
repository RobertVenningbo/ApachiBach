package main

import (
	"fmt"
	_ "log"
	_ "net/http"
	"os"
	_ "swag/backend"
	"swag/controller"
	"swag/database"
	"swag/model"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	db := database.Init()
	h := controller.New(db)
	db.AutoMigrate(&model.Log{})
	db.AutoMigrate(&model.User{})
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
		router.GET("/wait", controller.WaitHandler)
		router.GET("/papergraded", controller.GradedPaperHandler)
		router.POST("/upload", controller.UploadFile)
		router.Run(":" + serverport)

	} else if os.Args[1] == "reviewer" {
		router.GET("/", controller.PrepStageHandler)
		router.Run(":" + serverport)
	} else if os.Args[1] == "pc" {
		if !ispctaken {
			ispctaken = true //TODO: make this work
			router.GET("/", controller.PCHomeHandler)
			router.GET("/decision", controller.DecisionHandler)
			router.GET("/sharereviews", controller.ShareReviewsHandler)
			router.GET("/distributepapers", controller.DistributePapersHandler)
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
