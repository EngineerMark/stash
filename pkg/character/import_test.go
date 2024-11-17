package character

import (
	"context"
	"errors"
	"testing"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const image = "aW1hZ2VCeXRlcw=="
const invalidImage = "aW1hZ2VCeXRlcw&&"

var imageBytes = []byte("imageBytes")

const (
	characterNameErr      = "characterNameErr"
	existingCharacterName = "existingCharacterName"

	existingCharacterID = 100
)

var testCtx = context.Background()

func TestImporterName(t *testing.T) {
	i := Importer{
		Input: jsonschema.Character{
			Name: characterName,
		},
	}

	assert.Equal(t, characterName, i.Name())
}

func TestImporterPreImport(t *testing.T) {
	i := Importer{
		Input: jsonschema.Character{
			Name:          characterName,
			Description:   description,
			Image:         invalidImage,
		},
	}

	err := i.PreImport(testCtx)

	assert.NotNil(t, err)

	i.Input.Image = image

	err = i.PreImport(testCtx)

	assert.Nil(t, err)
}

func TestImporterPostImport(t *testing.T) {
	db := mocks.NewDatabase()

	i := Importer{
		ReaderWriter: db.Character,
		Input: jsonschema.Character{
			Aliases: []string{"alias"},
		},
		imageData: imageBytes,
	}

	updateCharacterImageErr := errors.New("UpdateImage error")
	updateCharacterAliasErr := errors.New("UpdateAlias error")
	updateCharacterParentsErr := errors.New("UpdateParentCharacters error")

	db.Character.On("UpdateAliases", testCtx, characterID, i.Input.Aliases).Return(nil).Once()
	db.Character.On("UpdateAliases", testCtx, errAliasID, i.Input.Aliases).Return(updateCharacterAliasErr).Once()
	db.Character.On("UpdateAliases", testCtx, withParentsID, i.Input.Aliases).Return(nil).Once()
	db.Character.On("UpdateAliases", testCtx, errParentsID, i.Input.Aliases).Return(nil).Once()

	db.Character.On("UpdateImage", testCtx, characterID, imageBytes).Return(nil).Once()
	db.Character.On("UpdateImage", testCtx, errAliasID, imageBytes).Return(nil).Once()
	db.Character.On("UpdateImage", testCtx, errImageID, imageBytes).Return(updateCharacterImageErr).Once()
	db.Character.On("UpdateImage", testCtx, withParentsID, imageBytes).Return(nil).Once()
	db.Character.On("UpdateImage", testCtx, errParentsID, imageBytes).Return(nil).Once()

	var parentCharacters []int
	db.Character.On("UpdateParentCharacters", testCtx, characterID, parentCharacters).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, withParentsID, []int{100}).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, errParentsID, []int{100}).Return(updateCharacterParentsErr).Once()

	db.Character.On("FindByName", testCtx, "Parent", false).Return(&models.Character{ID: 100}, nil)

	err := i.PostImport(testCtx, characterID)
	assert.Nil(t, err)

	err = i.PostImport(testCtx, errImageID)
	assert.NotNil(t, err)

	err = i.PostImport(testCtx, errAliasID)
	assert.NotNil(t, err)

	i.Input.Parents = []string{"Parent"}
	err = i.PostImport(testCtx, withParentsID)
	assert.Nil(t, err)

	err = i.PostImport(testCtx, errParentsID)
	assert.NotNil(t, err)

	db.AssertExpectations(t)
}

func TestImporterPostImportParentMissing(t *testing.T) {
	db := mocks.NewDatabase()

	i := Importer{
		ReaderWriter: db.Character,
		Input:        jsonschema.Character{},
		imageData:    imageBytes,
	}

	createID := 1
	createErrorID := 2
	createFindErrorID := 3
	createFoundID := 4
	failID := 5
	failFindErrorID := 6
	failFoundID := 7
	ignoreID := 8
	ignoreFindErrorID := 9
	ignoreFoundID := 10

	findError := errors.New("failed finding parent")

	var emptyParents []int

	db.Character.On("UpdateImage", testCtx, mock.Anything, mock.Anything).Return(nil)
	db.Character.On("UpdateAliases", testCtx, mock.Anything, mock.Anything).Return(nil)

	db.Character.On("FindByName", testCtx, "Create", false).Return(nil, nil).Once()
	db.Character.On("FindByName", testCtx, "CreateError", false).Return(nil, nil).Once()
	db.Character.On("FindByName", testCtx, "CreateFindError", false).Return(nil, findError).Once()
	db.Character.On("FindByName", testCtx, "CreateFound", false).Return(&models.Character{ID: 101}, nil).Once()
	db.Character.On("FindByName", testCtx, "Fail", false).Return(nil, nil).Once()
	db.Character.On("FindByName", testCtx, "FailFindError", false).Return(nil, findError)
	db.Character.On("FindByName", testCtx, "FailFound", false).Return(&models.Character{ID: 102}, nil).Once()
	db.Character.On("FindByName", testCtx, "Ignore", false).Return(nil, nil).Once()
	db.Character.On("FindByName", testCtx, "IgnoreFindError", false).Return(nil, findError)
	db.Character.On("FindByName", testCtx, "IgnoreFound", false).Return(&models.Character{ID: 103}, nil).Once()

	db.Character.On("UpdateParentCharacters", testCtx, createID, []int{100}).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, createFoundID, []int{101}).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, failFoundID, []int{102}).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, ignoreID, emptyParents).Return(nil).Once()
	db.Character.On("UpdateParentCharacters", testCtx, ignoreFoundID, []int{103}).Return(nil).Once()

	db.Character.On("Create", testCtx, mock.MatchedBy(func(t *models.Character) bool {
		return t.Name == "Create"
	})).Run(func(args mock.Arguments) {
		t := args.Get(1).(*models.Character)
		t.ID = 100
	}).Return(nil).Once()
	db.Character.On("Create", testCtx, mock.MatchedBy(func(t *models.Character) bool {
		return t.Name == "CreateError"
	})).Return(errors.New("failed creating parent")).Once()

	i.MissingRefBehaviour = models.ImportMissingRefEnumCreate

	db.AssertExpectations(t)
}

func TestImporterFindExistingID(t *testing.T) {
	db := mocks.NewDatabase()

	i := Importer{
		ReaderWriter: db.Character,
		Input: jsonschema.Character{
			Name: characterName,
		},
	}

	errFindByName := errors.New("FindByName error")
	db.Character.On("FindByName", testCtx, characterName, false).Return(nil, nil).Once()
	db.Character.On("FindByName", testCtx, existingCharacterName, false).Return(&models.Character{
		ID: existingCharacterID,
	}, nil).Once()
	db.Character.On("FindByName", testCtx, characterNameErr, false).Return(nil, errFindByName).Once()

	id, err := i.FindExistingID(testCtx)
	assert.Nil(t, id)
	assert.Nil(t, err)

	i.Input.Name = existingCharacterName
	id, err = i.FindExistingID(testCtx)
	assert.Equal(t, existingCharacterID, *id)
	assert.Nil(t, err)

	i.Input.Name = characterNameErr
	id, err = i.FindExistingID(testCtx)
	assert.Nil(t, id)
	assert.NotNil(t, err)

	db.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
	db := mocks.NewDatabase()

	character := models.Character{
		Name: characterName,
	}

	characterErr := models.Character{
		Name: characterNameErr,
	}

	i := Importer{
		ReaderWriter: db.Character,
		character:          character,
	}

	errCreate := errors.New("Create error")
	db.Character.On("Create", testCtx, &character).Run(func(args mock.Arguments) {
		t := args.Get(1).(*models.Character)
		t.ID = characterID
	}).Return(nil).Once()
	db.Character.On("Create", testCtx, &characterErr).Return(errCreate).Once()

	id, err := i.Create(testCtx)
	assert.Equal(t, characterID, *id)
	assert.Nil(t, err)

	i.character = characterErr
	id, err = i.Create(testCtx)
	assert.Nil(t, id)
	assert.NotNil(t, err)

	db.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	db := mocks.NewDatabase()

	character := models.Character{
		Name: characterName,
	}

	characterErr := models.Character{
		Name: characterNameErr,
	}

	i := Importer{
		ReaderWriter: db.Character,
		character:          character,
	}

	errUpdate := errors.New("Update error")

	// id needs to be set for the mock input
	character.ID = characterID
	db.Character.On("Update", testCtx, &character).Return(nil).Once()

	err := i.Update(testCtx, characterID)
	assert.Nil(t, err)

	i.character = characterErr

	// need to set id separately
	characterErr.ID = errImageID
	db.Character.On("Update", testCtx, &characterErr).Return(errUpdate).Once()

	err = i.Update(testCtx, errImageID)
	assert.NotNil(t, err)

	db.AssertExpectations(t)
}
