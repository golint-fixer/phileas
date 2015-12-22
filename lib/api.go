package lib

import (
	"encoding/json"
	"net/http"

	"github.com/kpawlik/geojson"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// PhileasAPI Provides the data for phileas's API
type PhileasAPI struct {
	googleKey string
	db        *gorm.DB
}

// NewPhileasAPI Go-style constructor to provide an instance of Phileas's API
func NewPhileasAPI(cfg *Cfg, db *gorm.DB) *PhileasAPI {
	api := &PhileasAPI{}
	api.googleKey = cfg.Common.GoogleMapsKey
	api.db = db

	return api
}

func (pe *PhileasAPI) ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (pe *PhileasAPI) mapper(c *gin.Context) {
	c.HTML(http.StatusOK, "mapper.tmpl", gin.H{
		"title": "Top destinations",
		"key":   pe.googleKey,
	})
}

func (pe *PhileasAPI) topJSON(c *gin.Context) {
	var locs []Location
	pe.db.Find(&locs)
	col := makeGeoJSON(locs)

	if body, err := json.Marshal(col); err == nil {
		c.ContentType()
		c.String(http.StatusOK, string(body))
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func makeGeoJSON(locs []Location) *geojson.FeatureCollection {
	var all []*geojson.Feature

	for _, loc := range locs {
		p := geojson.NewPoint(geojson.Coordinate{geojson.CoordType(loc.Lat), geojson.CoordType(loc.Long)})
		f := geojson.NewFeature(p, nil, nil)
		all = append(all, f)
	}

	col := geojson.NewFeatureCollection(all)
	return col
}
