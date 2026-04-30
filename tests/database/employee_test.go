package database_test

import (
	"schedulemate/internal/database"
	"schedulemate/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) {
	t.Helper()
	err := database.InitDB(":memory:")
	require.NoError(t, err, "failed to init test db")
}

func TestCreateEmployee_AppearsInGetAll(t *testing.T) {
	setup(t)

	_, err := database.CreateEmployee(models.Employee{Name: "Alice", Role: "Cashier", DesiredHours: 32, Wage: 15.50})
	require.NoError(t, err)

	employees, err := database.GetAllEmployees()
	require.NoError(t, err)
	require.Len(t, employees, 1)
	assert.Equal(t, "Alice", employees[0].Name)
}

func TestCreateEmployee_MissingRequiredFields(t *testing.T) {
	setup(t)

	_, err := database.CreateEmployee(models.Employee{Name: "", Role: "Cashier", DesiredHours: 32})
	assert.Error(t, err)
}

func TestDeleteEmployee_RemovedFromGetAll(t *testing.T) {
	setup(t)

	id, err := database.CreateEmployee(models.Employee{Name: "Bob", Role: "Stock", DesiredHours: 20, Wage: 14.00})
	require.NoError(t, err)

	err = database.DeleteEmployee(int(id))
	require.NoError(t, err)

	employees, err := database.GetAllEmployees()
	require.NoError(t, err)
	assert.Empty(t, employees)
}

func TestDeleteEmployee_NotFound(t *testing.T) {
	setup(t)

	err := database.DeleteEmployee(9999)
	assert.Error(t, err)
}

func TestUpdateEmployee_ChangesReflected(t *testing.T) {
	setup(t)

	id, err := database.CreateEmployee(models.Employee{Name: "Carol", Role: "Cashier", DesiredHours: 24, Wage: 15.00})
	require.NoError(t, err)

	err = database.UpdateEmployee(models.Employee{ID: int(id), Name: "Carol Updated", Role: "Supervisor", DesiredHours: 40, Wage: 20.00})
	require.NoError(t, err)

	employees, err := database.GetAllEmployees()
	require.NoError(t, err)
	assert.Equal(t, "Carol Updated", employees[0].Name)
	assert.Equal(t, "Supervisor", employees[0].Role)
	assert.Equal(t, 40, employees[0].DesiredHours)
	assert.Equal(t, 20.00, employees[0].Wage)
}
