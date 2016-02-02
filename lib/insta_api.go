package lib

import (
	"net/url"
	"time"

	"github.com/gedex/go-instagram/instagram"
	"github.com/jinzhu/gorm"
)

const (
	waitBetweenChecks = 10 * time.Hour
	backfillWait      = 5 * time.Second
)

// DI
var (
	getLikedMedia = (*instagram.UsersService).LikedMedia
	timeSleep     = time.Sleep

	instaAPISaveLocation = (*InstaAPI).saveLocation
)

// InstaAPI encapsulate functionality for all instagram functionality
type InstaAPI struct {
	client *instagram.Client
	db     *gorm.DB
}

// NewInstaAPI Provider for InstaAPI
func NewInstaAPI(cfg *Cfg, db *gorm.DB) *InstaAPI {
	i := new(InstaAPI)

	i.client = instagram.NewClient(nil)
	i.client.ClientID = cfg.Instagram.ClientID
	i.client.ClientSecret = cfg.Instagram.Secret
	i.client.AccessToken = cfg.Instagram.Token
	i.db = db

	return i
}

// SaveLikes Inserts instagram likes into the DB
func (i *InstaAPI) SaveLikes() {
	for {
		media, _, _ := getLikedMedia(i.client.Users, nil)

		for _, m := range media {
			i.saveMedia(&m)
		}

		timeSleep(waitBetweenChecks)
	}
}

// Backfill Puts in historical likes
func (i *InstaAPI) Backfill(maxLikeID string) {
	logger.Infof("Running backfill for %s", maxLikeID)

	media, after, _ := getLikedMedia(i.client.Users, &instagram.Parameters{MaxID: maxLikeID})
	afterURL, _ := url.Parse(after.NextURL)
	maxLikeID = afterURL.Query().Get("max_like_id")

	for _, m := range media {
		i.saveMedia(&m)
	}

	if maxLikeID != "" {
		timeSleep(backfillWait)
		i.Backfill(maxLikeID)
	}
}

func (i *InstaAPI) saveMedia(m *instagram.Media) {
	if !i.isLocationOk(m) {
		return
	}

	loc := instaAPISaveLocation(i, m)
	var e Entry
	i.db.FirstOrCreate(&e, Entry{
		Type:       "instagram",
		VendorID:   m.ID,
		Thumbnail:  m.Images.Thumbnail.URL,
		URL:        m.Link,
		Caption:    m.Caption.Text,
		Timestamp:  m.CreatedTime,
		LocationID: loc.ID,
	})
}

func (i *InstaAPI) saveLocation(m *instagram.Media) *Location {
	var l Location
	i.db.FirstOrCreate(&l, Location{
		Name: m.Location.Name,
		Lat:  m.Location.Latitude,
		Long: m.Location.Longitude,
	})

	return &l
}

func (i *InstaAPI) isLocationOk(media *instagram.Media) bool {
	return media.Location != nil
}
