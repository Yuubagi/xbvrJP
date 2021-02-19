package config

import (
	"encoding/json"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
)

type ObjectState struct {
	Web struct {
		TagSort   string `json:"tagSort"`
		SceneEdit bool   `json:"sceneEdit"`
	} `json:"web"`
	DLNA struct {
		Running  bool     `json:"running"`
		Images   []string `json:"images"`
		RecentIP []string `json:"recentIp"`
	} `json:"dlna"`
	CacheSize struct {
		Images      int64 `json:"images"`
		Previews    int64 `json:"previews"`
		SearchIndex int64 `json:"searchIndex"`
	} `json:"cacheSize"`
}

var State ObjectState

func LoadState() {
	db, _ := models.GetDB()
	defer db.Close()

	var obj models.KV
	err := db.Where(&models.KV{Key: "state"}).First(&obj).Error
	if err == nil {
		if err := json.Unmarshal([]byte(obj.Value), &State); err != nil {
			common.Log.Error("Failed to load state from database")
		}
	}
}

func SaveState() {
	data, err := json.Marshal(State)
	if err == nil {
		obj := models.KV{Key: "state", Value: string(data)}
		obj.Save()
		common.Log.Info("Saved state")
	}
}
