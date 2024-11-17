package api

import (
	"context"
	"strconv"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/sliceutil/stringslice"
)

func (r *queryResolver) FindCharacter(ctx context.Context, id string) (ret *models.Character, err error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.Character.Find(ctx, idInt)
		return err
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *queryResolver) FindCharacters(ctx context.Context, characterFilter *models.CharacterFilterType, filter *models.FindFilterType, ids []string) (ret *FindCharactersResultType, err error) {
	idInts, err := stringslice.StringSliceToIntSlice(ids)
	if err != nil {
		return nil, err
	}

	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		var characters []*models.Character
		var err error
		var total int

		if len(idInts) > 0 {
			characters, err = r.repository.Character.FindMany(ctx, idInts)
			total = len(characters)
		} else {
			characters, total, err = r.repository.Character.Query(ctx, characterFilter, filter)
		}

		if err != nil {
			return err
		}

		ret = &FindCharactersResultType{
			Count: total,
			Characters:  characters,
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *queryResolver) AllCharacters(ctx context.Context) (ret []*models.Character, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.Character.All(ctx)
		return err
	}); err != nil {
		return nil, err
	}

	return ret, nil
}
