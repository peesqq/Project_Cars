package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "hghghghghghg/docs"
	"log"
	"net/http"
	"os"
)

type Car struct {
	ID     uint   `gorm:"primaryKey"`
	RegNum string `gorm:"unique_index" json:"regNum"`
	Mark   string `json:"mark"`
	Models string `json:"model"`
	Year   int    `json:"year"`
	Owner  uint   `json:"owner"`
}

type People struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type CarNumbers struct {
	Numbers []string `json:"regNums"`
}

var db *gorm.DB

//	@title			Car Management API
//	@description	API для управления автомобилями и владельцами
//	@version		2.0
//	@host			localhost:8080
//	@BasePath		/

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	r := mux.NewRouter()
	registerHandler(r)

	db, err = connectDB()

	errMigrate := db.AutoMigrate(&People{}, &Car{})
	if errMigrate != nil {
		log.Fatalf("Error database migrate: %v", errMigrate)
	}

	port := os.Getenv("PORT")
	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func connectDB() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", dbHost, dbPort, dbUser, dbName, dbPassword)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func registerHandler(r *mux.Router) {
	r.HandleFunc("/cars", getAllCars).Methods("GET")
	r.HandleFunc("/cars", createCar).Methods("POST")
	r.HandleFunc("/cars/{id}", updateCar).Methods("PUT")
	r.HandleFunc("/cars/{id}", getCar).Methods("GET")
	r.HandleFunc("/cars/{id}", deleteCar).Methods("DELETE")
	r.HandleFunc("/info", getInfo).Methods("GET")
	r.HandleFunc("/carsbynum", createCarsByNums).Methods("POST")

	// Добавляем обработчик Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("doc.json"), // URL для получения JSON-спецификации
	))
}

// @Summary Получить все автомобили
// @Description Получить список всех автомобилей
// @Tags Cars
// @Accept  json
// @Produce  json
// @Success 200 {array} Car
// @Router /cars [get]
func getAllCars(w http.ResponseWriter, _ *http.Request) {
	var cars []Car
	db.Find(&cars)
	err := json.NewEncoder(w).Encode(cars)
	if err != nil {
		log.Fatalf("Error get all cars: %v", err)
	}
}

// @Summary Получить информацию о машине
// @Description Получить информацию о машине по её ID
// @Tags Cars
// @Accept  json
// @Produce  json
// @Param id path int true "ID машины"
// @Success 200 {object} Car
// @Router /cars/{id} [get]
func getCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var car Car
	db.First(&car, params["id"])
	err := json.NewEncoder(w).Encode(&car)
	if err != nil {
		log.Fatalf("Error get all cars: %v", err)
	}
}

// @Summary Создать новую машину
// @Description Создать новую машину
// @Tags Cars
// @Accept  json
// @Produce  json
// @Param car body Car true "Данные машины"
// @Success 200 {object} Car
// @Router /cars [post]
func createCar(_ http.ResponseWriter, r *http.Request) {
	var car Car
	err := json.NewDecoder(r.Body).Decode(&car)
	db.Create(&car)
	if err != nil {
		log.Fatalf("Error creating car: %v", err)
	}
}

// @Summary Создать автомобили по массиву номеров
// @Description Создает новые записи об автомобилях на основе переданных регистрационных номеров
// @Tags Сars
// @Accept  json
// @Produce  json
// @Param carNumbers body CarNumbers true "Список регистрационных номеров автомобилей в двойных кавычках через запятую"
// @Success 200 {string} string "Успешно создано"
// @Router /carsbynum [post]
func createCarsByNums(_ http.ResponseWriter, r *http.Request) {
	var carNumbers CarNumbers
	fmt.Println("111")
	err := json.NewDecoder(r.Body).Decode(&carNumbers)
	if err != nil {
		log.Fatalf("Error creating car: %v", err)
	}
	fmt.Println(carNumbers)
	for _, num := range carNumbers.Numbers {
		// Создаем автомобиль на основе номера
		car := Car{RegNum: num}
		fmt.Println(num)
		db.Create(&car)

	}

}

// @Summary Обновить информацию о машине
// @Description Обновить информацию о машине по её ID
// @Tags Cars
// @Accept  json
// @Produce  json
// @Param id path int true "ID машины"
// @Param car body Car true "Новые данные машины, нужно указать тот же ID машины при изменение"
// @Success 200 {object} Car
// @Router /cars/{id} [put]
func updateCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var car Car
	db.First(&car, params["id"])

	// Декодируем данные из запроса и обновляем только необходимые поля
	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Save(&car)

	// Возвращаем обновленный автомобиль
	err = json.NewEncoder(w).Encode(car)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// @Summary Удалить машину
// @Description Удалить машину по её ID
// @Tags Cars
// @Accept  json
// @Produce  json
// @Param id path int true "ID машины"
// @Success 200
// @Router /cars/{id} [delete]
func deleteCar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	db.Delete(&Car{}, id)
	getAllCars(w, r)
}

func getInfo(_ http.ResponseWriter, r *http.Request) {
	return
}
