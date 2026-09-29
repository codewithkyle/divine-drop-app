package controllers

import (
	"app/helpers"
	"app/models"
)

// Cards.front, Cards.back and Deck_Cards.print hold a look hash, and Cards.art
// holds a card id. None of them is an address, so a card cannot be rendered
// until it has been through one of these. They are the only place a hash
// becomes a URL, which is what makes the set of conversions auditable: every
// query that reaches a template is followed by a call to one of them.
//
// helpers.Card*URL returns an empty string for an empty hash, so a card the
// processor has not reached yet renders nothing rather than a broken address.

func resolveCards(cards []models.Card) {
	for i := range cards {
		cards[i].Front = helpers.CardFrontURL(cards[i].Front)
		cards[i].Back = helpers.CardBackURL(cards[i].Back)
		cards[i].Art = helpers.CardArtURL(cards[i].Art)
	}
}

func resolveCard(card *models.Card) {
	card.Front = helpers.CardFrontURL(card.Front)
	card.Back = helpers.CardBackURL(card.Back)
	card.Art = helpers.CardArtURL(card.Art)
}

// resolveDeckCards takes the faces already resolved in SQL, which folds a
// player's chosen look into front and back before they get here.
func resolveDeckCards(cards []models.DeckCard) {
	for i := range cards {
		cards[i].Front = helpers.CardFrontURL(cards[i].Front)
		cards[i].Back = helpers.CardBackURL(cards[i].Back)
		cards[i].Art = helpers.CardArtURL(cards[i].Art)
	}
}

func resolveDeckCard(card *models.DeckCard) {
	card.Front = helpers.CardFrontURL(card.Front)
	card.Back = helpers.CardBackURL(card.Back)
	card.Art = helpers.CardArtURL(card.Art)
}

func resolveDeckCardsMetadata(cards []models.DeckCardMetadata) {
	for i := range cards {
		cards[i].Front = helpers.CardFrontURL(cards[i].Front)
		cards[i].Back = helpers.CardBackURL(cards[i].Back)
	}
}

// resolveSleeves turns stored sleeve keys into URLs. Sleeves.image_url holds a
// key for the same reason Cards.front holds a hash: the origin belongs to
// CDN_URL, not to a row written years ago.
func resolveSleeves(sleeves []models.Sleeve) {
	for i := range sleeves {
		sleeves[i].Image = helpers.AssetURL(sleeves[i].Image)
	}
}
