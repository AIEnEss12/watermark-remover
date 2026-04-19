package api

import (
	"crypto/sha256"
	"encoding/hex"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CachedImage struct {
	Data        []byte
	ContentType string
}

var imageCache *lru.Cache[string, CachedImage]

func init() {
	var err error
	// Cache up to 200 images
	imageCache, err = lru.New[string, CachedImage](200)
	if err != nil {
		panic(err)
	}
}

// GetCachedImage retrieves an image from the cache by URL.
func GetCachedImage(url string) (CachedImage, bool) {
	key := hashURL(url)
	return imageCache.Get(key)
}

// AddCachedImage adds an image to the cache.
func AddCachedImage(url string, data []byte, contentType string) {
	key := hashURL(url)
	imageCache.Add(key, CachedImage{
		Data:        data,
		ContentType: contentType,
	})
}

func hashURL(url string) string {
	h := sha256.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}
