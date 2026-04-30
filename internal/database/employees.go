package database

import (
	"fmt"
	"schedulemate/internal/models"
)

func GetAllEmployees() ([]models.Employee, error) {
	rows, err := DB.Query("SELECT id, name, role, desired_hours, wage FROM employee")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := []models.Employee{}
	for rows.Next() { // iter
		var e models.Employee
		err := rows.Scan(&e.ID, &e.Name, &e.Role, &e.DesiredHours, &e.Wage)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	fmt.Printf("Returning %d employees\n", len(employees))
	return employees, nil
}

func CreateEmployee(e models.Employee) (int64, error) {
	if e.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if e.DesiredHours <= 0 {
		return 0, fmt.Errorf("desired hours must be greater than 0")
	}
	fmt.Printf("Adding employee named %s\n", e.Name)
	result, err := DB.Exec(
		"INSERT INTO employee (name, role, desired_hours, wage) VALUES (?, ?, ?, ?)",
		e.Name, e.Role, e.DesiredHours, e.Wage,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateEmployee(e models.Employee) error {
	fmt.Printf("Updating employee id=%d name=%s\n", e.ID, e.Name)
	_, err := DB.Exec(
		"UPDATE employee SET name=?, role=?, desired_hours=?, wage=? WHERE id=?",
		e.Name, e.Role, e.DesiredHours, e.Wage, e.ID,
	)
	return err
}

func DeleteEmployee(id int) error {
	result, err := DB.Exec("DELETE FROM employee WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %d not found", id)
	}
	return nil
}
