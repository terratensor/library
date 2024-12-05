package workerpool

import (
	"fmt"
)

type Task struct {
	Err  error
	Data interface{}
	f    func(interface{}) error
}

func NewTask(f func(interface{}) error, data interface{}) *Task {
	return &Task{f: f, Data: data}
}

func process(workerID int, task *Task) {
	fmt.Printf("Worker %d processes task %v\n", workerID, task.Data)
	task.Err = task.f(task.Data)
	if task.Err != nil {
		appendErrorToLog("./error_log.txt", task.Err)
	}
}

func appendErrorToLog(logFilePath string, errStr error) {
	// TODO записываем ошибку в файл
}
