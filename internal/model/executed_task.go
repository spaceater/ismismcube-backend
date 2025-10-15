package model

import (
	"ismismcube-backend/internal/config"
	"os"
	"strconv"
	"sync"
)

type ExecutedTaskModel struct {
	mutex sync.Mutex
}

var ExecutedTask *ExecutedTaskModel

func init() {
	ExecutedTask = &ExecutedTaskModel{}
}

func (etm *ExecutedTaskModel) Get() (int, error) {
	etm.mutex.Lock()
	defer etm.mutex.Unlock()
	executedTaskFile := config.ExecutedTaskFile
	data, err := os.ReadFile(executedTaskFile)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(executedTaskFile, []byte("0"), 0644); err != nil {
				return -1, err
			}
			return 0, nil
		}
		return -1, err
	}
	executedTask, err := strconv.Atoi(string(data))
	if err != nil {
		if err := os.WriteFile(executedTaskFile, []byte("0"), 0644); err != nil {
			return -1, err
		}
		return 0, nil
	}
	return executedTask, nil
}

func (etm *ExecutedTaskModel) GetAndIncrement() (int, error) {
	etm.mutex.Lock()
	defer etm.mutex.Unlock()
	executedTaskFile := config.ExecutedTaskFile
	data, err := os.ReadFile(executedTaskFile)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(executedTaskFile, []byte("1"), 0644); err != nil {
				return -1, err
			}
			return 1, nil
		}
		return -1, err
	}
	executedTask, err := strconv.Atoi(string(data))
	if err != nil {
		if err := os.WriteFile(executedTaskFile, []byte("1"), 0644); err != nil {
			return -1, err
		}
		return 1, nil
	}
	executedTask++
	if err := os.WriteFile(executedTaskFile, []byte(strconv.Itoa(executedTask)), 0644); err != nil {
		return -1, err
	}
	return executedTask, nil
}
