package main

import (
	"context"
	"fmt"

	"schedulemate/internal/database"
	"schedulemate/internal/models"
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

func (a *App) ClearData() error {
	return database.ClearData()
}
