package model

import (
	"ismismcube-backend/internal/config"
	"os"
	"strconv"
	"sync"
)

type PageViewModel struct {
	mutex sync.Mutex
}

var PageView *PageViewModel

func init() {
	PageView = &PageViewModel{}
}

func (pvm *PageViewModel) Get() (int, error) {
	pvm.mutex.Lock()
	defer pvm.mutex.Unlock()
	pageViewFile := config.PageViewFile
	data, err := os.ReadFile(pageViewFile)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(pageViewFile, []byte("0"), 0644); err != nil {
				return -1, err
			}
			return 0, nil
		}
		return -1, err
	}
	pageView, err := strconv.Atoi(string(data))
	if err != nil {
		if err := os.WriteFile(pageViewFile, []byte("0"), 0644); err != nil {
			return -1, err
		}
		return 0, nil
	}
	return pageView, nil
}

func (pvm *PageViewModel) GetAndIncrement() (int, error) {
	pvm.mutex.Lock()
	defer pvm.mutex.Unlock()
	pageViewFile := config.PageViewFile
	data, err := os.ReadFile(pageViewFile)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(pageViewFile, []byte("1"), 0644); err != nil {
				return -1, err
			}
			return 1, nil
		}
		return -1, err
	}
	pageView, err := strconv.Atoi(string(data))
	if err != nil {
		if err := os.WriteFile(pageViewFile, []byte("1"), 0644); err != nil {
			return -1, err
		}
		return 1, nil
	}
	pageView++
	if err := os.WriteFile(pageViewFile, []byte(strconv.Itoa(pageView)), 0644); err != nil {
		return -1, err
	}
	return pageView, nil
}
