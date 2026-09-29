package helpers

import (
	"strings"
	"sync"
)

// defaultCDNURL is where assets are served from when CDN_URL is unset. It names
// the Cloudflare origin in use now rather than the DigitalOcean host it used to
// fall back to: that bucket is being decommissioned, so a default aimed there
// would start returning nothing. Moving hosts stays a change to the
// environment, not a release.
const defaultCDNURL = "https://cdn.divinedrop.app"

var (
	cdnURL     string
	cdnURLOnce sync.Once
)

// CDNBaseURL is the origin every asset URL below is built from, without a
// trailing slash. Set CDN_URL to move all of them at once.
func CDNBaseURL() string {
	cdnURLOnce.Do(func() {
		cdnURL = strings.TrimRight(envOr("CDN_URL", defaultCDNURL), "/")
	})
	return cdnURL
}

// CardFrontURL is the front face of a printing, addressed by its look.
//
// The hash is the whole address: it identifies an artwork and treatment rather
// than a card plus a date, so every printing that renders identically shares
// one object. An empty hash yields an empty string rather than a URL with a
// hole in it, because a card with no processed image has nothing to point at.
func CardFrontURL(hash string) string {
	if hash == "" {
		return ""
	}
	return CDNBaseURL() + "/cards/" + strings.ToLower(hash) + "-front.webp"
}

// CardBackURL is the reverse face, which only double faced cards have.
func CardBackURL(hash string) string {
	if hash == "" {
		return ""
	}
	return CDNBaseURL() + "/cards/" + strings.ToLower(hash) + "-back.webp"
}

// CardArtURL is the cropped illustration used for deck banners, addressed by the
// art hash in Cards.art.
//
// Read that column rather than reusing front: they hold the same hash for a
// card's default printing, but a card can have a front and no crop, and nothing
// in the bucket is addressed by a database id, which resolves differently per
// environment.
func CardArtURL(hash string) string {
	if hash == "" {
		return ""
	}
	return CDNBaseURL() + "/cards/" + strings.ToLower(hash) + "-art.webp"
}

// CardSleeveURL is the plain card back shown for a deck with no sleeve set. It
// is a single static object, not a per-card image, so it keeps its own name.
func CardSleeveURL() string {
	return CDNBaseURL() + "/back.png"
}

// UserUploadKey names a sleeve a user uploaded, in the object store and in
// Sleeves.image_url alike. Storing the key rather than a URL is what keeps
// CDN_URL the single authority on where assets are served from, and it means
// the delete path addresses exactly the object the upload wrote instead of
// rebuilding the name and hoping the two agree.
func UserUploadKey(userId string, fileId string) string {
	return "users/" + userId + "/" + fileId
}

// AssetURL serves any stored key. Keys are relative by construction, so a value
// that still looks absolute is one the migration has not reached, and is passed
// through rather than mangled into an origin followed by another origin.
func AssetURL(key string) string {
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	return CDNBaseURL() + "/" + strings.TrimLeft(key, "/")
}
