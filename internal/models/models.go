package models

type Employee struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Role         string  `json:"role"`
	DesiredHours int     `json:"desiredHours"`
	Wage         float64 `json:"wage"`
}

type Availability struct {
	ID         int    `json:"id"`
	EmployeeID int    `json:"employeeId"`
	DayOfWeek  int    `json:"dayOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type Schedule struct {
	ID     int    `json:"id"`
	WeekOf string `json:"weekOf"`
	Notes  string `json:"notes"`
}

type Shift struct {
	ID         int    `json:"id"`
	ScheduleID int    `json:"scheduleId"`
	EmployeeID int    `json:"employeeId"`
	DayOfWeek  int    `json:"dayOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

// -- Validation Result Struct for conflict detection --

type ValidationResult struct {
	Errors   []string // blocks the save — business logic violations
	Warnings []string // doesn't block — soft constraints, shown to user
	Fatal    error    // unexpected failure — DB error, not a validation issue
}
