package character

import (
	"context"
	"fmt"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/utils"
)

type ImporterReaderWriter interface {
	models.CharacterCreatorUpdater
	FindByName(ctx context.Context, name string, nocase bool) (*models.Character, error)
}

type ParentCharacterNotExistError struct {
	missingParent string
}

func (e ParentCharacterNotExistError) Error() string {
	return fmt.Sprintf("parent character <%s> does not exist", e.missingParent)
}

func (e ParentCharacterNotExistError) MissingParent() string {
	return e.missingParent
}

type Importer struct {
	ReaderWriter        ImporterReaderWriter
	Input               jsonschema.Character
	MissingRefBehaviour models.ImportMissingRefEnum

	character       models.Character
	imageData []byte
}

func (i *Importer) PreImport(ctx context.Context) error {
	i.character = models.Character{
		Name:          i.Input.Name,
		Description:   i.Input.Description,
		Favorite:      i.Input.Favorite,
		CreatedAt:     i.Input.CreatedAt.GetTime(),
		UpdatedAt:     i.Input.UpdatedAt.GetTime(),
	}

	var err error
	if len(i.Input.Image) > 0 {
		i.imageData, err = utils.ProcessBase64Image(i.Input.Image)
		if err != nil {
			return fmt.Errorf("invalid image: %v", err)
		}
	}

	return nil
}

func (i *Importer) PostImport(ctx context.Context, id int) error {
	if len(i.imageData) > 0 {
		if err := i.ReaderWriter.UpdateImage(ctx, id, i.imageData); err != nil {
			return fmt.Errorf("error setting character image: %v", err)
		}
	}

	if err := i.ReaderWriter.UpdateAliases(ctx, id, i.Input.Aliases); err != nil {
		return fmt.Errorf("error setting character aliases: %v", err)
	}

	parents, err := i.getParents(ctx)
	if err != nil {
		return err
	}

	if err := i.ReaderWriter.UpdateParentCharacters(ctx, id, parents); err != nil {
		return fmt.Errorf("error setting parents: %v", err)
	}

	return nil
}

func (i *Importer) Name() string {
	return i.Input.Name
}

func (i *Importer) FindExistingID(ctx context.Context) (*int, error) {
	const nocase = false
	existing, err := i.ReaderWriter.FindByName(ctx, i.Name(), nocase)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		id := existing.ID
		return &id, nil
	}

	return nil, nil
}

func (i *Importer) Create(ctx context.Context) (*int, error) {
	err := i.ReaderWriter.Create(ctx, &i.character)
	if err != nil {
		return nil, fmt.Errorf("error creating character: %v", err)
	}

	id := i.character.ID
	return &id, nil
}

func (i *Importer) Update(ctx context.Context, id int) error {
	character := i.character
	character.ID = id
	err := i.ReaderWriter.Update(ctx, &character)
	if err != nil {
		return fmt.Errorf("error updating existing character: %v", err)
	}

	return nil
}

func (i *Importer) getParents(ctx context.Context) ([]int, error) {
	var parents []int
	for _, parent := range i.Input.Parents {
		character, err := i.ReaderWriter.FindByName(ctx, parent, false)
		if err != nil {
			return nil, fmt.Errorf("error finding parent by name: %v", err)
		}

		if character == nil {
			if i.MissingRefBehaviour == models.ImportMissingRefEnumFail {
				return nil, ParentCharacterNotExistError{missingParent: parent}
			}

			if i.MissingRefBehaviour == models.ImportMissingRefEnumIgnore {
				continue
			}

			if i.MissingRefBehaviour == models.ImportMissingRefEnumCreate {
				parentID, err := i.createParent(ctx, parent)
				if err != nil {
					return nil, err
				}
				parents = append(parents, parentID)
			}
		} else {
			parents = append(parents, character.ID)
		}
	}

	return parents, nil
}

func (i *Importer) createParent(ctx context.Context, name string) (int, error) {
	newCharacter := models.NewCharacter()
	newCharacter.Name = name

	err := i.ReaderWriter.Create(ctx, &newCharacter)
	if err != nil {
		return 0, err
	}

	return newCharacter.ID, nil
}
