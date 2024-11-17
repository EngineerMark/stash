package character

import (
	"context"
	"fmt"

	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/json"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/utils"
)

type FinderAliasImageGetter interface {
	GetAliases(ctx context.Context, studioID int) ([]string, error)
	GetImage(ctx context.Context, characterID int) ([]byte, error)
	FindByChildCharacterID(ctx context.Context, childID int) ([]*models.Character, error)
}

// ToJSON converts a Character object into its JSON equivalent.
func ToJSON(ctx context.Context, reader FinderAliasImageGetter, character *models.Character) (*jsonschema.Character, error) {
	newCharacterJSON := jsonschema.Character{
		Name:          character.Name,
		Description:   character.Description,
		Favorite:      character.Favorite,
		CreatedAt:     json.JSONTime{Time: character.CreatedAt},
		UpdatedAt:     json.JSONTime{Time: character.UpdatedAt},
	}

	aliases, err := reader.GetAliases(ctx, character.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting character aliases: %v", err)
	}

	newCharacterJSON.Aliases = aliases

	image, err := reader.GetImage(ctx, character.ID)
	if err != nil {
		logger.Errorf("Error getting character image: %v", err)
	}

	if len(image) > 0 {
		newCharacterJSON.Image = utils.GetBase64StringFromData(image)
	}

	parents, err := reader.FindByChildCharacterID(ctx, character.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting parents: %v", err)
	}

	newCharacterJSON.Parents = GetNames(parents)

	return &newCharacterJSON, nil
}

func GetIDs(characters []*models.Character) []int {
	var results []int
	for _, character := range characters {
		results = append(results, character.ID)
	}

	return results
}

func GetNames(characters []*models.Character) []string {
	var results []string
	for _, character := range characters {
		results = append(results, character.Name)
	}

	return results
}
