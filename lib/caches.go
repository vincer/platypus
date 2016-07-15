package lib

import (
	"github.com/vincer/libhdplatinum"
	"time"
)

type ShadeDataCache struct {
	ShadeData ([]libhdplatinum.Shade)
	CacheTime time.Time
}
