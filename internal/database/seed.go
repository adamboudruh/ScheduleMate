package database

import (
	"fmt"
	"log"
)

func Seed() error {
	fmt.Print("Seeding db...\n")

	if err := ClearData(); err != nil {
		return fmt.Errorf("clearing data: %w", err)
	}

	employees := []struct {
		name         string
		role         string
		desiredHours int
		maxHours     int
		wage         float64
	}{
		{"Alice", "cashier", 20, 40, 15.50},
		{"Bob", "cashier", 20, 40, 15.50},
		{"Charlie", "stocker", 40, 40, 15.00},
		{"Diana", "stocker", 40, 40, 15.00},
	}

	fmt.Printf("Seeding %d employees...\n", len(employees))

	availabilities := []struct {
		employeeIndex int
		dayOfWeek     int
		startTime     string
		endTime       string
	}{
		// Alice: Mon, Tue, Wed(closed), Thu(off), Fri, Sat(off), Sun(closed)
		{0, 2, "09:00", "21:00"},
		{0, 3, "09:00", "21:00"},
		{0, 5, "09:00", "21:00"},
		{0, 6, "09:00", "21:00"},
		{0, 7, "09:00", "21:00"},

		// Bob: Mon, Tue(off), Wed(closed), Thu, Fri(off), Sat, Sun(closed)
		{1, 2, "09:00", "21:00"},
		{1, 3, "09:00", "21:00"},
		{1, 5, "09:00", "21:00"},
		{1, 6, "09:00", "21:00"},
		{1, 7, "09:00", "21:00"},

		// Charlie: Mon, Tue, Wed(closed+off), Thu, Fri, Sat, Sun(closed+off)
		{2, 2, "09:00", "21:00"},
		{2, 3, "09:00", "21:00"},
		{2, 5, "09:00", "21:00"},
		{2, 6, "09:00", "21:00"},
		{2, 7, "09:00", "21:00"},

		// Diana: Mon(off), Tue, Wed(closed), Thu, Fri, Sat, Sun(closed)
		{3, 2, "09:00", "21:00"},
		{3, 3, "09:00", "21:00"},
		{3, 5, "09:00", "21:00"},
		{3, 6, "09:00", "21:00"},
		{3, 7, "09:00", "21:00"},
	}

	daySettings := map[int]struct {
		open             string
		close            string
		schedulableOpen  string
		schedulableClose string
		needed           int
	}{
		1: {"00:00", "00:00", "00:00", "00:00", 0},
		2: {"09:00", "21:00", "09:00", "21:00", 1},
		3: {"09:00", "21:00", "09:00", "21:00", 1},
		4: {"00:00", "00:00", "00:00", "00:00", 0},
		5: {"09:00", "21:00", "09:00", "21:00", 1},
		6: {"09:00", "21:00", "09:00", "21:00", 2},
		7: {"09:00", "21:00", "09:00", "21:00", 2},
	}

	employeeIDs := make([]int64, len(employees))
	for i, emp := range employees {
		fmt.Printf("  Inserting employee: %s...\n", emp.name)
		result, err := DB.Exec(
			`INSERT INTO employee (name, role, desired_hours, max_hours, wage) VALUES (?, ?, ?, ?, ?)`,
			emp.name, emp.role, emp.desiredHours, emp.maxHours, emp.wage,
		)
		if err != nil {
			err = fmt.Errorf("inserting employee %s: %w", emp.name, err)
			log.Println("Seed error:", err)
			return err
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

	for day, ds := range daySettings {
		if _, err := DB.Exec(`
			UPDATE day_settings
			SET open_time=?, close_time=?, schedulable_open=?, schedulable_close=?, employees_needed=?
			WHERE day_of_week=?`,
			ds.open, ds.close, ds.schedulableOpen, ds.schedulableClose, ds.needed, day,
		); err != nil {
			return fmt.Errorf("updating day_settings for day %d: %w", day, err)
		}
	}
	return nil
}

func ClearData() error {
	if _, err := DB.Exec("DELETE FROM schedule"); err != nil {
		return err
	}
	_, err := DB.Exec("DELETE FROM employee")
	return err
}
