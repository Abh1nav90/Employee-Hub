package patterns

import "employee-management-web/models"

func NewEmployee(id int, name, email, phone, department, position string, salary float64, status string) models.Employee {
	return models.Employee{ID: id, Name: name, Email: email, Phone: phone, Department: department, Position: position, Salary: salary, Status: status}
}
