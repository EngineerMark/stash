package character

import (
	"context"

	"github.com/stashapp/stash/pkg/models"
)

func ByName(ctx context.Context, qb models.CharacterQueryer, name string) (*models.Character, error) {
	f := &models.CharacterFilterType{
		Name: &models.StringCriterionInput{
			Value:    name,
			Modifier: models.CriterionModifierEquals,
		},
	}

	pp := 1
	ret, count, err := qb.Query(ctx, f, &models.FindFilterType{
		PerPage: &pp,
	})

	if err != nil {
		return nil, err
	}

	if count > 0 {
		return ret[0], nil
	}

	return nil, nil
}

func ByAlias(ctx context.Context, qb models.CharacterQueryer, alias string) (*models.Character, error) {
	f := &models.CharacterFilterType{
		Aliases: &models.StringCriterionInput{
			Value:    alias,
			Modifier: models.CriterionModifierEquals,
		},
	}

	pp := 1
	ret, count, err := qb.Query(ctx, f, &models.FindFilterType{
		PerPage: &pp,
	})

	if err != nil {
		return nil, err
	}

	if count > 0 {
		return ret[0], nil
	}

	return nil, nil
}
