package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"server/db"
	"server/routes"
	"server/schemas"
	"testing"
)

var router *gin.Engine

func init() {
	db.ConnectToDb()
	db.InitCaches()

	router = gin.Default()
	routes.SetupRoutes(router)
}

type bannerRequest struct {
	TagIDs    []int64                `json:"tag_ids"`
	FeatureID int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

func getBannerJSON(t *testing.T, tagIDs []int64, featureID int, isActive bool, content string) []byte {
	t.Helper()

	banner := bannerRequest{
		TagIDs:    tagIDs,
		FeatureID: featureID,
		Content: map[string]interface{}{
			"content": content,
		},
		IsActive: isActive,
	}

	jsonData, err := json.Marshal(banner)
	require.NoError(t, err)

	return jsonData
}

func addBanner(t *testing.T, bannerJSON []byte) (ID int) {
	t.Helper()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/banner", bytes.NewBuffer(bannerJSON))
	require.NoError(t, err)
	req.Header.Set("token", "admin_token")

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response map[string]int
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	return response["banner_id"]
}

func TestGetUserBanner(t *testing.T) {
	activeFeature := int(rand.Int31())
	inactiveFeature := int(rand.Int31())
	activeBanner := getBannerJSON(t, []int64{1, 2, 3}, activeFeature, true, "active")
	inactiveBanner := getBannerJSON(t, []int64{1, 2, 3}, inactiveFeature, false, "inactive")
	addBanner(t, activeBanner)
	addBanner(t, inactiveBanner)

	var tests = []struct {
		name                  string
		token                 string
		tagId                 int
		featureId             int
		expectedStatus        int
		expectedBannerContent string
	}{
		{
			name:                  "OK user",
			token:                 "user_token",
			tagId:                 2,
			featureId:             activeFeature,
			expectedStatus:        http.StatusOK,
			expectedBannerContent: "active",
		},
		{
			name:                  "OK admin",
			token:                 "admin_token",
			tagId:                 2,
			featureId:             activeFeature,
			expectedStatus:        http.StatusOK,
			expectedBannerContent: "active",
		},
		{
			name:           "Inactive banner user",
			token:          "user_token",
			tagId:          1,
			featureId:      inactiveFeature,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid token",
			token:          "invalid",
			tagId:          2,
			featureId:      activeFeature,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Banner not exist",
			token:          "user_token",
			tagId:          2,
			featureId:      int(rand.Int31()),
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		path := fmt.Sprintf("/user_banner?tag_id=%v&feature_id=%v", test.tagId, test.featureId)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		req.Header.Set("token", test.token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, test.expectedStatus, w.Code)

		if test.expectedStatus == http.StatusOK {
			var actual map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &actual)
			require.NoError(t, err)
			require.Equal(t, test.expectedBannerContent, actual["content"].(string))
		}
	}
}

func TestGetUserBannerUseLastRevision(t *testing.T) {
	w := httptest.NewRecorder()
	feature := int(rand.Int31())
	oldBanner := getBannerJSON(t, []int64{1}, feature, true, "old")
	id := addBanner(t, oldBanner)

	path := fmt.Sprintf("/user_banner?tag_id=%v&feature_id=%v", 1, feature)
	reqGet, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	reqGet.Header.Set("token", "user_token")

	router.ServeHTTP(w, reqGet)
	require.Equal(t, http.StatusOK, w.Code)
	var actual map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err)
	require.Equal(t, "old", actual["content"])

	newBanner := getBannerJSON(t, []int64{1}, feature, true, "new")
	require.NoError(t, err)
	reqPatch, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/banner/%v", id), bytes.NewBuffer(newBanner))
	require.NoError(t, err)
	reqPatch.Header.Set("token", "admin_token")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, reqPatch)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, reqGet)
	require.Equal(t, http.StatusOK, w.Code)
	actual = make(map[string]interface{})
	err = json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err)
	require.Equal(t, "old", actual["content"])

	path = fmt.Sprintf("/user_banner?tag_id=%v&feature_id=%v&use_last_revision=true", 1, feature)
	reqGet, err = http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	reqGet.Header.Set("token", "user_token")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, reqGet)
	require.Equal(t, http.StatusOK, w.Code)
	actual = make(map[string]interface{})
	err = json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err)
	require.Equal(t, "new", actual["content"])
}

func TestGetBanner(t *testing.T) {
	uniqueFeature := int(rand.Int31())
	uniqueTag := int(rand.Int31())

	firstBanner := getBannerJSON(t, []int64{int64(uniqueTag)}, uniqueFeature, true, "first")
	secondBanner := getBannerJSON(t, []int64{int64(uniqueTag), 2, 3}, 1, true, "second")
	thirdBanner := getBannerJSON(t, []int64{1}, uniqueFeature, true, "third")
	idFirst := addBanner(t, firstBanner)
	idSecond := addBanner(t, secondBanner)
	idThird := addBanner(t, thirdBanner)

	var tests = []struct {
		name              string
		token             string
		tagID             int
		featureID         int
		expectedStatus    int
		expectedBannerIDs map[int]struct{}
	}{
		{
			name:              "OK by tag_id",
			token:             "admin_token",
			tagID:             uniqueTag,
			expectedStatus:    http.StatusOK,
			expectedBannerIDs: map[int]struct{}{idFirst: {}, idSecond: {}},
		},
		{
			name:              "OK by feature_id",
			token:             "admin_token",
			featureID:         uniqueFeature,
			expectedStatus:    http.StatusOK,
			expectedBannerIDs: map[int]struct{}{idFirst: {}, idThird: {}},
		},
		{
			name:              "OK by tag_id and feature_id",
			token:             "admin_token",
			featureID:         uniqueFeature,
			tagID:             uniqueTag,
			expectedStatus:    http.StatusOK,
			expectedBannerIDs: map[int]struct{}{idFirst: {}},
		},
		{
			name:           "Missing admin token",
			token:          "user_token",
			featureID:      1,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:              "OK no banners found",
			token:             "admin_token",
			featureID:         int(rand.Int31()),
			expectedStatus:    http.StatusOK,
			expectedBannerIDs: map[int]struct{}{},
		},
		{
			name:           "Nor tag_id nor feature_id",
			token:          "admin_token",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()

		path := fmt.Sprintf("/banner?tag_id=%v&feature_id=%v", test.tagID, test.featureID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		req.Header.Set("token", test.token)

		router.ServeHTTP(w, req)
		require.Equal(t, test.expectedStatus, w.Code)
		if test.expectedStatus == http.StatusOK {
			var banners []schemas.Banner
			err = json.NewDecoder(w.Body).Decode(&banners)
			require.NoError(t, err)
			actualBannerIDs := make(map[int]struct{})
			for _, b := range banners {
				actualBannerIDs[int(b.ID)] = struct{}{}
			}

			require.True(t, reflect.DeepEqual(test.expectedBannerIDs, actualBannerIDs))
		}
	}
}

func TestPostBanner(t *testing.T) {
	var tests = []struct {
		name           string
		token          string
		bannerJSON     []byte
		expectedStatus int
	}{
		{
			name:           "OK",
			token:          "admin_token",
			bannerJSON:     getBannerJSON(t, []int64{1, 2, 3}, 3, true, "post"),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Missing admin token",
			token:          "user_token",
			bannerJSON:     getBannerJSON(t, []int64{1, 2, 3}, 3, true, "post"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Without tag_ids",
			token:          "admin_token",
			bannerJSON:     getBannerJSON(t, nil, 3, true, "post"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Without feature_id",
			token:          "admin_token",
			bannerJSON:     getBannerJSON(t, []int64{1, 2, 3}, 0, true, "post"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid banner body",
			token:          "admin_token",
			bannerJSON:     []byte("random body"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		req, err := http.NewRequest(http.MethodPost, "/banner", bytes.NewBuffer(test.bannerJSON))
		require.NoError(t, err)
		req.Header.Set("token", test.token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, test.expectedStatus, w.Code)
	}
}

func TestPatchBanner(t *testing.T) {
	existingId := addBanner(t, getBannerJSON(t, []int64{1, 2, 3}, 4, true, "patchOld"))
	newBannerJSON := getBannerJSON(t, []int64{1, 2, 3}, 8, false, "patchNew")

	var tests = []struct {
		name           string
		id             int
		token          string
		bannerJSON     []byte
		expectedStatus int
	}{
		{
			name:           "OK",
			id:             existingId,
			token:          "admin_token",
			bannerJSON:     newBannerJSON,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing admin token",
			id:             existingId,
			token:          "user_token",
			bannerJSON:     newBannerJSON,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Banner not exists",
			id:             100,
			token:          "admin_token",
			bannerJSON:     newBannerJSON,
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("/banner/%v", test.id), bytes.NewBuffer(test.bannerJSON))
		require.NoError(t, err)
		req.Header.Set("token", test.token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, test.expectedStatus, w.Code)
	}
}

func TestDeleteBanner(t *testing.T) {
	existingId := addBanner(t, getBannerJSON(t, []int64{1, 2, 3}, 5, true, "delete"))

	var tests = []struct {
		name           string
		id             int
		token          string
		expectedStatus int
	}{
		{
			name:           "OK",
			token:          "admin_token",
			id:             existingId,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Missing admin token",
			token:          "user_token",
			id:             0,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Banner not exists",
			id:             100,
			token:          "admin_token",
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/banner/%v", test.id), nil)
		require.NoError(t, err)
		req.Header.Set("token", test.token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, test.expectedStatus, w.Code)
	}
}
