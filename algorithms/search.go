package algorithms

import "employee-management-web/models"

func LinearSearch(employees []models.Employee, query string) []models.Employee {
	result := make([]models.Employee, 0)
	for _, e := range employees {
		if contains(e.Name, query) || contains(e.Email, query) || contains(e.Department, query) || contains(e.Position, query) {
			result = append(result, e)
		}
	}
	return result
}

func contains(value, query string) bool {
	if query == "" {
		return true
	}
	return len(value) >= len(query) && indexOfFold(value, query)
}

func indexOfFold(value, query string) bool {
	v, q := []rune(value), []rune(query)
	for i := range v {
		if i+len(q) > len(v) {
			break
		}
		match := true
		for j := range q {
			if lower(v[i+j]) != lower(q[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
func lower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
