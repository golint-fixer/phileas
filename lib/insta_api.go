package lib

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gedex/go-instagram/instagram"
	"github.com/jinzhu/gorm"
)

const (
	waitBetweenChecks = 10 * time.Hour
	backfillWait      = 5 * time.Second
)

type InstaAPI struct {
	client *instagram.Client
	db     *gorm.DB
}

// NewInstaAPI Provider for InstaAPI
func NewInstaAPI(cfg *Cfg) *InstaAPI {
	i := new(InstaAPI)

	i.client = instagram.NewClient(nil)
	i.client.ClientID = cfg.Instagram.ClientID
	i.client.ClientSecret = cfg.Instagram.Secret
	i.client.AccessToken = cfg.Instagram.Token
	i.db = GetDB(cfg)

	return i
}

// SaveLikes Inserts instagram likes into the DB
func (i *InstaAPI) SaveLikes() {
	for {
		media, _, _ := i.client.Users.LikedMedia(nil)

		for _, m := range media {
			i.saveMedia(m)
		}

		time.Sleep(waitBetweenChecks)
	}
}

// Backfill Puts in historical likes
func (i *InstaAPI) Backfill(maxLikeID string) {
	media, after, _ := i.client.Users.LikedMedia(&instagram.Parameters{MaxID: maxLikeID})
	afterURL, _ := url.Parse(after.NextURL)
	maxLikeID = afterURL.Query().Get("max_like_id")

	logger.Info(fmt.Sprintf("Media found: %d; max like: %s", len(media), maxLikeID))

	time.Sleep(backfillWait)

	if maxLikeID != "" {
		i.Backfill(maxLikeID)
	}
}

func (i *InstaAPI) saveMedia(m *instagram.Media) {
	if i.isLocationOk(m) {
		var e Entry
		i.db.FirstOrCreate(&e, Entry{
			Type:      "instagram",
			VendorID:  m.ID,
			Timestamp: m.CreatedTime,
		})
	}
}

func (i *InstaAPI) isLocationOk(media *instagram.Media) bool {
	return media.Location != nil
}
