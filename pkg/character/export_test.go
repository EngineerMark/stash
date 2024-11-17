package character

import (
	"errors"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/json"
	"github.com/stashapp/stash/pkg/models/jsonschema"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

const (
	characterID         = 1
	noImageID     = 2
	errImageID    = 3
	errAliasID    = 4
	withParentsID = 5
	errParentsID  = 6
)

const (
	characterName     = "testCharacter"
	description = "description"
)

var (
	autoCharacterIgnored = true
	createTime     = time.Date(2001, 01, 01, 0, 0, 0, 0, time.UTC)
	updateTime     = time.Date(2002, 01, 01, 0, 0, 0, 0, time.UTC)
)

func createCharacter(id int) models.Character {
	return models.Character{
		ID:            id,
		Name:          characterName,
		Favorite:      true,
		Description:   description,
		CreatedAt:     createTime,
		UpdatedAt:     updateTime,
	}
}

func createJSONCharacter(aliases []string, image string, parents []string) *jsonschema.Character {
	return &jsonschema.Character{
		Name:          characterName,
		Favorite:      true,
		Description:   description,
		Aliases:       aliases,
		CreatedAt: json.JSONTime{
			Time: createTime,
		},
		UpdatedAt: json.JSONTime{
			Time: updateTime,
		},
		Image:   image,
		Parents: parents,
	}
}

type testScenario struct {
	character      models.Character
	expected *jsonschema.Character
	err      bool
}

var scenarios []testScenario

func initTestTable() {
	scenarios = []testScenario{
		{
			createCharacter(characterID),
			createJSONCharacter([]string{"alias"}, image, nil),
			false,
		},
		{
			createCharacter(noImageID),
			createJSONCharacter(nil, "", nil),
			false,
		},
		{
			createCharacter(errImageID),
			createJSONCharacter(nil, "", nil),
			// getting the image should not cause an error
			false,
		},
		{
			createCharacter(errAliasID),
			nil,
			true,
		},
		{
			createCharacter(withParentsID),
			createJSONCharacter(nil, image, []string{"parent"}),
			false,
		},
		{
			createCharacter(errParentsID),
			nil,
			true,
		},
	}
}

func TestToJSON(t *testing.T) {
	initTestTable()

	db := mocks.NewDatabase()

	imageErr := errors.New("error getting image")
	aliasErr := errors.New("error getting aliases")
	parentsErr := errors.New("error getting parents")

	db.Character.On("GetAliases", testCtx, characterID).Return([]string{"alias"}, nil).Once()
	db.Character.On("GetAliases", testCtx, noImageID).Return(nil, nil).Once()
	db.Character.On("GetAliases", testCtx, errImageID).Return(nil, nil).Once()
	db.Character.On("GetAliases", testCtx, errAliasID).Return(nil, aliasErr).Once()
	db.Character.On("GetAliases", testCtx, withParentsID).Return(nil, nil).Once()
	db.Character.On("GetAliases", testCtx, errParentsID).Return(nil, nil).Once()

	db.Character.On("GetImage", testCtx, characterID).Return(imageBytes, nil).Once()
	db.Character.On("GetImage", testCtx, noImageID).Return(nil, nil).Once()
	db.Character.On("GetImage", testCtx, errImageID).Return(nil, imageErr).Once()
	db.Character.On("GetImage", testCtx, withParentsID).Return(imageBytes, nil).Once()
	db.Character.On("GetImage", testCtx, errParentsID).Return(nil, nil).Once()

	db.Character.On("FindByChildCharacterID", testCtx, characterID).Return(nil, nil).Once()
	db.Character.On("FindByChildCharacterID", testCtx, noImageID).Return(nil, nil).Once()
	db.Character.On("FindByChildCharacterID", testCtx, withParentsID).Return([]*models.Character{{Name: "parent"}}, nil).Once()
	db.Character.On("FindByChildCharacterID", testCtx, errParentsID).Return(nil, parentsErr).Once()
	db.Character.On("FindByChildCharacterID", testCtx, errImageID).Return(nil, nil).Once()

	for i, s := range scenarios {
		character := s.character
		json, err := ToJSON(testCtx, db.Character, &character)

		switch {
		case !s.err && err != nil:
			t.Errorf("[%d] unexpected error: %s", i, err.Error())
		case s.err && err == nil:
			t.Errorf("[%d] expected error not returned", i)
		default:
			assert.Equal(t, s.expected, json, "[%d]", i)
		}
	}

	db.AssertExpectations(t)
}
