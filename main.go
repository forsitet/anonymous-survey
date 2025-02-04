package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

type SurveyResponse struct {
	Rating   int    `json:"rating"`
	Feedback string `json:"feedback"`
}

var db *sql.DB

func initDB() {
	// user, ok := os.LookupEnv("USERBD")
	// if !ok {
	// 	log.Fatal("User not found")
	// }

	// password, ok := os.LookupEnv("PASSWORDBD")
	// if !ok {
	// 	log.Fatal("Password not found")
	// }
	// _ = password
	// _ = user
	connStr := os.Getenv("DATABASE_URL")
	// connStr := fmt.Sprintf("sslmode=disable user=%s password=%s", user, password)
	tempDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	tempDB.Exec("CREATE DATABASE survey;")
	tempDB.Close()

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS responses (
		id SERIAL PRIMARY KEY,
		rating INTEGER NOT NULL,
		feedback TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func surveyHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil {
		http.Error(w, "Некорректное значение рейтинга", http.StatusBadRequest)
		return
	}

	survey := SurveyResponse{
		Rating:   rating,
		Feedback: r.FormValue("feedback"),
	}

	_, err = db.Exec("INSERT INTO responses (rating, feedback) VALUES ($1, $2)", survey.Rating, survey.Feedback)
	if err != nil {
		http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(survey)
	if err != nil {
		http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
		return
	}

	log.Println("Ответ получен:", string(jsonData))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<p>Спасибо за твою обратную связь! Желаю успехов во всех начинаниях:) </p>"))
	w.Write([]byte("<img src='static/cap.jpg' alt='Капибара' width='300' height='200'>"))

}

func main() {
	initDB()
	http.HandleFunc("/", surveyHandler)
	http.HandleFunc("/submit", submitHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	log.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
