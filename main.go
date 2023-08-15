package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Task struct {
	ID     int
	Name   string `gorm:"not null"`
	Status bool   `gorm:"default:0"`
}

var dsn = "root:root@tcp(127.0.0.1:3306)/todoList?charset=utf8mb4&parseTime=True&loc=Local"
var db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func main() {

	err := db.AutoMigrate(&Task{})

	router := mux.NewRouter()
	router.HandleFunc("/", showTasks).Methods("GET")
	router.HandleFunc("/", createTask).Methods("POST")
	router.HandleFunc("/", deleteTask).Methods("DELETE")
	router.HandleFunc("/", updateTask).Methods("PUT")
	http.Handle("/", router)
	log.Println("Start listening 127.0.0.1:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func showTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	db.Find(&tasks)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	task := &Task{Name: name}
	db.Create(&task)

	db.Last(&task)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	fmt.Println(id)
	err := getTaskByID(id)
	if err == false {
		io.WriteString(w, `{"error": "Record Not Found"}`)
	} else {
		task := &Task{}
		db.First(&task, id)
		db.Delete(&task)
		io.WriteString(w, `{"deleted": true}`)
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	err := getTaskByID(id)
	if err == false {
		io.WriteString(w, `{"error": "Record Not Found"}`)
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var task Task
		err = json.Unmarshal(body, &task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		oldTask := &Task{}
		db.First(&oldTask, id)
		oldTask.Name = task.Name
		oldTask.Status = task.Status
		db.Save(&oldTask)

		io.WriteString(w, `{"updated": true}`)
	}
}

func getTaskByID(ID int) bool {
	task := &Task{}
	result := db.First(&task, ID)
	if result.Error != nil {
		log.Println("Not found in database")
		return false
	}
	return true
}
