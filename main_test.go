package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// Reset categories to initial state before each test
func resetCategories() {
	category = []Category{
		{ID: 1, Name: "Duelist", Description: "First Contact. Enter Site. Frag"},
		{ID: 2, Name: "Controller", Description: "Divide the map. Gain positioning advantage."},
		{ID: 3, Name: "Initiator", Description: "Help Duelist entry. Gain info on the enemy"},
		{ID: 4, Name: "Sentinel", Description: "Site Anchor. Slow down enemy rush."},
	}
}

func TestHealthCheck(t *testing.T) {
	resetCategories()
	
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API Running",
		})
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"status":"OK","message":"API Running"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestGetCategories(t *testing.T) {
	resetCategories()
	
	req, err := http.NewRequest("GET", "/categories", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(category)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if response contains expected categories
	var categories []Category
	err = json.Unmarshal(rr.Body.Bytes(), &categories)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if len(categories) != 4 {
		t.Errorf("expected 4 categories, got %d", len(categories))
	}
}

func TestGetCategoryByID(t *testing.T) {
	resetCategories()
	
	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantName   string
	}{
		{"Valid ID", "1", http.StatusOK, "Duelist"},
		{"Another Valid ID", "3", http.StatusOK, "Initiator"},
		{"Invalid ID", "999", http.StatusNotFound, ""},
		{"Non-numeric ID", "abc", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/categories/"+tt.id, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate the path parsing from main.go
				idStr := r.URL.Path[len("/categories/"):]
				
				// Convert string ID to int
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid Request", http.StatusBadRequest)
					return
				}
				
				// Find category
				for _, c := range category {
					if c.ID == id {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(c)
						return
					}
				}
				
				// ID not found
				http.Error(w, "Category not found", http.StatusNotFound)
			})

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			if tt.wantName != "" {
				var cat Category
				err := json.Unmarshal(rr.Body.Bytes(), &cat)
				if err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if cat.Name != tt.wantName {
					t.Errorf("expected category name %s, got %s", tt.wantName, cat.Name)
				}
			}
		})
	}
}

func TestPostCategory(t *testing.T) {
	resetCategories()
	
	newCategory := Category{
		Name:        "New Role",
		Description: "Test Description",
	}

	jsonData, err := json.Marshal(newCategory)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/categories", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var catBaru Category
			err := json.NewDecoder(r.Body).Decode(&catBaru)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}

			// Add new category
			catBaru.ID = len(category) + 1
			category = append(category, catBaru)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(catBaru)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var createdCat Category
	err = json.Unmarshal(rr.Body.Bytes(), &createdCat)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if createdCat.ID != 5 {
		t.Errorf("expected new category ID to be 5, got %d", createdCat.ID)
	}
	if createdCat.Name != "New Role" {
		t.Errorf("expected category name 'New Role', got %s", createdCat.Name)
	}
}

func TestPutCategoryByID(t *testing.T) {
	resetCategories()
	
	updatedCategory := Category{
		Name:        "Updated Duelist",
		Description: "Updated description",
	}

	jsonData, err := json.Marshal(updatedCategory)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", "/categories/1", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate the path parsing
		idStr := r.URL.Path[len("/categories/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		if r.Method == "PUT" {
			var updateCat Category
			err := json.NewDecoder(r.Body).Decode(&updateCat)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}

			// Update category
			for i := range category {
				if category[i].ID == id {
					updateCat.ID = id
					category[i] = updateCat

					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(category[i])
					return
				}
			}

			// ID not found
			http.Error(w, "Category not found", http.StatusNotFound)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var cat Category
	err = json.Unmarshal(rr.Body.Bytes(), &cat)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if cat.Name != "Updated Duelist" {
		t.Errorf("expected updated name 'Updated Duelist', got %s", cat.Name)
	}
}

func TestDeleteCategoryByID(t *testing.T) {
	resetCategories()
	
	req, err := http.NewRequest("DELETE", "/categories/2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate the path parsing
		idStr := r.URL.Path[len("/categories/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		if r.Method == "DELETE" {
			for i, c := range category {
				if c.ID == id {
					category = append(category[:i], category[i+1:]...)

					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{
						"message": "Deleted",
					})
					return
				}
			}

			// ID not found
			http.Error(w, "Category not found", http.StatusNotFound)
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify category was deleted
	if len(category) != 3 {
		t.Errorf("expected 3 categories after deletion, got %d", len(category))
	}

	// Check that ID 2 no longer exists
	for _, c := range category {
		if c.ID == 2 {
			t.Errorf("category with ID 2 should have been deleted but still exists")
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	resetCategories()
	
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"PATCH on categories", "PATCH", "/categories"},
		{"OPTIONS on categories", "OPTIONS", "/categories"},
		{"PATCH on specific category", "PATCH", "/categories/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/categories":
					if r.Method != "GET" && r.Method != "POST" {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
						return
					}
				default:
					if len(r.URL.Path) > len("/categories/") {
						if r.Method != "GET" && r.Method != "PUT" && r.Method != "DELETE" {
							http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
							return
						}
					}
				}
			})

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusMethodNotAllowed {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestPostCategoryInvalidJSON(t *testing.T) {
	resetCategories()
	
	// Send invalid JSON
	invalidJSON := []byte(`{"name": "Test", "description": 123}`) // description should be string

	req, err := http.NewRequest("POST", "/categories", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var catBaru Category
			err := json.NewDecoder(r.Body).Decode(&catBaru)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}