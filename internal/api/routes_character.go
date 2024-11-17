package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/stashapp/stash/internal/static"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/utils"
)

type CharacterFinder interface {
	models.CharacterGetter
	GetImage(ctx context.Context, characterID int) ([]byte, error)
}

type characterRoutes struct {
	routes
	characterFinder CharacterFinder
}

func (rs characterRoutes) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{characterId}", func(r chi.Router) {
		r.Use(rs.CharacterCtx)
		r.Get("/image", rs.Image)
	})

	return r
}

func (rs characterRoutes) Image(w http.ResponseWriter, r *http.Request) {
	character := r.Context().Value(characterKey).(*models.Character)
	defaultParam := r.URL.Query().Get("default")

	var image []byte
	if defaultParam != "true" {
		readTxnErr := rs.withReadTxn(r, func(ctx context.Context) error {
			var err error
			image, err = rs.characterFinder.GetImage(ctx, character.ID)
			return err
		})
		if errors.Is(readTxnErr, context.Canceled) {
			return
		}
		if readTxnErr != nil {
			logger.Warnf("read transaction error on fetch character image: %v", readTxnErr)
		}
	}

	// fallback to default image
	if len(image) == 0 {
		image = static.ReadAll(static.DefaultCharacterImage)
	}

	utils.ServeImage(w, r, image)
}

func (rs characterRoutes) CharacterCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		characterID, err := strconv.Atoi(chi.URLParam(r, "characterId"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		var character *models.Character
		_ = rs.withReadTxn(r, func(ctx context.Context) error {
			var err error
			character, err = rs.characterFinder.Find(ctx, characterID)
			return err
		})
		if character == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		ctx := context.WithValue(r.Context(), characterKey, character)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
