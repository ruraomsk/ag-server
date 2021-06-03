package memDB

import (
	"fmt"
	"reflect"
	"sync"
)

type mDB struct {
	mu   sync.RWMutex
	Data map[string]interface{}
}
type Tx struct {
	name     string
	MDB      *mDB
	writable bool
	added    map[string]bool
	deleted  map[string]bool
	updated  map[string]bool
	ReadAll  func() map[string]interface{}
	WriteAll func() error
	AddFn    func(key string, value interface{}) string
	DeleteFn func(key string) string
	UpdateFn func(key string, value interface{}) string
}

func Create() *Tx {
	tx := Tx{writable: true, MDB: &mDB{Data: make(map[string]interface{})}, updated: make(map[string]bool), added: make(map[string]bool), deleted: make(map[string]bool)}
	return &tx
}

func (tx *Tx) Set(key string, value interface{}) {
	//tx.Lock()
	//defer tx.Unlock()
	if !tx.writable {
		return
	}
	oldvalue, is := tx.MDB.Data[key]
	if !is {
		tx.added[key] = true
		tx.MDB.Data[key] = value
	} else {
		if !reflect.DeepEqual(&oldvalue, &value) {
			tx.MDB.Data[key] = value
			tx.updated[key] = true
		}
	}
}
func (tx *Tx) GetAllKeys() []string {
	result := make([]string, 0)
	for key := range tx.MDB.Data {
		result = append(result, key)
	}
	return result
}

func (tx *Tx) Delete(key string) {
	if !tx.writable {
		return
	}
	delete(tx.MDB.Data, key)
	tx.deleted[key] = true
}

func (tx *Tx) Get(key string) (interface{}, error) {
	var err error
	_, is := tx.MDB.Data[key]
	if !is {
		err = fmt.Errorf("нет такого ключа %s", key)
	}
	return tx.MDB.Data[key], err
}

func (tx *Tx) Lock() {
	if tx.writable {
		tx.MDB.mu.Lock()
	} else {
		tx.MDB.mu.RLock()
	}
}

func (tx *Tx) Unlock() {
	if tx.writable {
		tx.MDB.mu.Unlock()
	} else {
		tx.MDB.mu.RUnlock()
	}
}
