package urlbuilders

import (
	"github.com/stashapp/stash/pkg/models"
	"strconv"
)

type CharacterURLBuilder struct {
	BaseURL   string
	CharacterID     string
	UpdatedAt string
}

func NewCharacterURLBuilder(baseURL string, character *models.Character) CharacterURLBuilder {
	return CharacterURLBuilder{
		BaseURL:   baseURL,
		CharacterID:     strconv.Itoa(character.ID),
		UpdatedAt: strconv.FormatInt(character.UpdatedAt.Unix(), 10),
	}
}

func (b CharacterURLBuilder) GetCharacterImageURL(hasImage bool) string {
	url := b.BaseURL + "/character/" + b.CharacterID + "/image?t=" + b.UpdatedAt
	if !hasImage {
		url += "&default=true"
	}
	return url
}
