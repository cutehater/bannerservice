package db

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var BannerCache *cache.Cache
var UserCache *cache.Cache

func InitCaches() {
	BannerCache = cache.New(5*time.Minute, 10*time.Minute)
	UserCache = cache.New(1*time.Hour, 24*time.Hour)
}
