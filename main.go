package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)
type Category struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
}

var category = []Category{
	{ID: 1, Name: "Duelist", Description: "First Contact. Enter Site. Frag"},
	{ID: 2, Name: "Controller", Description: "Divide the map. Gain positioning advantage."},
	{ID: 3, Name: "Initiator", Description: "Help Duelist entry. Gain info on the enemy"},
	{ID: 4, Name: "Sentinel", Description: "Site Anchor. Slow down enemy rush."},
}

func getCategorybyID(w http.ResponseWriter, id int){
	// // URL: /categories/1 -> ID = 1
	// idStr := strings.TrimPrefix(r.URL.Path, "/categories/")

	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	http.Error(w, "Invalid Request", http.StatusBadRequest)
	// 	return
	// }

	// Cari kategori
	for _, c := range category {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}
	}

	// ID not found
	http.Error(w, "Category not found", http.StatusNotFound)

}

func putCategorybyID(w http.ResponseWriter, r *http.Request, id int) {
	// URL: /categories/1 -> ID = 1
	// idStr := strings.TrimPrefix(r.URL.Path, "/categories/")

	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	http.Error(w, "Invalid Request", http.StatusBadRequest)
	// 	return
	// }

	var updateCat Category
	err := json.NewDecoder(r.Body).Decode(&updateCat)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	// replace
	for i:= range category {
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

func delCategorybyID(w http.ResponseWriter, r *http.Request, id int) {
	for i, c := range category {
		if c.ID == id {
			category = append(category[:i], category[i+1:]... )

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Deleted",
			})
			return
		}
	}

}

func main() {
	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"message": "API Running",
		})
	})

	// GET, PUT, DELETE
	// http://localhost:8080/categories/{id}
	http.HandleFunc("/categories/", func(w http.ResponseWriter, r *http.Request) {
		// URL: /categories/1 -> ID = 1
		idStr := strings.TrimPrefix(r.URL.Path, "/categories/")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			getCategorybyID(w, id)
		
		case "PUT":
			putCategorybyID(w, r, id)

		case "DELETE":
			delCategorybyID(w, r, id)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET, POST /categories
	http.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(category)

		case "POST":
			var catBaru Category
			// Decode JSON dan menyimpan di catBaru dan menyimpan nilai error dalam err
			err := json.NewDecoder(r.Body).Decode(&catBaru)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}

			// Masukkan data
			catBaru.ID = len(category) + 1
			category = append(category, catBaru)

			// Return Hasil
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated) //Return 201
			json.NewEncoder(w).Encode(catBaru)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("server running in localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("API fail to run. Contact your System Administrator")
	}
}
