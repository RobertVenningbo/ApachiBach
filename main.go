package main

import (
	"fmt"
	_ "log"
	_ "net/http"
	"os"
	"swag/backend"
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
	backend.InitGobs()
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
	router.GET("/testing", controller.TestPlatform) //TODO OBS.
	router.GET("/log", controller.LogHandler)

	serverport := os.Args[2]
	if os.Args[1] == "submitter" {
		router.GET("/", controller.SubmissionHandler)
		router.GET("/wait", controller.WaitHandler)
		router.GET("/papergraded", controller.GradedPaperHandler)
		router.POST("/upload", controller.UploadFile)
		router.GET("/getgrade", controller.GetGradesAndReviewsHandler)
		router.Run(":" + serverport)
	} else if os.Args[1] == "reviewer" {
		router.GET("/", controller.PrepStageHandler)
		router.GET("/paperbid", controller.PaperBidHandler)
		router.POST("/sendbid", controller.MakeBidHandler)
		router.GET("/downloadpaper", controller.WriteToFileHandler)
		router.GET("/makereview", controller.MakeReviewHandler)
		router.POST("/finishedreview", controller.FinishedReviewHandler)
		router.GET("/finishedreview", controller.GetFinishedReviewHandler)
		router.GET("/discussing", controller.DiscussingHandler)
		router.POST("/discussing", controller.PostMessageDiscussingHandler)
		router.POST("/submitgrade", controller.PostGradeDiscussingHandler)
		router.GET("/submitgrade", controller.GetGradeDiscussingHandler)
		router.GET("/signgradecommit", controller.GetAgreedGradeHandler)
		router.POST("/confirmgrade", controller.SignGradeHandler)
		router.Run(":" + serverport)
	} else if os.Args[1] == "pc" {
		router.GET("/", controller.PCHomeHandler)
		router.GET("/decision", controller.DecisionHandler)
		router.GET("/sharereviews", controller.ShareReviewsHandler)
		router.GET("/bidwait", controller.BidWaitHandler)
		router.GET("/getallbids", controller.GetAllBidsHandler)
		router.GET("/collectreviews", controller.ShareReviewsButtonHandler)
		router.GET("/checkreviews", controller.CheckReviewsHandler)
		router.POST("/decision", controller.AcceptPaperHandler)
		router.POST("/rejectpaper", controller.RejectPaperHandler)
		router.Run(":" + serverport)
	} else {
		fmt.Println("Wrong arguments given")
		os.Exit(1)
	}

}
