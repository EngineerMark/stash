package character

import (
	"context"
	"fmt"

	"github.com/stashapp/stash/pkg/models"
)

type NameExistsError struct {
	Name string
}

func (e *NameExistsError) Error() string {
	return fmt.Sprintf("character with name '%s' already exists", e.Name)
}

type NameUsedByAliasError struct {
	Name     string
	OtherCharacter string
}

func (e *NameUsedByAliasError) Error() string {
	return fmt.Sprintf("name '%s' is used as alias for '%s'", e.Name, e.OtherCharacter)
}

type InvalidCharacterHierarchyError struct {
	Direction       string
	CurrentRelation string
	InvalidCharacter      string
	ApplyingCharacter     string
	CharacterPath         string
}

func (e *InvalidCharacterHierarchyError) Error() string {
	if e.ApplyingCharacter == "" {
		return fmt.Sprintf("cannot apply character \"%s\" as a %s of character as it is already %s", e.InvalidCharacter, e.Direction, e.CurrentRelation)
	}

	return fmt.Sprintf("cannot apply character \"%s\" as a %s of \"%s\" as it is already %s (%s)", e.InvalidCharacter, e.Direction, e.ApplyingCharacter, e.CurrentRelation, e.CharacterPath)
}

// EnsureCharacterNameUnique returns an error if the character name provided
// is used as a name or alias of another existing character.
func EnsureCharacterNameUnique(ctx context.Context, id int, name string, qb models.CharacterQueryer) error {
	// ensure name is unique
	sameNameCharacter, err := ByName(ctx, qb, name)
	if err != nil {
		return err
	}

	if sameNameCharacter != nil && id != sameNameCharacter.ID {
		return &NameExistsError{
			Name: name,
		}
	}

	// query by alias
	sameNameCharacter, err = ByAlias(ctx, qb, name)
	if err != nil {
		return err
	}

	if sameNameCharacter != nil && id != sameNameCharacter.ID {
		return &NameUsedByAliasError{
			Name:     name,
			OtherCharacter: sameNameCharacter.Name,
		}
	}

	return nil
}

func EnsureAliasesUnique(ctx context.Context, id int, aliases []string, qb models.CharacterQueryer) error {
	for _, a := range aliases {
		if err := EnsureCharacterNameUnique(ctx, id, a, qb); err != nil {
			return err
		}
	}

	return nil
}

type RelationshipFinder interface {
	FindAllAncestors(ctx context.Context, characterID int, excludeIDs []int) ([]*models.CharacterPath, error)
	FindAllDescendants(ctx context.Context, characterID int, excludeIDs []int) ([]*models.CharacterPath, error)
	models.CharacterRelationLoader
}

func ValidateHierarchyNew(ctx context.Context, parentIDs, childIDs []int, qb RelationshipFinder) error {
	allAncestors := make(map[int]*models.CharacterPath)
	allDescendants := make(map[int]*models.CharacterPath)

	for _, parentID := range parentIDs {
		parentsAncestors, err := qb.FindAllAncestors(ctx, parentID, nil)
		if err != nil {
			return err
		}

		for _, ancestorCharacter := range parentsAncestors {
			allAncestors[ancestorCharacter.ID] = ancestorCharacter
		}
	}

	for _, childID := range childIDs {
		childsDescendants, err := qb.FindAllDescendants(ctx, childID, nil)
		if err != nil {
			return err
		}

		for _, descendentCharacter := range childsDescendants {
			allDescendants[descendentCharacter.ID] = descendentCharacter
		}
	}

	// Validate that the character is not a parent of any of its ancestors
	validateParent := func(testID int) error {
		if parentCharacter, exists := allDescendants[testID]; exists {
			return &InvalidCharacterHierarchyError{
				Direction:       "parent",
				CurrentRelation: "a descendant",
				InvalidCharacter:      parentCharacter.Name,
				CharacterPath:         parentCharacter.Path,
			}
		}

		return nil
	}

	// Validate that the character is not a child of any of its ancestors
	validateChild := func(testID int) error {
		if childCharacter, exists := allAncestors[testID]; exists {
			return &InvalidCharacterHierarchyError{
				Direction:       "child",
				CurrentRelation: "an ancestor",
				InvalidCharacter:      childCharacter.Name,
				CharacterPath:         childCharacter.Path,
			}
		}

		return nil
	}

	for _, parentID := range parentIDs {
		if err := validateParent(parentID); err != nil {
			return err
		}
	}

	for _, childID := range childIDs {
		if err := validateChild(childID); err != nil {
			return err
		}
	}

	return nil
}

func ValidateHierarchyExisting(ctx context.Context, character *models.Character, parentIDs, childIDs []int, qb RelationshipFinder) error {
	allAncestors := make(map[int]*models.CharacterPath)
	allDescendants := make(map[int]*models.CharacterPath)

	parentsAncestors, err := qb.FindAllAncestors(ctx, character.ID, nil)
	if err != nil {
		return err
	}

	for _, ancestorCharacter := range parentsAncestors {
		allAncestors[ancestorCharacter.ID] = ancestorCharacter
	}

	childsDescendants, err := qb.FindAllDescendants(ctx, character.ID, nil)
	if err != nil {
		return err
	}

	for _, descendentCharacter := range childsDescendants {
		allDescendants[descendentCharacter.ID] = descendentCharacter
	}

	validateParent := func(testID int) error {
		if parentCharacter, exists := allDescendants[testID]; exists {
			return &InvalidCharacterHierarchyError{
				Direction:       "parent",
				CurrentRelation: "a descendant",
				InvalidCharacter:      parentCharacter.Name,
				ApplyingCharacter:     character.Name,
				CharacterPath:         parentCharacter.Path,
			}
		}

		return nil
	}

	validateChild := func(testID int) error {
		if childCharacter, exists := allAncestors[testID]; exists {
			return &InvalidCharacterHierarchyError{
				Direction:       "child",
				CurrentRelation: "an ancestor",
				InvalidCharacter:      childCharacter.Name,
				ApplyingCharacter:     character.Name,
				CharacterPath:         childCharacter.Path,
			}
		}

		return nil
	}

	for _, parentID := range parentIDs {
		if err := validateParent(parentID); err != nil {
			return err
		}
	}

	for _, childID := range childIDs {
		if err := validateChild(childID); err != nil {
			return err
		}
	}

	return nil
}

func MergeHierarchy(ctx context.Context, destination int, sources []int, qb RelationshipFinder) ([]int, []int, error) {
	var mergedParents, mergedChildren []int
	allIds := append([]int{destination}, sources...)

	addTo := func(mergedItems []int, characterIDs []int) []int {
	Characters:
		for _, characterID := range characterIDs {
			// Ignore characters which are already set
			for _, existingItem := range mergedItems {
				if characterID == existingItem {
					continue Characters
				}
			}

			// Ignore characters which are being merged, as these are rolled up anyway (if A is merged into B any direct link between them can be ignored)
			for _, id := range allIds {
				if characterID == id {
					continue Characters
				}
			}

			mergedItems = append(mergedItems, characterID)
		}

		return mergedItems
	}

	for _, id := range allIds {
		parents, err := qb.GetParentIDs(ctx, id)
		if err != nil {
			return nil, nil, err
		}

		mergedParents = addTo(mergedParents, parents)

		children, err := qb.GetChildIDs(ctx, id)
		if err != nil {
			return nil, nil, err
		}

		mergedChildren = addTo(mergedChildren, children)
	}

	return mergedParents, mergedChildren, nil
}
