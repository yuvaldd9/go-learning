package api

import (
	"encoding/json"
	"go-learning/internal/core"
	"go-learning/internal/model"
	"go-learning/internal/storage"
	"net/http"
	"strconv"
	"sync"
)

const FeaturesLength = 256 // Define a constant for features length

type APIInterface interface {
	Start(port int)
}

type API struct {
	personStorage *storage.MemoryStore
}

// AddPersonRequest represents the request payload for adding a person.
type AddPersonRequest struct {
	Name     string    `json:"name"`
	Features [256]float64 `json:"features"`
}

// GetSimilarPersonResponse represents the response payload for similar persons.
type GetSimilarPersonResponse struct {
	Persons []core.MatchResult `json:"persons"`
}

func NewAPI(storage *storage.MemoryStore) *API {
	return &API{
		personStorage: storage,
	}
}

func (api *API) addPersonHandler(w http.ResponseWriter, r *http.Request) {
	var req AddPersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	person := model.Person{
		Name:     req.Name,
		Features: req.Features,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var addPersonErr error
	go func() {
		defer wg.Done()
		addPersonErr = api.personStorage.AddPerson(person)
	}()

	wg.Wait()

	if addPersonErr != nil {
		http.Error(w, "Failed to add person: "+addPersonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (api *API) getSimilarPersonHandler(w http.ResponseWriter, r *http.Request) {
	features := r.URL.Query()["features"]
	topNStr := r.URL.Query().Get("top_n")
	if topNStr == "" {
		topNStr = "3"
	}

	topN, err := strconv.Atoi(topNStr)
	if err != nil || topN <= 0 {
		http.Error(w, "Invalid top_n value: must be a positive integer", http.StatusBadRequest)
		return
	}

	var personFeatures [256]float64
	for index, f := range features {
		val, err := strconv.ParseFloat(f, 64)
		if err != nil {
			http.Error(w, "Invalid features value: "+err.Error(), http.StatusBadRequest)
			return
		}
		personFeatures[index] = val
	}

	if len(personFeatures) != FeaturesLength {
		http.Error(w, "Features must be an array of size 256", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var matches []core.MatchResult
	go func() {
		defer wg.Done()
		matches = core.GetSimilarPersons(personFeatures, api.personStorage, topN)
	}()

	wg.Wait()

	response := GetSimilarPersonResponse{
		Persons: matches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) Start(port int) {
	http.HandleFunc("/add_person", api.addPersonHandler)
	http.HandleFunc("/get_similar_person", api.getSimilarPersonHandler)

	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
