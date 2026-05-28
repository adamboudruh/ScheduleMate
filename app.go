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
