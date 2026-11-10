package services

import (
	"employee-management-web/algorithms"
	"employee-management-web/models"
	"employee-management-web/patterns"
	"errors"
	"sort"
	"strings"
	"time"
)

type EmployeeService struct{ Store *patterns.Store }

func (s *EmployeeService) List(query, department, status, sortDir string) []models.Employee {
	emps := s.Store.Snapshot()
	if query != "" {
		emps = algorithms.LinearSearch(emps, query)
	}
	out := emps[:0]
	for _, e := range emps {
		if department != "" && !strings.EqualFold(e.Department, department) {
			continue
		}
		if status != "" && !strings.EqualFold(e.Status, status) {
			continue
		}
		out = append(out, e)
	}
	if sortDir == "asc" {
		return algorithms.SortByName(out, true)
	}
	if sortDir == "desc" {
		return algorithms.SortByName(out, false)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *EmployeeService) Add(e models.Employee) (models.Employee, error) {
	if strings.TrimSpace(e.Name) == "" || strings.TrimSpace(e.Email) == "" {
		return e, errors.New("name and email are required")
	}
	s.Store.Employees = s.Store.Snapshot()
	max := 0
	for _, x := range s.Store.Employees {
		if x.ID > max {
			max = x.ID
		}
	}
	e.ID = max + 1
	if e.Status == "" {
		e.Status = "Active"
	}
	if e.JoinDate.IsZero() {
		e.JoinDate = time.Now()
	}
	s.Store.Employees = append(s.Store.Employees, e)
	return e, s.Store.Save()
}
func (s *EmployeeService) Update(id int, e models.Employee) (models.Employee, error) {
	s.Store.Employees = s.Store.Snapshot()
	for i := range s.Store.Employees {
		if s.Store.Employees[i].ID == id {
			e.ID = id
			s.Store.Employees[i] = e
			return e, s.Store.Save()
		}
	}
	return e, errors.New("employee not found")
}
func (s *EmployeeService) Delete(id int) error {
	s.Store.Employees = s.Store.Snapshot()
	for i, e := range s.Store.Employees {
		if e.ID == id {
			s.Store.Employees = append(s.Store.Employees[:i], s.Store.Employees[i+1:]...)
			return s.Store.Save()
		}
	}
	return errors.New("employee not found")
}
