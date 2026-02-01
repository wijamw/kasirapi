package main

import (
	"encoding/json"
	"fmt"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DBCONN"`
}

func main() {
	// viper module config
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DBCONN"),
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// declaration + routes
	catRepo := repositories.NewCategoryRepository(db)
	catService := services.NewCategoryService(catRepo)
	catHandler := handlers.NewCategoryHandler(catService)

	http.HandleFunc("/categories", catHandler.HandleCategories)
	http.HandleFunc("/categories/", catHandler.HandleCategoryByID)
	
	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"message": "API Running",
		})
	})

	fmt.Println("server running in localhost:" +config.Port)

	err = http.ListenAndServe(":" +config.Port, nil)
	if err != nil {
		fmt.Println("API fail to run. Contact your System Administrator")
	}
}
