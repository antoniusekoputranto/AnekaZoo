package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Animal struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
	Legs  int    `json:"legs"`
}

// AnimalStore defines the interface for animal data operations.
// This abstraction makes it easier to switch between different storage implementations (e.g., in-memory, database).
type AnimalStore interface {
	GetAllAnimals() ([]Animal, error)
	GetAnimalByID(id int) (*Animal, error)
	CreateAnimal(animal Animal) error
	UpdateAnimal(id int, animal Animal) error // For PUT: updates if exists
	UpsertAnimal(id int, animal Animal) error // For PUT: creates if not exists, updates if exists
	DeleteAnimal(id int) error
}

// InMemoryAnimalStore implements AnimalStore using a map in memory.
type InMemoryAnimalStore struct {
	animals map[int]Animal // Stores animals by their ID
	mu      sync.Mutex     // Mutex to protect access to the animals map for thread safety
	nextID  int            // For auto-generating IDs if needed (though problem implies ID comes from payload)
}

// NewInMemoryAnimalStore creates and initializes a new InMemoryAnimalStore.
func NewInMemoryAnimalStore() *InMemoryAnimalStore {
	return &InMemoryAnimalStore{
		animals: make(map[int]Animal),
		nextID:  1, // Start ID from 1
	}
}

// GetAllAnimals retrieves all animals from the store.
func (s *InMemoryAnimalStore) GetAllAnimals() ([]Animal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.animals) == 0 {
		return nil, fmt.Errorf("no animals found") // Indicate no animals exist
	}

	var all []Animal
	for _, animal := range s.animals {
		all = append(all, animal)
	}
	return all, nil
}

// GetAnimalByID retrieves a single animal by its ID.
func (s *InMemoryAnimalStore) GetAnimalByID(id int) (*Animal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	animal, ok := s.animals[id]
	if !ok {
		return nil, fmt.Errorf("animal with ID %d not found", id)
	}
	return &animal, nil
}

// CreateAnimal adds a new animal to the store.
// Returns an error if an animal with the same ID already exists.
func (s *InMemoryAnimalStore) CreateAnimal(animal Animal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if animal.ID == 0 {
		// If ID is not provided (0 value), generate one.
		// NOTE: The problem statement implies ID is usually provided in the payload for POST.
		// This is a fallback for robustness.
		animal.ID = s.nextID
		s.nextID++
	} else if _, exists := s.animals[animal.ID]; exists {
		return fmt.Errorf("animal with ID %d already exists", animal.ID)
	}

	s.animals[animal.ID] = animal
	return nil
}

// UpdateAnimal updates an existing animal in the store.
// Returns an error if the animal with the specified ID does not exist.
func (s *InMemoryAnimalStore) UpdateAnimal(id int, animal Animal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.animals[id]; !exists {
		return fmt.Errorf("animal with ID %d not found for update", id)
	}
	// Ensure the ID in the payload matches the path ID
	animal.ID = id
	s.animals[id] = animal
	return nil
}

// UpsertAnimal updates an existing animal or creates a new one if it doesn't exist.
func (s *InMemoryAnimalStore) UpsertAnimal(id int, animal Animal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	animal.ID = id // Ensure the ID from the path is used
	s.animals[id] = animal
	return nil
}

// DeleteAnimal removes an animal from the store by its ID.
// Returns an error if the animal with the specified ID does not exist.
func (s *InMemoryAnimalStore) DeleteAnimal(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.animals[id]; !exists {
		return fmt.Errorf("animal with ID %d not found for deletion", id)
	}
	delete(s.animals, id)
	return nil
}

// --- HTTP Handlers ---

// getAnimalsHandler handles GET requests for all animals.
func getAnimalsHandler(store AnimalStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		animals, err := store.GetAllAnimals()
		if err != nil {
			// If no animals found, return 404 Not Found as per problem statement
			if err.Error() == "no animals found" {
				http.Error(w, "No animals found in the system", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(animals)
	}
}

// getAnimalHandler handles GET requests for a single animal by ID.
func getAnimalHandler(store AnimalStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid animal ID", http.StatusBadRequest)
			return
		}

		animal, err := store.GetAnimalByID(id)
		if err != nil {
			// If animal not found, return 404 Not Found
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(animal)
	}
}

// createAnimalHandler handles POST requests to create a new animal.
// Denies duplicate entries based on ID.
func createAnimalHandler(store AnimalStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var animal Animal
		if err := json.NewDecoder(r.Body).Decode(&animal); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Ensure ID is provided and valid for creation
		if animal.ID == 0 {
			http.Error(w, "Animal ID is required for creation", http.StatusBadRequest)
			return
		}

		// Check if animal with this ID already exists to deny duplicate entry
		_, err := store.GetAnimalByID(animal.ID)
		if err == nil { // No error means animal found
			http.Error(w, fmt.Sprintf("Animal with ID %d already exists", animal.ID), http.StatusConflict) // 409 Conflict
			return
		}

		// Attempt to create the animal
		if err := store.CreateAnimal(animal); err != nil {
			// Specific check for "already exists" error from store is redundant here due to prior check,
			// but good for other potential store errors.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(animal)
	}
}

// updateAnimalHandler handles PUT requests to update an existing animal or create a new one (upsert).
func updateAnimalHandler(store AnimalStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid animal ID in path", http.StatusBadRequest)
			return
		}

		var animal Animal
		if err := json.NewDecoder(r.Body).Decode(&animal); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Ensure the ID from the path is used for the operation, ignoring ID in body if different
		animal.ID = id

		// Check if the animal exists to determine if it's an update or create
		_, existsErr := store.GetAnimalByID(id)

		if existsErr == nil {
			// Animal exists, perform update
			if err := store.UpdateAnimal(id, animal); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK) // 200 OK for update
			json.NewEncoder(w).Encode(animal)
		} else {
			// Animal does not exist, perform creation (upsert)
			if err := store.UpsertAnimal(id, animal); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated) // 201 Created for new resource
			json.NewEncoder(w).Encode(animal)
		}
	}
}

// deleteAnimalHandler handles DELETE requests to delete an animal by ID.
func deleteAnimalHandler(store AnimalStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, "Invalid animal ID", http.StatusBadRequest)
			return
		}

		if err := store.DeleteAnimal(id); err != nil {
			// If animal not found for deletion, return 404 Not Found
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
	}
}

func main() {
	// Initialize the in-memory animal store
	animalStore := NewInMemoryAnimalStore()

	// Add some initial dummy data
	_ = animalStore.CreateAnimal(Animal{ID: 1, Name: "lion", Class: "mammal", Legs: 4})
	_ = animalStore.CreateAnimal(Animal{ID: 2, Name: "eagle", Class: "bird", Legs: 2})
	_ = animalStore.CreateAnimal(Animal{ID: 3, Name: "snake", Class: "reptile", Legs: 0})

	r := mux.NewRouter()

	// Define API routes with a /v1/animals prefix
	r.HandleFunc("/v1/animals", getAnimalsHandler(animalStore)).Methods("GET")
	r.HandleFunc("/v1/animals/{id}", getAnimalHandler(animalStore)).Methods("GET")
	r.HandleFunc("/v1/animals", createAnimalHandler(animalStore)).Methods("POST")
	r.HandleFunc("/v1/animals/{id}", updateAnimalHandler(animalStore)).Methods("PUT")
	r.HandleFunc("/v1/animals/{id}", deleteAnimalHandler(animalStore)).Methods("DELETE")

	fmt.Print("Starting server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", r))
}
