package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/plugin/hook"
	"github.com/stashapp/stash/pkg/sliceutil/stringslice"
	"github.com/stashapp/stash/pkg/character"
	"github.com/stashapp/stash/pkg/utils"
)

func (r *mutationResolver) getCharacter(ctx context.Context, id int) (ret *models.Character, err error) {
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.Character.Find(ctx, id)
		return err
	}); err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *mutationResolver) CharacterCreate(ctx context.Context, input CharacterCreateInput) (*models.Character, error) {
	translator := changesetTranslator{
		inputMap: getUpdateInputMap(ctx),
	}

	// Populate a new character from the input
	newCharacter := models.NewCharacter()

	newCharacter.Name = input.Name
	newCharacter.Aliases = models.NewRelatedStrings(input.Aliases)
	newCharacter.Favorite = translator.bool(input.Favorite)
	newCharacter.Description = translator.string(input.Description)

	var err error

	// Process the base 64 encoded image string
	var imageData []byte
	if input.Image != nil {
		imageData, err = utils.ProcessImageInput(ctx, *input.Image)
		if err != nil {
			return nil, fmt.Errorf("processing image: %w", err)
		}
	}

	// Start the transaction and save the character
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		qb := r.repository.Character

		if err := character.ValidateCreate(ctx, newCharacter, qb); err != nil {
			return err
		}

		err = qb.Create(ctx, &newCharacter)
		if err != nil {
			return err
		}

		// update image table
		if len(imageData) > 0 {
			if err := qb.UpdateImage(ctx, newCharacter.ID, imageData); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	r.hookExecutor.ExecutePostHooks(ctx, newCharacter.ID, hook.CharacterCreatePost, input, nil)
	return r.getCharacter(ctx, newCharacter.ID)
}

func (r *mutationResolver) CharacterUpdate(ctx context.Context, input CharacterUpdateInput) (*models.Character, error) {
	characterID, err := strconv.Atoi(input.ID)
	if err != nil {
		return nil, fmt.Errorf("converting id: %w", err)
	}

	translator := changesetTranslator{
		inputMap: getUpdateInputMap(ctx),
	}

	// Populate character from the input
	updatedCharacter := models.NewCharacterPartial()

	updatedCharacter.Name = translator.optionalString(input.Name, "name")
	updatedCharacter.Favorite = translator.optionalBool(input.Favorite, "favorite")
	updatedCharacter.Description = translator.optionalString(input.Description, "description")

	updatedCharacter.Aliases = translator.updateStrings(input.Aliases, "aliases")

	var imageData []byte
	imageIncluded := translator.hasField("image")
	if input.Image != nil {
		imageData, err = utils.ProcessImageInput(ctx, *input.Image)
		if err != nil {
			return nil, fmt.Errorf("processing image: %w", err)
		}
	}

	// Start the transaction and save the character
	var t *models.Character
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		qb := r.repository.Character

		if err := character.ValidateUpdate(ctx, characterID, updatedCharacter, qb); err != nil {
			return err
		}

		t, err = qb.UpdatePartial(ctx, characterID, updatedCharacter)
		if err != nil {
			return err
		}

		// update image table
		if imageIncluded {
			if err := qb.UpdateImage(ctx, characterID, imageData); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	r.hookExecutor.ExecutePostHooks(ctx, t.ID, hook.CharacterUpdatePost, input, translator.getFields())
	return r.getCharacter(ctx, t.ID)
}

func (r *mutationResolver) BulkCharacterUpdate(ctx context.Context, input BulkCharacterUpdateInput) ([]*models.Character, error) {
	characterIDs, err := stringslice.StringSliceToIntSlice(input.Ids)
	if err != nil {
		return nil, fmt.Errorf("converting ids: %w", err)
	}

	translator := changesetTranslator{
		inputMap: getUpdateInputMap(ctx),
	}

	// Populate scene from the input
	updatedCharacter := models.NewCharacterPartial()

	updatedCharacter.Description = translator.optionalString(input.Description, "description")
	updatedCharacter.Favorite = translator.optionalBool(input.Favorite, "favorite")

	updatedCharacter.Aliases = translator.updateStringsBulk(input.Aliases, "aliases")

	ret := []*models.Character{}

	// Start the transaction and save the scenes
	if err := r.withTxn(ctx, func(ctx context.Context) error {
		qb := r.repository.Character

		for _, characterID := range characterIDs {
			if err := character.ValidateUpdate(ctx, characterID, updatedCharacter, qb); err != nil {
				return err
			}

			character, err := qb.UpdatePartial(ctx, characterID, updatedCharacter)
			if err != nil {
				return err
			}

			ret = append(ret, character)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// execute post hooks outside of txn
	var newRet []*models.Character
	for _, character := range ret {
		r.hookExecutor.ExecutePostHooks(ctx, character.ID, hook.CharacterUpdatePost, input, translator.getFields())

		character, err = r.getCharacter(ctx, character.ID)
		if err != nil {
			return nil, err
		}

		newRet = append(newRet, character)
	}

	return newRet, nil
}

func (r *mutationResolver) CharacterDestroy(ctx context.Context, input CharacterDestroyInput) (bool, error) {
	characterID, err := strconv.Atoi(input.ID)
	if err != nil {
		return false, fmt.Errorf("converting id: %w", err)
	}

	if err := r.withTxn(ctx, func(ctx context.Context) error {
		return r.repository.Character.Destroy(ctx, characterID)
	}); err != nil {
		return false, err
	}

	r.hookExecutor.ExecutePostHooks(ctx, characterID, hook.CharacterDestroyPost, input, nil)

	return true, nil
}

func (r *mutationResolver) CharactersDestroy(ctx context.Context, characterIDs []string) (bool, error) {
	ids, err := stringslice.StringSliceToIntSlice(characterIDs)
	if err != nil {
		return false, fmt.Errorf("converting ids: %w", err)
	}

	if err := r.withTxn(ctx, func(ctx context.Context) error {
		qb := r.repository.Character
		for _, id := range ids {
			if err := qb.Destroy(ctx, id); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return false, err
	}

	for _, id := range ids {
		r.hookExecutor.ExecutePostHooks(ctx, id, hook.CharacterDestroyPost, characterIDs, nil)
	}

	return true, nil
}
