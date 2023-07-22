package delete

import (
	"net/http"
	"url-shortner/internal/lib/api/response"
	"url-shortner/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
)

type Request struct {
	Alias string `json:"alias" validate:"required, len-6"`
}

type Response struct {
	response.Response
	DeletedCount int64
}

type URLDeleter interface {
	DeleteURL(alias string) (int64, error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validatorErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(validatorErr))
			return
		}

		if req.Alias == "" {
			log.Error("invalid request")
			render.JSON(w, r, response.Error("alias is empty"))
			return
		}

		count, err := urlDeleter.DeleteURL(req.Alias)
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))

			render.JSON(w, r, response.Error("failed to delete url"))
			return
		}

		render.JSON(w, r, Response{
			Response:     response.OK(),
			DeletedCount: count,
		})
	}
}
