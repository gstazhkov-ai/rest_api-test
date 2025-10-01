package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// Task - структура нашей задачи
type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

// Хранилище для наших задач (замена базы данных)
var (
	tasks  = make(map[int]Task)
	nextID = 1
	mutex  = &sync.Mutex{}
)

// tasksHandler обрабатывает запросы к /tasks (GET all, POST)
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasks(w, r)
	case http.MethodPost:
		createTask(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// taskHandler обрабатывает запросы к /tasks/{id} (GET one, PUT, DELETE)
// --- ИЗМЕНЕНА ЭТА ФУНКЦИЯ ---
func taskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	// В зависимости от HTTP метода вызываем нужную функцию
	switch r.Method {
	case http.MethodGet:
		getTask(w, r, id)
	case http.MethodPut:
		updateTask(w, r, id)
	case http.MethodDelete:
		deleteTask(w, r, id)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// getTasks возвращает все задачи
func getTasks(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	taskList := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		taskList = append(taskList, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskList)
}

// createTask создает новую задачу
func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	task.ID = nextID
	tasks[task.ID] = task
	nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// getTask возвращает одну задачу по ID
func getTask(w http.ResponseWriter, r *http.Request, id int) {
	mutex.Lock()
	task, found := tasks[id]
	mutex.Unlock()

	if !found {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// --- НОВАЯ ФУНКЦИЯ ---
// updateTask обновляет существующую задачу
func updateTask(w http.ResponseWriter, r *http.Request, id int) {
	mutex.Lock()
	_, found := tasks[id]
	mutex.Unlock()

	if !found {
		http.Error(w, "Задача для обновления не найдена", http.StatusNotFound)
		return
	}

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Убедимся, что ID в URL и в теле запроса совпадает
	updatedTask.ID = id
	tasks[id] = updatedTask

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// --- НОВАЯ ФУНКЦИЯ ---
// deleteTask удаляет задачу
func deleteTask(w http.ResponseWriter, r *http.Request, id int) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, found := tasks[id]; !found {
		http.Error(w, "Задача для удаления не найдена", http.StatusNotFound)
		return
	}

	delete(tasks, id)
	w.WriteHeader(http.StatusNoContent) // Стандартный ответ для успешного удаления без тела
}

func main() {
	// Создаем несколько задач для старта
	tasks[0] = Task{ID: 0, Name: "Выучить Go", Done: false}
	
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/", taskHandler)

	println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
