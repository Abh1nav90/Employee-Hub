package algorithms

import "employee-management-web/models"

func SortByName(employees []models.Employee, ascending bool) []models.Employee {
	out := append([]models.Employee(nil), employees...)
	for i := 0; i < len(out); i++ {
		for j := 0; j < len(out)-i-1; j++ {
			if shouldSwap(out[j].Name, out[j+1].Name, ascending) {
				out[j], out[j+1] = out[j+1], out[j]
			}
		}
	}
	return out
}

func shouldSwap(a, b string, asc bool) bool {
	if asc {
		return a > b
	}
	return a < b
}
