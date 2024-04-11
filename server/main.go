package main

import (
	"github.com/gin-gonic/gin"

	"server/db"
	"server/routes"
)

func main() {
	db.ConnectToDb()
	db.InitCaches()

	r := gin.Default()
	routes.SetupRoutes(r)

	r.Run()
}
