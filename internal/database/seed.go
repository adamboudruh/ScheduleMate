package database

import (
	"fmt"
)

func Seed() error {
	employees := []struct {
		name         string
		role         string
		desiredHours int
		maxHours     int
		wage         float64
	}{
		{"Alice", "cashier", 24, 32, 15.50},
		{"Bob", "cashier", 16, 20, 15.50},
		{"Charlie", "cashier", 32, 40, 16.00},
		{"Diana", "stocker", 20, 28, 15.00},
		{"Eve", "stocker", 32, 40, 15.00},
		{"Frank", "stocker", 16, 24, 15.00},
	}

	availabilities := []struct {
		employeeIndex int
		dayOfWeek     int
		startTime     string
		endTime       string
	}{
		{0, 2, "09:00", "17:00"},
		{0, 3, "09:00", "17:00"},
		{0, 5, "12:00", "21:00"},
		{0, 6, "09:00", "21:00"},
		{0, 7, "09:00", "17:00"},

		{1, 2, "12:00", "21:00"},
		{1, 4, "12:00", "21:00"},
		{1, 6, "09:00", "18:00"},
		{1, 7, "12:00", "21:00"},

		{2, 2, "09:00", "21:00"},
		{2, 3, "13:00", "21:00"},
		{2, 4, "09:00", "21:00"},
		{2, 5, "09:00", "21:00"},
		{2, 6, "13:00", "21:00"},
		{2, 7, "09:00", "21:00"},

		{3, 2, "09:00", "17:00"},
		{3, 3, "09:00", "17:00"},
		{3, 4, "09:00", "17:00"},
		{3, 6, "09:00", "15:00"},
		{3, 7, "09:00", "15:00"},

		{4, 2, "09:00", "21:00"},
		{4, 3, "09:00", "21:00"},
		{4, 4, "09:00", "21:00"},
		{4, 5, "09:00", "21:00"},
		{4, 6, "09:00", "21:00"},
		{4, 7, "09:00", "21:00"},

		{5, 3, "14:00", "21:00"},
		{5, 4, "14:00", "21:00"},
		{5, 5, "14:00", "21:00"},
		{5, 6, "09:00", "21:00"},
		{5, 7, "09:00", "21:00"},
	}

	employeeIDs := make([]int64, len(employees))
	for i, emp := range employees {
		result, err := DB.Exec(
			`INSERT INTO employee (name, role, desired_hours, max_hours, wage) VALUES (?, ?, ?, ?, ?)`,
			emp.name, emp.role, emp.desiredHours, emp.maxHours, emp.wage,
		)
		if err != nil {
			return fmt.Errorf("inserting employee %s: %w", emp.name, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		employeeIDs[i] = id
		fmt.Printf("  Inserted employee: %s (id=%d)\n", emp.name, id)
	}

	for _, avail := range availabilities {
		empID := employeeIDs[avail.employeeIndex]
		if _, err := DB.Exec(
			`INSERT INTO availability (employee_id, day_of_week, start_time, end_time) VALUES (?, ?, ?, ?)`,
			empID, avail.dayOfWeek, avail.startTime, avail.endTime,
		); err != nil {
			return fmt.Errorf("inserting availability for employee_id=%d day=%d: %w", empID, avail.dayOfWeek, err)
		}
	}
	return nil
}

func ClearData() error {
	_, err := DB.Exec("DELETE FROM employee")
	return err
}
