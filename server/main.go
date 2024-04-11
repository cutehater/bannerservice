package main

import (
	"bannerservice/server/controllers"
	db2 "bannerservice/server/db"
	"bannerservice/server/middlewares"
	"github.com/gin-gonic/gin"
)

func main() {
	db2.ConnectToDb()
	db2.InitCaches()

	r := gin.Default()

	r.GET("/user_banner", middlewares.IsAuthorized(false), controllers.GetUserBanner)
	r.GET("/banner", middlewares.IsAuthorized(true), controllers.GetBanners)
	r.POST("/banner", middlewares.IsAuthorized(true), controllers.PostBanner)
	r.PATCH("/banner/:id", middlewares.IsAuthorized(true), controllers.UpdateBanner)
	r.DELETE("/banner/:id", middlewares.IsAuthorized(true), controllers.DeleteBanner)

	r.Run()
}
