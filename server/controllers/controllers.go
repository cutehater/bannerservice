package controllers

import (
	"bannerservice/db"
	"bannerservice/schemas"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"net/http"
	"strconv"
)

func parseQueries(c *gin.Context) (tagId int, featureId int, useLastRevision bool, limit int, offset int, err error) {
	tagQuery := c.Query("tag_id")
	if len(tagQuery) > 0 {
		tagId, err = strconv.Atoi(tagQuery)
		if err != nil {
			err = errors.New("invalid tag_id: must be integer")
			return
		}
	}

	featureQuery := c.Query("feature_id")
	if len(featureQuery) > 0 {
		featureId, err = strconv.Atoi(featureQuery)
		if err != nil {
			err = errors.New("invalid feature_id: must be integer")
			return
		}
	}

	useLastRevisionQuery := c.Query("use_last_revision")
	if len(useLastRevisionQuery) > 0 {
		useLastRevision, err = strconv.ParseBool(useLastRevisionQuery)
		if err != nil {
			err = errors.New("invalid use_last_revision: must be boolean")
			return
		}
	}

	limitQuery := c.Query("limit")
	if len(limitQuery) > 0 {
		limit, err = strconv.Atoi(limitQuery)
		if err != nil {
			err = errors.New("invalid limit: must be integer")
			return
		}
	}

	offsetQuery := c.Query("offset")
	if len(offsetQuery) > 0 {
		offset, err = strconv.Atoi(offsetQuery)
		if err != nil {
			err = errors.New("invalid offset: must be integer")
			return
		}
	}

	return
}

func GetUserBanner(c *gin.Context) {
	tagId, featureId, useLastRevision, _, _, err := parseQueries(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("error parsing query params: %w", err).Error()})
		return
	}

	if tagId == 0 || featureId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_id and feature_id are required"})
		return
	}

	cacheEntryKey := fmt.Sprintf("%d,%d", tagId, featureId)
	var banner schemas.Banner
	if v, ok := db.BannerCache.Get(cacheEntryKey); !ok || useLastRevision {
		banner, ok = v.(schemas.Banner)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid banner cache entry"})
			return
		}
	} else {
		db.DB.Model(&schemas.Banner{}).Where("tag_ids @> ARRAY[?]::integer[] AND feature_id = ?", tagId, featureId).First(&banner)
	}

	db.BannerCache.Set(cacheEntryKey, banner, cache.DefaultExpiration)
	if banner.ID == 0 {
		c.Status(http.StatusNotFound)
	} else if !banner.IsActive {
		c.Status(http.StatusForbidden)
	} else {
		c.JSON(http.StatusOK, banner.Content)
	}
}

func GetBanners(c *gin.Context) {
	tagId, featureId, _, limit, offset, err := parseQueries(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("error parsing query params: %w", err).Error()})
		return
	}

	if tagId == 0 && featureId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag_id or feature_id is required"})
		return
	}

	var banners []schemas.Banner
	dbQuery := db.DB.Model(&schemas.Banner{})

	if featureId != 0 {
		dbQuery = dbQuery.Where("feature_id = ?", featureId)
	}
	if tagId != 0 {
		dbQuery = dbQuery.Where("tag_ids @> ARRAY[?]::integer[]", pq.Array([]int64{int64(tagId)}))
	}
	if limit != 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset != 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	if err = dbQuery.Find(&banners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("error getting banners from database: %w", err).Error()})
	} else {
		resp, err := json.Marshal(banners)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("error getting marshalling banners: %w", err).Error()})
		} else {
			c.JSON(http.StatusOK, resp)
		}
	}
}

func PostBanner(c *gin.Context) {
	var banner schemas.Banner
	if err := c.BindJSON(&banner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid banner body: %w", err).Error()})
		return
	}
	if banner.FeatureID == 0 || banner.TagIDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feature_id and tag_ids should be non-empty"})
	}

	res := db.DB.Model(&schemas.Banner{}).Create(&banner)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("error creating banner in database: %w", res.Error).Error()})
	} else {
		c.JSON(http.StatusCreated, gin.H{"banner_id": banner.ID})
	}
}

func findBannerbyId(c *gin.Context) *schemas.Banner {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, fmt.Errorf("invalid banner id: %w", err).Error())
		return nil
	}

	var banner schemas.Banner
	if err = db.DB.Model(&schemas.Banner{}).First(&banner, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return nil
	}

	return &banner
}

func UpdateBanner(c *gin.Context) {
	banner := findBannerbyId(c)
	if banner == nil {
		return
	}

	if err := c.BindJSON(banner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid banner body: %w", err).Error()})
		return
	}

	if err := db.DB.Model(&schemas.Banner{}).Save(banner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Errorf("error saving banner to database: %w", err).Error())
		return
	} else {
		c.Status(http.StatusOK)
	}
}

func DeleteBanner(c *gin.Context) {
	banner := findBannerbyId(c)
	if banner == nil {
		return
	}

	if err := db.DB.Model(&schemas.Banner{}).Delete(banner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("error deleting banner from database: %w", err).Error()})
		return
	} else {
		c.Status(http.StatusNoContent)
	}

}
