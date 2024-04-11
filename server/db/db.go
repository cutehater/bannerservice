package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"server/schemas"
)

var DB *gorm.DB

func ConnectToDb() {
	dsn := os.Getenv("DB_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := DB.AutoMigrate(&schemas.User{}, &schemas.Banner{}); err != nil {
		log.Fatal(err)
	}

	user := schemas.User{Token: "user_token", IsAdmin: false}
	admin := schemas.User{Token: "admin_token", IsAdmin: true}
	res := DB.Model(&schemas.User{}).Create(&user)
	if res.Error != nil {
		log.Fatal(err)
	}
	res = DB.Model(&schemas.User{}).Create(&admin)
	if res.Error != nil {
		log.Fatal(err)
	}
}
