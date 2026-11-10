# EmployeeHub — Final College Project

A self-contained Go-powered Employee & HR Management web application designed for a college project.

## No database setup

This project intentionally uses local JSON files in `data/`.

You do **not** need MongoDB, MongoDB Atlas, MySQL, PostgreSQL, a database URL, API keys, or an `.env` file.

## Run in VS Code

1. Open this folder (the folder containing `go.mod` and `main.go`) in VS Code.
2. Run:

```bash
go run .
```

3. Open `http://localhost:8080`.

If port 8080 is busy:

**PowerShell**
```powershell
$env:PORT="8081"
go run .
```

Then open `http://localhost:8081`.

## Included modules

- Demo login with Admin / HR / Employee role selection
- Workforce dashboard and action-required alerts
- Employee directory, profiles, favorites, bulk actions, search/filter
- Employee onboarding and checklist progress
- Manager/department organization view
- Attendance records
- Leave requests and approve/reject workflow
- Payroll / payslip register
- Performance reviews and goals
- Company calendar/events
- Help desk/support tickets
- Employee document metadata vault
- Reports, CSV export, summary export and print directory
- Local activity/audit trail
- Command palette (`Ctrl + K`)
- Light/dark theme
- Local JSON persistence

## Data files

The server automatically creates and uses:

`data/employees.json`, `attendance.json`, `leaves.json`, `payroll.json`, `performance.json`, `tickets.json`, `documents.json`, `events.json`, `onboarding.json`, and `notifications.json`.

This keeps the project portable and easy to demonstrate on a college computer.
