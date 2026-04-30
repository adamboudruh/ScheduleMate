package database

func Seed() error {
	employees := []struct {
		name         string
		role         string
		desiredHours int
		wage         float64
	}{
		{"Alice Johnson", "Cashier", 32, 15.50},
		{"Bob Smith", "Supervisor", 40, 22.00},
		{"Carol White", "Cashier", 24, 15.50},
		{"David Lee", "Stock", 20, 14.00},
	}

	for _, e := range employees {
		_, err := DB.Exec(
			"INSERT INTO employee (name, role, desired_hours, wage) VALUES (?, ?, ?, ?)",
			e.name, e.role, e.desiredHours, e.wage,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ClearData() error {
	_, err := DB.Exec("DELETE FROM employee")
	return err
}
