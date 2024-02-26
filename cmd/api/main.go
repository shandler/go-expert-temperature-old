package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/shandler/go-expert-temperature/config"
	"github.com/shandler/go-expert-temperature/internal/dto"
	"github.com/shandler/go-expert-temperature/internal/entity"
	"github.com/shandler/go-expert-temperature/internal/infra/api"
	"github.com/shandler/go-expert-temperature/internal/usecase"

	"github.com/go-chi/chi/v5"
)

func main() {
	var cfg config.Config
	viperCfg := config.NewViper(
		fmt.Sprintf(
			"env.%s",
			os.Getenv("ENVIRONMENT"),
		),
	)
	viperCfg.ReadViper(&cfg)

	gw := usecase.NewGetWeather(
		api.NewCEPFromAPI(&cfg),
		api.NewWeatherFromAPI(&cfg),
	)

	r := chi.NewRouter()
	r.Get(
		"/temperature",
		func(w http.ResponseWriter, r *http.Request) {
			temperature, execErr := gw.Execute(r.Context(), dto.LocationInput{
				CEP: r.URL.Query().Get("cep"),
			})
			switch {
			case errors.Is(execErr, entity.ErrCEPNotValid):
				http.Error(w, entity.ErrCEPNotValid.Error(), http.StatusUnprocessableEntity)
				return
			case errors.Is(execErr, entity.ErrCEPNotFound):
				http.Error(w, entity.ErrCEPNotFound.Error(), http.StatusNotFound)
				return
			case execErr != nil:
				http.Error(w, execErr.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(temperature); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
		},
	)

	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		return
	}
}
