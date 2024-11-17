package models

import "context"

// CharacterGetter provides methods to get characters by ID.
type CharacterGetter interface {
	// TODO - rename this to Find and remove existing method
	FindMany(ctx context.Context, ids []int) ([]*Character, error)
	Find(ctx context.Context, id int) (*Character, error)
}

// CharacterFinder provides methods to find characters.
type CharacterFinder interface {
	CharacterGetter
	FindAllAncestors(ctx context.Context, characterID int, excludeIDs []int) ([]*CharacterPath, error)
	FindAllDescendants(ctx context.Context, characterID int, excludeIDs []int) ([]*CharacterPath, error)
	FindByParentCharacterID(ctx context.Context, parentID int) ([]*Character, error)
	FindByChildCharacterID(ctx context.Context, childID int) ([]*Character, error)
	FindBySceneID(ctx context.Context, sceneID int) ([]*Character, error)
	FindByImageID(ctx context.Context, imageID int) ([]*Character, error)
	FindByGalleryID(ctx context.Context, galleryID int) ([]*Character, error)
	FindByPerformerID(ctx context.Context, performerID int) ([]*Character, error)
	FindByGroupID(ctx context.Context, groupID int) ([]*Character, error)
	FindBySceneMarkerID(ctx context.Context, sceneMarkerID int) ([]*Character, error)
	FindByStudioID(ctx context.Context, studioID int) ([]*Character, error)
	FindByTagID(ctx context.Context, tagID int) ([]*Character, error)
	FindByName(ctx context.Context, name string, nocase bool) (*Character, error)
	FindByNames(ctx context.Context, names []string, nocase bool) ([]*Character, error)
}

// CharacterQueryer provides methods to query characters.
type CharacterQueryer interface {
	Query(ctx context.Context, characterFilter *CharacterFilterType, findFilter *FindFilterType) ([]*Character, int, error)
}

type CharacterAutoCharacterQueryer interface {
	CharacterQueryer
	AliasLoader

	// TODO - this interface is temporary until the filter schema can fully
	// support the query needed
	QueryForAutoCharacter(ctx context.Context, words []string) ([]*Character, error)
}

// CharacterCounter provides methods to count characters.
type CharacterCounter interface {
	Count(ctx context.Context) (int, error)
	CountByParentCharacterID(ctx context.Context, parentID int) (int, error)
	CountByChildCharacterID(ctx context.Context, childID int) (int, error)
}

// CharacterCreator provides methods to create characters.
type CharacterCreator interface {
	Create(ctx context.Context, newCharacter *Character) error
}

// CharacterUpdater provides methods to update characters.
type CharacterUpdater interface {
	Update(ctx context.Context, updatedCharacter *Character) error
	UpdatePartial(ctx context.Context, id int, updateCharacter CharacterPartial) (*Character, error)
	UpdateAliases(ctx context.Context, characterID int, aliases []string) error
	UpdateImage(ctx context.Context, characterID int, image []byte) error
	UpdateParentCharacters(ctx context.Context, characterID int, parentIDs []int) error
	UpdateChildCharacters(ctx context.Context, characterID int, parentIDs []int) error
}

// CharacterDestroyer provides methods to destroy characters.
type CharacterDestroyer interface {
	Destroy(ctx context.Context, id int) error
}

type CharacterFinderCreator interface {
	CharacterFinder
	CharacterCreator
}

type CharacterCreatorUpdater interface {
	CharacterCreator
	CharacterUpdater
}

// CharacterReader provides all methods to read characters.
type CharacterReader interface {
	CharacterFinder
	CharacterQueryer
	CharacterAutoCharacterQueryer
	CharacterCounter

	AliasLoader
	CharacterRelationLoader

	All(ctx context.Context) ([]*Character, error)
	GetImage(ctx context.Context, characterID int) ([]byte, error)
	HasImage(ctx context.Context, characterID int) (bool, error)
}

// CharacterWriter provides all methods to modify characters.
type CharacterWriter interface {
	CharacterCreator
	CharacterUpdater
	CharacterDestroyer

	Merge(ctx context.Context, source []int, destination int) error
}

// CharacterReaderWriter provides all characters methods.
type CharacterReaderWriter interface {
	CharacterReader
	CharacterWriter
}
