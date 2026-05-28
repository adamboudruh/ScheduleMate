package models

type Employee struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Role         string  `json:"role"`
	DesiredHours int     `json:"desiredHours"`
	MaxHours     int     `json:"maxHours"`
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

type DaySettings struct {
	DayOfWeek        int    `json:"dayOfWeek"`
	OpenTime         string `json:"openTime"`
	CloseTime        string `json:"closeTime"`
	SchedulableOpen  string `json:"schedulableOpen"`  // nullable
	SchedulableClose string `json:"schedulableClose"` // nullable
	EmployeesNeeded  int    `json:"employeesNeeded"`
}

type Settings struct {
	ID                int  `json:"id"`
	MinShiftLength    int  `json:"minShiftLength"`
	MaxShiftLength    int  `json:"maxShiftLength"`
	AllowOutsideHours bool `json:"allowOutsideHours"`
	TimeBetweenShifts int  `json:"timeBetweenShifts"`
}

// ValidationResult Struct for conflict detection

type ValidationResult struct {
	Errors   []string // blocks the save - business logic violations
	Warnings []string // doesn't block - soft constraints, shown to user
	Fatal    error    // unexpected failure - DB error, not a validation issue
}
