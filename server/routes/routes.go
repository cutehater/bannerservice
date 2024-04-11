package routes

import (
	"github.com/gin-gonic/gin"

	"server/controllers"
	"server/middlewares"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/user_banner", middlewares.IsAuthorized(false), controllers.GetUserBanner)
	r.GET("/banner", middlewares.IsAuthorized(true), controllers.GetBanners)
	r.POST("/banner", middlewares.IsAuthorized(true), controllers.PostBanner)
	r.PATCH("/banner/:id", middlewares.IsAuthorized(true), controllers.UpdateBanner)
	r.DELETE("/banner/:id", middlewares.IsAuthorized(true), controllers.DeleteBanner)
}
