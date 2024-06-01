package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	ID            int64   `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
}

type TaskResult struct {
	ID     int64   `json:"id"`
	Result float64 `json:"result"`
}

func getTask() (*Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task, status code: %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func sendResult(taskResult TaskResult) error {
	data, err := json.Marshal(taskResult)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send result, status code: %d", resp.StatusCode)
	}

	return nil
}

func compute(task *Task) (float64, error) {
	switch task.Operation {
	case "add":
		return task.Arg1 + task.Arg2, nil
	case "sub":
		return task.Arg1 - task.Arg2, nil
	case "mul":
		return task.Arg1 * task.Arg2, nil
	case "div":
		if task.Arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, fmt.Errorf("unsupported operation: %s", task.Operation)
	}
}

func main() {
	for {
		task, err := getTask()
		if err != nil {
			fmt.Println("Error getting task:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Received task:", task)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		result, err := compute(task)
		if err != nil {
			fmt.Println("Error computing task:", err)
			continue
		}

		taskResult := TaskResult{
			ID:     task.ID,
			Result: result,
		}

		if err := sendResult(taskResult); err != nil {
			fmt.Println("Error sending result:", err)
		}
	}
}
