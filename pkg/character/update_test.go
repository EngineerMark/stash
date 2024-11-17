package character

import (
	"context"
	"fmt"
	"testing"

	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/models/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testUniqueHierarchyCharacters = map[int]*models.Character{
	1: {
		ID:   1,
		Name: "one",
	},
	2: {
		ID:   2,
		Name: "two",
	},
	3: {
		ID:   3,
		Name: "three",
	},
	4: {
		ID:   4,
		Name: "four",
	},
}

var testUniqueHierarchyCharacterPaths = map[int]*models.CharacterPath{
	1: {
		Character: *testUniqueHierarchyCharacters[1],
	},
	2: {
		Character: *testUniqueHierarchyCharacters[2],
	},
	3: {
		Character: *testUniqueHierarchyCharacters[3],
	},
	4: {
		Character: *testUniqueHierarchyCharacters[4],
	},
}

type testUniqueHierarchyCase struct {
	id       int
	parents  []*models.Character
	children []*models.Character

	onFindAllAncestors   []*models.CharacterPath
	onFindAllDescendants []*models.CharacterPath

	expectedError string
}

var testUniqueHierarchyCases = []testUniqueHierarchyCase{
	{
		id:                   1,
		parents:              []*models.Character{},
		children:             []*models.Character{},
		onFindAllAncestors:   []*models.CharacterPath{},
		onFindAllDescendants: []*models.CharacterPath{},
		expectedError:        "",
	},
	{
		id:       1,
		parents:  []*models.Character{testUniqueHierarchyCharacters[2]},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		expectedError: "",
	},
	{
		id:       2,
		parents:  []*models.Character{testUniqueHierarchyCharacters[3]},
		children: make([]*models.Character, 0),
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		expectedError: "",
	},
	{
		id: 2,
		parents: []*models.Character{
			testUniqueHierarchyCharacters[3],
			testUniqueHierarchyCharacters[4],
		},
		children: []*models.Character{},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3], testUniqueHierarchyCharacterPaths[4],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		expectedError: "",
	},
	{
		id:       2,
		parents:  []*models.Character{},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		expectedError: "",
	},
	{
		id:      2,
		parents: []*models.Character{},
		children: []*models.Character{
			testUniqueHierarchyCharacters[3],
			testUniqueHierarchyCharacters[4],
		},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3], testUniqueHierarchyCharacterPaths[4],
		},
		expectedError: "",
	},
	{
		id:       1,
		parents:  []*models.Character{testUniqueHierarchyCharacters[2]},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2], testUniqueHierarchyCharacterPaths[3],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		expectedError: "cannot apply character \"three\" as a child of \"one\" as it is already an ancestor ()",
	},
	{
		id:       1,
		parents:  []*models.Character{testUniqueHierarchyCharacters[2]},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3], testUniqueHierarchyCharacterPaths[2],
		},
		expectedError: "cannot apply character \"two\" as a parent of \"one\" as it is already a descendant ()",
	},
	{
		id:       1,
		parents:  []*models.Character{testUniqueHierarchyCharacters[3]},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3],
		},
		expectedError: "cannot apply character \"three\" as a parent of \"one\" as it is already a descendant ()",
	},
	{
		id: 1,
		parents: []*models.Character{
			testUniqueHierarchyCharacters[2],
		},
		children: []*models.Character{
			testUniqueHierarchyCharacters[3],
		},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3], testUniqueHierarchyCharacterPaths[2],
		},
		expectedError: "cannot apply character \"two\" as a parent of \"one\" as it is already a descendant ()",
	},
	{
		id:       1,
		parents:  []*models.Character{testUniqueHierarchyCharacters[2]},
		children: []*models.Character{testUniqueHierarchyCharacters[2]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[2],
		},
		expectedError: "cannot apply character \"two\" as a parent of \"one\" as it is already a descendant ()",
	},
	{
		id:       2,
		parents:  []*models.Character{testUniqueHierarchyCharacters[1]},
		children: []*models.Character{testUniqueHierarchyCharacters[3]},
		onFindAllAncestors: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[1],
		},
		onFindAllDescendants: []*models.CharacterPath{
			testUniqueHierarchyCharacterPaths[3], testUniqueHierarchyCharacterPaths[1],
		},
		expectedError: "cannot apply character \"one\" as a parent of \"two\" as it is already a descendant ()",
	},
}

func TestEnsureHierarchy(t *testing.T) {
	for _, tc := range testUniqueHierarchyCases {
		testEnsureHierarchy(t, tc)
	}
}

func testEnsureHierarchy(t *testing.T, tc testUniqueHierarchyCase) {
	db := mocks.NewDatabase()

	var parentIDs, childIDs []int
	find := make(map[int]*models.Character)
	find[tc.id] = testUniqueHierarchyCharacters[tc.id]
	if tc.parents != nil {
		parentIDs = make([]int, 0)
		for _, parent := range tc.parents {
			if parent.ID != tc.id {
				find[parent.ID] = parent
				parentIDs = append(parentIDs, parent.ID)
			}
		}
	}

	if tc.children != nil {
		childIDs = make([]int, 0)
		for _, child := range tc.children {
			if child.ID != tc.id {
				find[child.ID] = child
				childIDs = append(childIDs, child.ID)
			}
		}
	}

	db.Character.On("FindAllAncestors", testCtx, mock.AnythingOfType("int"), []int(nil)).Return(func(ctx context.Context, characterID int, excludeIDs []int) []*models.CharacterPath {
		return tc.onFindAllAncestors
	}, func(ctx context.Context, characterID int, excludeIDs []int) error {
		if tc.onFindAllAncestors != nil {
			return nil
		}
		return fmt.Errorf("undefined ancestors for: %d", characterID)
	}).Maybe()

	db.Character.On("FindAllDescendants", testCtx, mock.AnythingOfType("int"), []int(nil)).Return(func(ctx context.Context, characterID int, excludeIDs []int) []*models.CharacterPath {
		return tc.onFindAllDescendants
	}, func(ctx context.Context, characterID int, excludeIDs []int) error {
		if tc.onFindAllDescendants != nil {
			return nil
		}
		return fmt.Errorf("undefined descendants for: %d", characterID)
	}).Maybe()

	res := ValidateHierarchyExisting(testCtx, testUniqueHierarchyCharacters[tc.id], parentIDs, childIDs, db.Character)

	assert := assert.New(t)

	if tc.expectedError != "" {
		if assert.NotNil(res) {
			assert.Equal(tc.expectedError, res.Error())
		}
	} else {
		assert.Nil(res)
	}

	db.AssertExpectations(t)
}
