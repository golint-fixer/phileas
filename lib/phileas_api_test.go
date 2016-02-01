package lib

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/json"

	testdb "github.com/erikstmartin/go-testdb"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var (
	service *gin.Engine
	minHTML = minify.New()
	minJSON = minify.New()
)

func init() {
	minHTML.AddFunc("text/html", html.Minify)
	minJSON.AddFunc("text/json", json.Minify)
}

func init() {
	db, _ := gorm.Open("testdb", "")

	cfg := &Cfg{}
	cfg.Common.GoogleMapsKey = "test-key"

	ginLoadHTMLGlob = func(*gin.Engine, string) {}
	service = NewService(cfg, &db, &InstaAPI{})
}

func peformRequest(method string, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	service.ServeHTTP(w, req)

	return w
}

func TestPing(t *testing.T) {
	w := peformRequest("GET", "/ping")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestMapper(t *testing.T) {
	tmpl, _ := template.New("mapper.tmpl").Parse(`{{ .title }} | {{ .key }}`)
	service.SetHTMLTemplate(tmpl)

	w := peformRequest("GET", "/top")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Top destinations | test-key", w.Body.String())
}

func TestTopJSON(t *testing.T) {
	sql := `SELECT * from "locations"`
	result := `
	1, location-1, 1.0, 1.0, test address, UK, London, NA, NA
	2, location-2, 1.0, 1.0, test address, IN, Mumbai, NA, NA
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(locationCols, result))
	expected := `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[1,1]},"properties":{"id":1,"name":"location-1"}},{"type":"Feature","geometry":{"type":"Point","coordinates":[1,1]},"properties":{"id":2,"name":"location-2"}}]}`

	w := peformRequest("GET", "/top.json")

	assert.Equal(t, minifyJSON(expected), minifyJSON(w.Body.String()))
}

func TestLocation(t *testing.T) {
	sql := `SELECT  * FROM "entries"  WHERE ("location_id" = ?)`
	result := `
	1, instagram, 12345, http://thumbnail/url, http://full/url, test caption, 12345678, 1
	`
	expected := `
	<div class="row">
		<div class="left"><a href="http://full/url" target="_blank"><img src="http://thumbnail/url" /></a></div>
        <div class="right">test caption</div>
    </div>
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString(entryCols, result))
	w := peformRequest("GET", "/loc/1")

	assert.Equal(t, minifyHTML(expected), minifyHTML(w.Body.String()))
}

func minifyHTML(raw string) string {
	out, _ := minHTML.String("text/html", raw)

	return out
}

func minifyJSON(raw string) string {
	out, _ := minJSON.String("text/json", raw)

	return out
}
