package models

import (
	"context"
	"time"
)

type Character struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Favorite      bool      `json:"favorite"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Aliases   RelatedStrings `json:"aliases"`
	ParentIDs RelatedIDs     `json:"parent_ids"`
	ChildIDs  RelatedIDs     `json:"character_ids"`
}

func NewCharacter() Character {
	currentTime := time.Now()
	return Character{
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
}

func (s *Character) LoadAliases(ctx context.Context, l AliasLoader) error {
	return s.Aliases.load(func() ([]string, error) {
		return l.GetAliases(ctx, s.ID)
	})
}

func (s *Character) LoadParentIDs(ctx context.Context, l CharacterRelationLoader) error {
	return s.ParentIDs.load(func() ([]int, error) {
		return l.GetParentIDs(ctx, s.ID)
	})
}

func (s *Character) LoadChildIDs(ctx context.Context, l CharacterRelationLoader) error {
	return s.ChildIDs.load(func() ([]int, error) {
		return l.GetChildIDs(ctx, s.ID)
	})
}

type CharacterPartial struct {
	Name          OptionalString
	Description   OptionalString
	Favorite      OptionalBool
	CreatedAt     OptionalTime
	UpdatedAt     OptionalTime

	Aliases   *UpdateStrings
	ParentIDs *UpdateIDs
	ChildIDs  *UpdateIDs
}

func NewCharacterPartial() CharacterPartial {
	currentTime := time.Now()
	return CharacterPartial{
		UpdatedAt: NewOptionalTime(currentTime),
	}
}

type CharacterPath struct {
	Character
	Path string `json:"path"`
}
