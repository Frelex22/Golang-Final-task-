package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Expression struct {
	ID     string  `json:"id"`
	Expr   string  `json:"expression"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

var expressions = make(map[string]Expression)
var tasks = make(chan Task, 100)
var mu sync.Mutex

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/calculate", calculateHandler).Methods("POST")
	router.HandleFunc("/api/v1/expressions", listExpressionsHandler).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", getExpressionHandler).Methods("GET")
	router.HandleFunc("/internal/task", getTaskHandler).Methods("GET")
	router.HandleFunc("/internal/task", postResultHandler).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expr := req["expression"]
	id := strconv.Itoa(rand.Int())
	mu.Lock()
	expressions[id] = Expression{ID: id, Expr: expr, Status: "pending"}
	mu.Unlock()

	tasks <- Task{ID: id, Expr: expr}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func listExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var exprList []Expression
	for _, expr := range expressions {
		exprList = append(exprList, expr)
	}

	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": exprList})
}

func getExpressionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.Lock()
	expr, exists := expressions[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-tasks:
		json.NewEncoder(w).Encode(map[string]Task{"task": task})
	default:
		http.Error(w, "No tasks available", http.StatusNotFound)
	}
}

func postResultHandler(w http.ResponseWriter, r *http.Request) {
	var res TaskResult
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	expr := expressions[res.ID]
	expr.Status = "completed"
	expr.Result = res.Result
	expressions[res.ID] = expr
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

type Task struct {
	ID   string `json:"id"`
	Expr string `json:"expression"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}
