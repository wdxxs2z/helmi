package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/rs/cors"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/monostream/helmi/pkg/catalog"
	"github.com/monostream/helmi/pkg/release"
)

type App struct {
	Catalog catalog.Catalog

	Router *mux.Router
}

func (a *App) Initialize(path string) {
	a.Catalog.Parse(path)

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	var handler http.Handler

	handler = a.Router
	handler = handlers.ProxyHeaders(handler)
	handler = handlers.CompressHandler(handler)

	handler = cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	}).Handler(handler)

	os.Stdout.WriteString("Helmi is ready and available on port " + strings.TrimPrefix(addr, ":"))
	log.Fatal(http.ListenAndServe(addr, handler))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/v2/catalog", auth(a.getCatalog)).Methods(http.MethodGet)
	a.Router.HandleFunc("/v2/service_instances/{serviceId}", auth(a.createInstance)).Methods(http.MethodPut)
	a.Router.HandleFunc("/v2/service_instances/{serviceId}", auth(a.deleteInstance)).Methods(http.MethodDelete)

	a.Router.HandleFunc("/v2/service_instances/{serviceId}/last_operation", auth(a.queryInstance)).Methods(http.MethodGet)

	a.Router.HandleFunc("/v2/service_instances/{serviceId}/service_bindings/{bindingId}", auth(a.bindInstance)).Methods(http.MethodPut)
	a.Router.HandleFunc("/v2/service_instances/{serviceId}/service_bindings/{bindingId}", auth(a.unbindInstance)).Methods(http.MethodDelete)

	// endpoint to check if webservice is up
	a.Router.HandleFunc("/liveness", a.livenessCheck).Methods(http.MethodGet)
}

// used by kubernetes
// if this fails kubernetes will restart the container
func (a *App) livenessCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, nil)
}

func checkCredentials(username string, password string) bool {
	// if env variables are empty or not set ignore credentials
	if user, isUserSet := os.LookupEnv("USERNAME"); isUserSet && len(user) > 0 {
		if pass, isPassSet := os.LookupEnv("PASSWORD"); isPassSet && (pass != password || user != username) {
			return false
		}
	}
	return true
}

func auth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()

		if !checkCredentials(user, pass) {
			http.Error(w, "Unauthorized.", 401)
			return
		}

		handler(w, r)
	}
}

func (a *App) getCatalog(w http.ResponseWriter, r *http.Request) {
	type PlanEntry struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`

		IsFree      bool `json:"free"`
		IsBindable  bool `json:"bindable"`
	}

	type ServiceEntry struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`

		IsBindable  bool `json:"bindable"`
		IsUpdatable bool `json:"plan_updateable"`

		Plans []    PlanEntry `json:"plans"`
	}

	type Services struct {
		Services [] ServiceEntry `json:"services"`
	}

	var serviceEntries [] ServiceEntry

	for _, service := range a.Catalog.Services {
		serviceEntry := ServiceEntry{
			Id:   service.Id,
			Name: service.Name,

			Description: service.Description,

			IsBindable:  true,
			IsUpdatable: false,
		}

		var planEntries [] PlanEntry

		for _, plan := range service.Plans {
			planEntry := PlanEntry{
				Id:   plan.Id,
				Name: plan.Name,

				Description: plan.Description,

				IsFree:     true,
				IsBindable: true,
			}

			planEntries = append(planEntries, planEntry)
		}

		serviceEntry.Plans = planEntries

		serviceEntries = append(serviceEntries, serviceEntry)
	}
	var services = Services{
		Services: serviceEntries,
	}

	respondWithJSON(w, http.StatusOK, services)
}

func (a *App) createInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceId := vars["serviceId"]
	acceptsIncomplete := strings.EqualFold(r.URL.Query().Get("accepts_incomplete"), "true")

	type requestData struct {
		ServiceId string `json:"service_id"`
		PlanId    string `json:"plan_id"`
	}

	var data requestData
	decoder := json.NewDecoder(r.Body)
	decoderErr := decoder.Decode(&data)

	if decoderErr != nil || len(data.ServiceId) == 0 || len(data.PlanId) == 0 {
		respondWithUserError(w, "Invalid Request")
		return
	}

	err := release.Install(&a.Catalog, data.ServiceId, data.PlanId, serviceId, acceptsIncomplete)

	if err != nil {
		exists, existsErr := release.Exists(serviceId)

		if existsErr == nil && exists {
			respondWithJSON(w, http.StatusConflict, nil)
			return
		}

		respondWithServerError(w, err)
		return
	}

	if acceptsIncomplete {
		respondWithJSON(w, http.StatusAccepted, nil)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}

func (a *App) deleteInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceId := vars["serviceId"]
	acceptsIncomplete := strings.EqualFold(r.URL.Query().Get("accepts_incomplete"), "true")

	err := release.Delete(serviceId)

	if err != nil {
		respondWithServerError(w, err)
		return
	}

	if acceptsIncomplete {
		respondWithJSON(w, http.StatusAccepted, nil)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}

func (a *App) queryInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceId := vars["serviceId"]

	status, err := release.GetStatus(serviceId)

	if err != nil {
		exists, existsErr := release.Exists(serviceId)

		if existsErr == nil && !exists {
			respondWithJSON(w, http.StatusGone, nil)
			return
		}

		respondWithServerError(w, err)
		return
	}

	respondWithState := func(state string) {
		respondWithJSON(w, http.StatusOK, map[string]string{
			"state": state,
		})
	}

	if status.IsFailed {
		respondWithState("failed")
		return
	}

	if status.IsAvailable {
		respondWithState("succeeded")
		return
	}

	respondWithState("in progress")
}

func (a *App) bindInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceId := vars["serviceId"]
	bindingId := vars["bindingId"]

	type requestData struct {
		ServiceId string `json:"service_id"`
		PlanId    string `json:"plan_id"`
	}

	type credentialsWrapper struct {
		UserCredentials map[string]interface{} `json:"credentials"`
	}

	var data requestData
	decoder := json.NewDecoder(r.Body)
	decoderErr := decoder.Decode(&data)

	if decoderErr != nil || len(data.ServiceId) == 0 || len(data.PlanId) == 0 {
		respondWithUserError(w, "Invalid Request")
		return
	}

	credentials, err := release.GetCredentials(&a.Catalog, data.ServiceId, data.PlanId, serviceId)

	if err != nil {
		exists, existsErr := release.Exists(serviceId)

		if existsErr == nil && !exists {
			respondWithJSON(w, http.StatusConflict, nil)
			return
		}

		respondWithServerError(w, err)
		return
	}

	_ = bindingId

	respondWithJSON(w, http.StatusOK, credentialsWrapper{ UserCredentials: credentials })
}

func (a *App) unbindInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceId := vars["serviceId"]
	bindingId := vars["bindingId"]

	exists, err := release.Exists(serviceId)

	if err != nil {
		respondWithServerError(w, err)
		return
	}

	if !exists {
		respondWithJSON(w, http.StatusGone, nil)
		return
	}

	_ = bindingId

	respondWithJSON(w, http.StatusOK, nil)
}

func respondWithUserError(w http.ResponseWriter, description string) {
	respondWithJSONError(w, http.StatusBadRequest, "", description)
}

func respondWithServerError(w http.ResponseWriter, error error) {
	respondWithJSONError(w, http.StatusInternalServerError, "", error.Error())
}

func respondWithJSONError(w http.ResponseWriter, code int, error string, description string) {
	payload := map[string]string{}

	if len(error) > 0 {
		payload["error"] = error
	}

	if len(description) > 0 {
		payload["description"] = description
	}

	if len(payload) == 0 {
		payload = nil
	}

	respondWithJSON(w, code, payload)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if payload == nil {
		w.Write([]byte("{}"))
	} else {
		response, _ := json.Marshal(payload)
		w.Write(response)
	}
}
