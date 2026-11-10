package main

import (
	"employee-management-web/models"
	"employee-management-web/patterns"
	"employee-management-web/services"
	"employee-management-web/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JSONCollection struct {
	mu   sync.RWMutex
	File string
	Data []map[string]any
}

func NewCollection(file string) *JSONCollection {
	return &JSONCollection{File: file, Data: []map[string]any{}}
}
func (c *JSONCollection) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, e := os.ReadFile(c.File)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	return json.Unmarshal(b, &c.Data)
}
func (c *JSONCollection) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	b, e := json.MarshalIndent(c.Data, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(c.File, b, 0644)
}
func (c *JSONCollection) All() []map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]map[string]any, len(c.Data))
	for i, v := range c.Data {
		out[i] = clone(v)
	}
	return out
}
func clone(m map[string]any) map[string]any {
	n := map[string]any{}
	for k, v := range m {
		n[k] = v
	}
	return n
}
func (c *JSONCollection) Add(v map[string]any) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	v = clone(v)
	v["id"] = nextID(c.Data)
	v["createdAt"] = time.Now().Format(time.RFC3339)
	c.Data = append(c.Data, v)
	_ = c.saveLocked()
	return clone(v)
}
func (c *JSONCollection) Update(id int, v map[string]any) (map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, x := range c.Data {
		if toInt(x["id"]) == id {
			v = clone(v)
			v["id"] = id
			if _, ok := v["createdAt"]; !ok {
				v["createdAt"] = x["createdAt"]
			}
			c.Data[i] = v
			_ = c.saveLocked()
			return clone(v), true
		}
	}
	return nil, false
}
func (c *JSONCollection) Delete(id int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, x := range c.Data {
		if toInt(x["id"]) == id {
			c.Data = append(c.Data[:i], c.Data[i+1:]...)
			_ = c.saveLocked()
			return true
		}
	}
	return false
}
func (c *JSONCollection) saveLocked() error {
	b, e := json.MarshalIndent(c.Data, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(c.File, b, 0644)
}
func nextID(a []map[string]any) int {
	max := 0
	for _, v := range a {
		if n := toInt(v["id"]); n > max {
			max = n
		}
	}
	return max + 1
}
func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

func main() {
	_ = os.MkdirAll("data", 0755)
	store := patterns.NewStore("data/employees.json")
	if err := store.Load(); err != nil {
		log.Fatal(err)
	}
	if len(store.Employees) == 0 {
		seed(store)
	}
	svc := &services.EmployeeService{Store: store}
	cols := map[string]*JSONCollection{}
	for _, name := range []string{"attendance", "leaves", "payroll", "performance", "tickets", "documents", "events", "onboarding", "notifications"} {
		c := NewCollection(filepath.Join("data", name+".json"))
		if err := c.Load(); err != nil {
			log.Fatal(err)
		}
		cols[name] = c
	}
	seedHRCollections(cols)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			utils.Error(w, 405, "method not allowed")
			return
		}
		var in struct{ Email, Password, Role string }
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			utils.Error(w, 400, "invalid JSON")
			return
		}
		role := in.Role
		if role == "" {
			role = "Admin"
		}
		utils.JSON(w, 200, map[string]any{"token": "local-demo-session", "user": map[string]any{"name": "Admin User", "email": in.Email, "role": role}})
	})
	mux.HandleFunc("/api/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			utils.JSON(w, 200, svc.List(r.URL.Query().Get("q"), r.URL.Query().Get("department"), r.URL.Query().Get("status"), r.URL.Query().Get("sort")))
		case "POST":
			var e models.Employee
			if json.NewDecoder(r.Body).Decode(&e) != nil {
				utils.Error(w, 400, "invalid JSON")
				return
			}
			x, er := svc.Add(e)
			if er != nil {
				utils.Error(w, 400, er.Error())
				return
			}
			utils.JSON(w, 201, x)
		default:
			utils.Error(w, 405, "method not allowed")
		}
	})
	mux.HandleFunc("/api/employees/", func(w http.ResponseWriter, r *http.Request) {
		id, e := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/employees/"), "/"))
		if e != nil {
			utils.Error(w, 400, "invalid employee id")
			return
		}
		switch r.Method {
		case "GET":
			for _, x := range svc.List("", "", "", "") {
				if x.ID == id {
					utils.JSON(w, 200, x)
					return
				}
			}
			utils.Error(w, 404, "employee not found")
		case "PUT":
			var x models.Employee
			if json.NewDecoder(r.Body).Decode(&x) != nil {
				utils.Error(w, 400, "invalid JSON")
				return
			}
			y, er := svc.Update(id, x)
			if er != nil {
				utils.Error(w, 404, er.Error())
				return
			}
			utils.JSON(w, 200, y)
		case "DELETE":
			if er := svc.Delete(id); er != nil {
				utils.Error(w, 404, er.Error())
				return
			}
			utils.JSON(w, 200, map[string]bool{"deleted": true})
		default:
			utils.Error(w, 405, "method not allowed")
		}
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		emps := svc.List("", "", "", "")
		deps := map[string]int{}
		active := 0
		salary := 0.0
		for _, e := range emps {
			deps[e.Department]++
			if strings.EqualFold(e.Status, "Active") {
				active++
			}
			salary += e.Salary
		}
		utils.JSON(w, 200, map[string]any{"total": len(emps), "active": active, "inactive": len(emps) - active, "departments": deps, "averageSalary": func() float64 {
			if len(emps) == 0 {
				return 0
			}
			return salary / float64(len(emps))
		}()})
	})
	for name, c := range cols {
		registerCollection(mux, name, c)
	}
	mux.HandleFunc("/api/action-required", func(w http.ResponseWriter, r *http.Request) {
		leaves := cols["leaves"].All()
		tickets := cols["tickets"].All()
		docs := cols["documents"].All()
		pendingL := 0
		openT := 0
		missing := 0
		for _, v := range leaves {
			if strings.EqualFold(fmt.Sprint(v["status"]), "Pending") {
				pendingL++
			}
		}
		for _, v := range tickets {
			if !strings.EqualFold(fmt.Sprint(v["status"]), "Resolved") {
				openT++
			}
		}
		for _, v := range docs {
			if strings.EqualFold(fmt.Sprint(v["status"]), "Missing") {
				missing++
			}
		}
		utils.JSON(w, 200, map[string]int{"leaveRequests": pendingL, "openTickets": openT, "missingDocuments": missing})
	})
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("EmployeeHub: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func registerCollection(mux *http.ServeMux, name string, c *JSONCollection) {
	path := "/api/" + name
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			utils.JSON(w, 200, c.All())
		case "POST":
			var v map[string]any
			if json.NewDecoder(r.Body).Decode(&v) != nil {
				utils.Error(w, 400, "invalid JSON")
				return
			}
			utils.JSON(w, 201, c.Add(v))
		default:
			utils.Error(w, 405, "method not allowed")
		}
	})
	mux.HandleFunc(path+"/", func(w http.ResponseWriter, r *http.Request) {
		id, e := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, path+"/"), "/"))
		if e != nil {
			utils.Error(w, 400, "invalid id")
			return
		}
		switch r.Method {
		case "PUT":
			var v map[string]any
			if json.NewDecoder(r.Body).Decode(&v) != nil {
				utils.Error(w, 400, "invalid JSON")
				return
			}
			x, ok := c.Update(id, v)
			if !ok {
				utils.Error(w, 404, "record not found")
				return
			}
			utils.JSON(w, 200, x)
		case "DELETE":
			if !c.Delete(id) {
				utils.Error(w, 404, "record not found")
				return
			}
			utils.JSON(w, 200, map[string]bool{"deleted": true})
		default:
			utils.Error(w, 405, "method not allowed")
		}
	})
}

func seed(s *patterns.Store) {
	now := time.Now()
	s.Employees = []models.Employee{{ID: 1, Name: "Aarav Sharma", Email: "aarav@example.com", Phone: "9876543210", Department: "Engineering", Position: "Software Engineer", Salary: 75000, Status: "Active", EmploymentType: "Full-time", Location: "Bengaluru", JoinDate: now.AddDate(0, -10, 0)}, {ID: 2, Name: "Priya Mehta", Email: "priya@example.com", Phone: "9876543211", Department: "HR", Position: "HR Manager", Salary: 68000, Status: "Active", EmploymentType: "Full-time", Location: "Gurugram", JoinDate: now.AddDate(0, -7, 0)}, {ID: 3, Name: "Rohan Verma", Email: "rohan@example.com", Phone: "9876543212", Department: "Finance", Position: "Financial Analyst", Salary: 62000, Status: "Inactive", EmploymentType: "Full-time", Location: "Mumbai", JoinDate: now.AddDate(0, -4, 0)}, {ID: 4, Name: "Ananya Singh", Email: "ananya@example.com", Phone: "9876543213", Department: "Engineering", Position: "UI Developer", Salary: 71000, Status: "Active", EmploymentType: "Contract", Location: "Hyderabad", JoinDate: now.AddDate(0, -2, 0)}}
	_ = s.Save()
}

func seedHRCollections(cols map[string]*JSONCollection) {
	if len(cols["attendance"].All()) == 0 {
		for _, v := range []map[string]any{
			{"employeeId": 1, "employeeName": "Aarav Sharma", "date": time.Now().Format("2006-01-02"), "status": "Present", "checkIn": "09:12", "checkOut": "18:03"},
			{"employeeId": 2, "employeeName": "Priya Mehta", "date": time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "status": "Present", "checkIn": "09:01", "checkOut": "18:10"},
			{"employeeId": 4, "employeeName": "Ananya Singh", "date": time.Now().AddDate(0, 0, -2).Format("2006-01-02"), "status": "Late", "checkIn": "10:18", "checkOut": "18:30"},
		} {
			cols["attendance"].Add(v)
		}
	}
	if len(cols["leaves"].All()) == 0 {
		cols["leaves"].Add(map[string]any{"employeeId": 1, "employeeName": "Aarav Sharma", "type": "Casual", "start": time.Now().AddDate(0, 0, 5).Format("2006-01-02"), "end": time.Now().AddDate(0, 0, 6).Format("2006-01-02"), "reason": "Personal work", "status": "Pending"})
		cols["leaves"].Add(map[string]any{"employeeId": 2, "employeeName": "Priya Mehta", "type": "Earned", "start": time.Now().AddDate(0, 0, -10).Format("2006-01-02"), "end": time.Now().AddDate(0, 0, -9).Format("2006-01-02"), "reason": "Family event", "status": "Approved"})
	}
	if len(cols["payroll"].All()) == 0 {
		cols["payroll"].Add(map[string]any{"employeeName": "Aarav Sharma", "month": time.Now().Format("2006-01"), "grossSalary": 75000, "deductions": 6500, "netSalary": 68500, "status": "Paid"})
		cols["payroll"].Add(map[string]any{"employeeName": "Priya Mehta", "month": time.Now().Format("2006-01"), "grossSalary": 68000, "deductions": 5900, "netSalary": 62100, "status": "Paid"})
	}
	if len(cols["performance"].All()) == 0 {
		cols["performance"].Add(map[string]any{"employeeName": "Aarav Sharma", "reviewDate": time.Now().AddDate(0, -1, 0).Format("2006-01-02"), "rating": "4 — Strong", "goal": "Lead a backend improvement", "managerComment": "Consistent delivery and ownership."})
		cols["performance"].Add(map[string]any{"employeeName": "Priya Mehta", "reviewDate": time.Now().AddDate(0, -2, 0).Format("2006-01-02"), "rating": "5 — Exceptional", "goal": "Improve onboarding process", "managerComment": "Strong people leadership."})
	}
	if len(cols["tickets"].All()) == 0 {
		cols["tickets"].Add(map[string]any{"title": "Laptop access issue", "employeeName": "Ananya Singh", "priority": "High", "category": "IT", "description": "VPN access is not working.", "status": "Open"})
		cols["tickets"].Add(map[string]any{"title": "HR portal access", "employeeName": "Rohan Verma", "priority": "Medium", "category": "Access", "description": "Needs access to HR portal.", "status": "In progress"})
	}
	if len(cols["documents"].All()) == 0 {
		cols["documents"].Add(map[string]any{"employeeName": "Aarav Sharma", "type": "Resume", "fileName": "aarav-resume.pdf", "status": "Verified"})
		cols["documents"].Add(map[string]any{"employeeName": "Priya Mehta", "type": "ID proof", "fileName": "", "status": "Missing"})
		cols["documents"].Add(map[string]any{"employeeName": "Ananya Singh", "type": "Offer letter", "fileName": "ananya-offer.pdf", "status": "Verified"})
	}
	if len(cols["events"].All()) == 0 {
		cols["events"].Add(map[string]any{"title": "Engineering sync", "type": "Team event", "date": time.Now().AddDate(0, 0, 2).Format("2006-01-02"), "location": "Meeting Room A", "description": "Weekly engineering planning."})
		cols["events"].Add(map[string]any{"title": "Company holiday", "type": "Holiday", "date": time.Now().AddDate(0, 0, 7).Format("2006-01-02"), "location": "Company", "description": "Office holiday."})
	}
	if len(cols["onboarding"].All()) == 0 {
		cols["onboarding"].Add(map[string]any{"employeeName": "Ananya Singh", "position": "UI Developer", "startDate": time.Now().AddDate(0, 0, -5).Format("2006-01-02"), "completed": 4, "total": 6})
		cols["onboarding"].Add(map[string]any{"employeeName": "Rohan Verma", "position": "Financial Analyst", "startDate": time.Now().AddDate(0, 0, -2).Format("2006-01-02"), "completed": 2, "total": 6})
	}
}
