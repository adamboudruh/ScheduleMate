package main

import (
	"context"
	"fmt"

	"schedulemate/internal/database"
	"schedulemate/internal/models"
	"schedulemate/internal/solver"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	println("Initializing database...")
	if err := database.InitDB(""); err != nil {
		panic(err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Addition(x, y float64) float64 {
	return x + y
}

func (a *App) GetAllEmployees() ([]models.Employee, error) {
	return database.GetAllEmployees()
}

func (a *App) CreateEmployee(e models.Employee) (int64, error) {
	return database.CreateEmployee(e)
}

func (a *App) UpdateEmployee(e models.Employee) error {
	return database.UpdateEmployee(e)
}

func (a *App) DeleteEmployee(id int) error {
	return database.DeleteEmployee(id)
}

func (a *App) Seed() error {
	return database.Seed()
}

func (a *App) GenerateSchedule(scheduleID int) (solver.SolverResult, error) {
	// load everything from DB into memory
	employees, err := database.GetAllEmployees()
	if err != nil {
		return solver.SolverResult{}, err
	}

	avails, err := database.GetAllAvailabilities()
	if err != nil {
		return solver.SolverResult{}, err
	}

	daySettings, err := database.GetAllDaySettings()
	if err != nil {
		return solver.SolverResult{}, err
	}

	settings, err := database.GetSettings()
	if err != nil {
		return solver.SolverResult{}, err
	}

	// hand it all to the solver — it runs entirely in memory
	result := solver.Run(solver.SolverInput{
		Employees:      employees,
		Availabilities: avails,
		DaySettings:    daySettings,
		Settings:       settings,
	})

	// if solved, write shifts to DB
	if result.Solved {
		database.DeleteShiftsBySchedule(scheduleID) // clear out any old shifts for this schedule
		for i := range result.Shifts {
			result.Shifts[i].ScheduleID = scheduleID
		}
		database.CreateShiftsBulk(result.Shifts)
	}

	return result, nil
}

func (a *App) ClearData() error {
	return database.ClearData()
}

func (a *App) GetAllAvailabilities() ([]models.Availability, error) {
	return database.GetAllAvailabilities()
}

func (a *App) GetAvailabilityByEmployee(employeeID int) ([]models.Availability, error) {
	return database.GetAvailabilityByEmployee(employeeID)
}

func (a *App) CreateAvailability(av models.Availability) (int64, models.ValidationResult) {
	return database.CreateAvailability(av)
}

func (a *App) UpdateAvailability(av models.Availability) models.ValidationResult {
	return database.UpdateAvailability(av)
}

func (a *App) DeleteAvailability(id int) error {
	return database.DeleteAvailability(id)
}

func (a *App) GetAllSchedules() ([]models.Schedule, error) {
	return database.GetAllSchedules()
}

func (a *App) GetScheduleByWeek(weekOf string) (models.Schedule, error) {
	return database.GetScheduleByWeek(weekOf)
}

func (a *App) CreateSchedule(s models.Schedule) (int64, error) {
	return database.CreateSchedule(s)
}

func (a *App) UpdateSchedule(s models.Schedule) error {
	return database.UpdateSchedule(s)
}

func (a *App) DeleteSchedule(id int) error {
	return database.DeleteSchedule(id)
}

func (a *App) GetShiftsBySchedule(scheduleID int) ([]models.Shift, error) {
	return database.GetShiftsBySchedule(scheduleID)
}

func (a *App) GetShiftsByEmployee(employeeID int) ([]models.Shift, error) {
	return database.GetShiftsByEmployee(employeeID)
}

func (a *App) CreateShift(s models.Shift) (int64, models.ValidationResult) {
	return database.CreateShift(s)
}

func (a *App) UpdateShift(s models.Shift) models.ValidationResult {
	return database.UpdateShift(s)
}

func (a *App) DeleteShift(id int) error {
	return database.DeleteShift(id)
}

func (a *App) DeleteShiftsBySchedule(scheduleID int) error {
	return database.DeleteShiftsBySchedule(scheduleID)
}

func (a *App) GetAllDaySettings() (map[int]models.DaySettings, error) {
	return database.GetAllDaySettings()
}

func (a *App) GetDaySettings(dayOfWeek int) (models.DaySettings, error) {
	return database.GetDaySettings(dayOfWeek)
}

func (a *App) UpdateDaySettings(ds models.DaySettings) error {
	return database.UpdateDaySettings(ds)
}

func (a *App) GetSettings() (models.Settings, error) {
	return database.GetSettings()
}

func (a *App) UpdateSettings(s models.Settings) error {
	return database.UpdateSettings(s)
}
