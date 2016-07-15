package lib

import (
	"github.com/vincer/libhdplatinum"
	"time"
	"errors"
)

var shadeDataCache ShadeDataCache

func GetShadeViews() []ShadeView {
	output := []ShadeView{}
	for _, s := range getShadeData() {
		output = append(output, ShadeViewFromShade(s))
	}

	return output
}

func FindShade(id string) (libhdplatinum.Shade, error) {
	shades := getShadeData()
	for _, s := range shades {
		if s.Id() == id {
			return s, nil
		}
	}
	return libhdplatinum.Shade{}, errors.New("Not found")
}

func getShadeData() ([]libhdplatinum.Shade) {
	if (time.Since(shadeDataCache.CacheTime).Seconds() > 10) {
		Log.Info("Shade data cache is too old. Refreshing.")
		RefreshShadeCache()
	}
	return shadeDataCache.ShadeData
}

func RefreshShadeCache() {
	Log.Debug("Refreshed shade data")
	shadeDataCache = ShadeDataCache{ShadeData: libhdplatinum.GetShades(Config.Ip, Config.Port), CacheTime: time.Now()}
}
