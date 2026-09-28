package helpers

import (
	"strings"
	"sync"
)

// defaultCDNURL is where assets have always been served from. It stays as the
// fallback so a build that ships before CDN_URL is set anywhere still serves
// images: moving hosts is then a change to the environment, not a release, and
// reverting it is the same.
const defaultCDNURL = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com"

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

// CardFrontURL is the face of a card as a given printing shows it.
//
// print is whatever currently identifies a printing, which is its release date.
// It is a string rather than an int because that is all these names treat it
// as, and because it is what changes when printings are addressed by the
// artwork they carry instead.
func CardFrontURL(cardId string, print string) string {
	return CDNBaseURL() + "/cards/" + strings.ToUpper(cardId) + "-" + print + "-front.png"
}

// CardBackURL is the reverse face, which only double faced cards have.
func CardBackURL(cardId string, print string) string {
	return CDNBaseURL() + "/cards/" + strings.ToUpper(cardId) + "-" + print + "-back.png"
}

// CardArtURL is the cropped illustration used for deck banners. Its name
// carries no printing, so every printing of a card shares one.
func CardArtURL(cardId string) string {
	return CDNBaseURL() + "/cards/" + strings.ToUpper(cardId) + "-art.png"
}

// CardSleeveURL is the plain card back shown for a deck with no sleeve set.
func CardSleeveURL() string {
	return CDNBaseURL() + "/back.png"
}

// UserUploadURL is where a sleeve a user uploaded is served from.
func UserUploadURL(userId string, fileId string) string {
	return CDNBaseURL() + "/users/" + userId + "/" + fileId
}
