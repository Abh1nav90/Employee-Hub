package patterns

import (
	"employee-management-web/models"
	"encoding/json"
	"os"
	"sync"
)

type Store struct {
	mu        sync.RWMutex
	File      string
	Employees []models.Employee
}

func NewStore(file string) *Store { return &Store{File: file, Employees: []models.Employee{}} }
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.File)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.Employees)
}
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s.Employees, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.File, data, 0644)
}
func (s *Store) Snapshot() []models.Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Employee(nil), s.Employees...)
}
