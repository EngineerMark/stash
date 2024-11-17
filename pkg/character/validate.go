package character

import (
	"context"
	"errors"
	"fmt"

	"github.com/stashapp/stash/pkg/models"
)

var (
	ErrNameMissing = errors.New("character name must not be blank")
)

type NotFoundError struct {
	id int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("character with id %d not found", e.id)
}

func ValidateCreate(ctx context.Context, character models.Character, qb models.CharacterReader) error {
	if character.Name == "" {
		return ErrNameMissing
	}

	if err := EnsureCharacterNameUnique(ctx, 0, character.Name, qb); err != nil {
		return err
	}

	if character.Aliases.Loaded() {
		if err := EnsureAliasesUnique(ctx, character.ID, character.Aliases.List(), qb); err != nil {
			return err
		}
	}

	return nil
}

func ValidateUpdate(ctx context.Context, id int, partial models.CharacterPartial, qb models.CharacterReader) error {
	existing, err := qb.Find(ctx, id)
	if err != nil {
		return err
	}

	if existing == nil {
		return &NotFoundError{id}
	}

	if partial.Name.Set {
		if partial.Name.Value == "" {
			return ErrNameMissing
		}

		if err := EnsureCharacterNameUnique(ctx, id, partial.Name.Value, qb); err != nil {
			return err
		}
	}

	if partial.Aliases != nil {
		if err := existing.LoadAliases(ctx, qb); err != nil {
			return err
		}

		if err := EnsureAliasesUnique(ctx, id, partial.Aliases.Apply(existing.Aliases.List()), qb); err != nil {
			return err
		}
	}

	return nil
}
