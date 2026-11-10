package models

import "time"

type Employee struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Department     string    `json:"department"`
	Position       string    `json:"position"`
	Salary         float64   `json:"salary"`
	Status         string    `json:"status"`
	EmploymentType string    `json:"employmentType"`
	Location       string    `json:"location"`
	JoinDate       time.Time `json:"joinDate"`
}
