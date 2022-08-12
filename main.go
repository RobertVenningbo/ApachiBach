package main

import (
	"fmt"
	"os"
	"swag/backend"
	"swag/controller"
	"swag/database"
	"swag/model"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	db := database.Init()
	if (db.Migrator().HasTable(&model.Log{}) && os.Args[1] == "pc") {
		db.Migrator().DropTable(&model.Log{})
	}
	if (db.Migrator().HasTable(&model.User{}) && os.Args[1] == "pc") {
		db.Migrator().DropTable(&model.User{})
	}
	db.AutoMigrate(&model.Log{})
	db.AutoMigrate(&model.User{})
	backend.InitGobs()

	router.GET("/testing", controller.TestPlatform) 
	router.GET("/log", controller.LogHandler)

	serverport := os.Args[2]
	if os.Args[1] == "submitter" {
		router.GET("/", controller.SubmissionHandler)
		router.GET("/wait", controller.WaitHandler)
		router.POST("/upload", controller.UploadFile)
		router.GET("/getgrade", controller.GetGradesAndReviewsHandler)
		router.GET("/claimgrade", controller.ClaimPaperHandler)
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
		router.POST("/compilegrades", controller.CompileGradesHandler)
		router.POST("/confirmowner", controller.ConfirmOwnershipHandler)
		router.GET("/confirmowner", controller.GetConfirmOwnershipHandler)
		router.Run(":" + serverport)
	} else {
		fmt.Println("Wrong arguments given")
		os.Exit(1)
	}

}
