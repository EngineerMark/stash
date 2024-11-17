package api

import (
	"context"

	"github.com/stashapp/stash/internal/api/loaders"
	"github.com/stashapp/stash/internal/api/urlbuilders"
	"github.com/stashapp/stash/pkg/gallery"
	"github.com/stashapp/stash/pkg/group"
	"github.com/stashapp/stash/pkg/image"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/performer"
	"github.com/stashapp/stash/pkg/scene"
	"github.com/stashapp/stash/pkg/studio"
)

func (r *characterResolver) Parents(ctx context.Context, obj *models.Character) (ret []*models.Character, err error) {
	if !obj.ParentIDs.Loaded() {
		if err := r.withReadTxn(ctx, func(ctx context.Context) error {
			return obj.LoadParentIDs(ctx, r.repository.Character)
		}); err != nil {
			return nil, err
		}
	}

	var errs []error
	ret, errs = loaders.From(ctx).CharacterByID.LoadAll(obj.ParentIDs.List())
	return ret, firstError(errs)
}

func (r *characterResolver) Children(ctx context.Context, obj *models.Character) (ret []*models.Character, err error) {
	if !obj.ChildIDs.Loaded() {
		if err := r.withReadTxn(ctx, func(ctx context.Context) error {
			return obj.LoadChildIDs(ctx, r.repository.Character)
		}); err != nil {
			return nil, err
		}
	}

	var errs []error
	ret, errs = loaders.From(ctx).CharacterByID.LoadAll(obj.ChildIDs.List())
	return ret, firstError(errs)
}

func (r *characterResolver) Aliases(ctx context.Context, obj *models.Character) (ret []string, err error) {
	if !obj.Aliases.Loaded() {
		if err := r.withReadTxn(ctx, func(ctx context.Context) error {
			return obj.LoadAliases(ctx, r.repository.Character)
		}); err != nil {
			return nil, err
		}
	}

	return obj.Aliases.List(), nil
}

func (r *characterResolver) SceneCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = scene.CountByCharacterID(ctx, r.repository.Scene, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) SceneMarkerCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = scene.MarkerCountByCharacterID(ctx, r.repository.SceneMarker, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) ImageCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = image.CountByCharacterID(ctx, r.repository.Image, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) GalleryCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = gallery.CountByCharacterID(ctx, r.repository.Gallery, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) PerformerCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = performer.CountByCharacterID(ctx, r.repository.Performer, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) StudioCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = studio.CountByCharacterID(ctx, r.repository.Studio, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) GroupCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = group.CountByCharacterID(ctx, r.repository.Group, obj.ID, depth)
		return err
	}); err != nil {
		return 0, err
	}

	return ret, nil
}

func (r *characterResolver) MovieCount(ctx context.Context, obj *models.Character, depth *int) (ret int, err error) {
	return r.GroupCount(ctx, obj, depth)
}

func (r *characterResolver) ImagePath(ctx context.Context, obj *models.Character) (*string, error) {
	var hasImage bool
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		var err error
		hasImage, err = r.repository.Character.HasImage(ctx, obj.ID)
		return err
	}); err != nil {
		return nil, err
	}

	baseURL, _ := ctx.Value(BaseURLCtxKey).(string)
	imagePath := urlbuilders.NewCharacterURLBuilder(baseURL, obj).GetCharacterImageURL(hasImage)
	return &imagePath, nil
}

func (r *characterResolver) ParentCount(ctx context.Context, obj *models.Character) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.Character.CountByParentCharacterID(ctx, obj.ID)
		return err
	}); err != nil {
		return ret, err
	}

	return ret, nil
}

func (r *characterResolver) ChildCount(ctx context.Context, obj *models.Character) (ret int, err error) {
	if err := r.withReadTxn(ctx, func(ctx context.Context) error {
		ret, err = r.repository.Character.CountByChildCharacterID(ctx, obj.ID)
		return err
	}); err != nil {
		return ret, err
	}

	return ret, nil
}
