package database

import (
	"fmt"
	"schedulemate/internal/models"
)

func GetAvailabilityByEmployee(employeeID int) ([]models.Availability, error) {
	rows, err := DB.Query(
		"SELECT * FROM availability WHERE employee_id = ?",
		employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // close the database connection when we're done with it

	availability := []models.Availability{}
	for rows.Next() { // loop through every row of results and convert to Availability struct
		var a models.Availability
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.DayOfWeek, &a.StartTime, &a.EndTime); err != nil {
			return nil, err
		}
		availability = append(availability, a)
	}
	return availability, nil
}

func GetAllAvailabilities() ([]models.Availability, error) {
	rows, err := DB.Query(
		"SELECT * FROM availability",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	availability := []models.Availability{}
	for rows.Next() {
		var a models.Availability
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.DayOfWeek, &a.StartTime, &a.EndTime); err != nil {
			return nil, err
		}
		availability = append(availability, a)
	}
	return availability, nil
}

func CreateAvailability(a models.Availability) (int64, models.ValidationResult) {
	// input validation
	if a.EmployeeID == 0 {
		return 0, models.ValidationResult{Errors: []string{"Employee ID is required"}}
	}
	if a.DayOfWeek < 1 || a.DayOfWeek > 7 {
		return 0, models.ValidationResult{Errors: []string{"Day of week must be between 1 and 7"}}
	}
	if a.StartTime == "" || a.EndTime == "" {
		return 0, models.ValidationResult{Errors: []string{"Start time and end time are required"}}
	}
	if a.StartTime >= a.EndTime {
		return 0, models.ValidationResult{Errors: []string{"Start time must be before end time"}}
	}

	// conflict detection:
	// result := CheckAvailabilityConflict(a)
	// if len(result.Errors) > 0 {
	// 	return 0, models.ValidationResult{Errors: result.Errors}
	// }
	// pass result.Warnings back up to caller when wired in

	res, err := DB.Exec(
		"INSERT INTO availability (employee_id, day_of_week, start_time, end_time) VALUES (?, ?, ?, ?)",
		a.EmployeeID, a.DayOfWeek, a.StartTime, a.EndTime,
	)
	if err != nil {
		return 0, models.ValidationResult{Fatal: err}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, models.ValidationResult{Fatal: err}
	}
	return id, models.ValidationResult{}
}

func UpdateAvailability(a models.Availability) models.ValidationResult {
	// input validation
	if a.ID == 0 {
		return models.ValidationResult{Errors: []string{"ID is required"}}
	}
	if a.DayOfWeek < 1 || a.DayOfWeek > 7 {
		return models.ValidationResult{Errors: []string{"Day of week must be between 1 and 7"}}
	}
	if a.StartTime == "" || a.EndTime == "" {
		return models.ValidationResult{Errors: []string{"Start time and end time are required"}}
	}
	if a.StartTime >= a.EndTime {
		return models.ValidationResult{Errors: []string{"Start time must be before end time"}}
	}

	// conflict detection
	// result := CheckAvailabilityConflict(a)
	// if len(result.Errors) > 0 {
	// 	return models.ValidationResult{Errors: result.Errors}
	// }

	_, err := DB.Exec(
		"UPDATE availability SET day_of_week=?, start_time=?, end_time=? WHERE id=?",
		a.DayOfWeek, a.StartTime, a.EndTime, a.ID,
	)
	if err != nil {
		return models.ValidationResult{Fatal: err}
	}
	return models.ValidationResult{}
}

func DeleteAvailability(id int) error {
	result, err := DB.Exec("DELETE FROM availability WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("availability with id %d not found", id)
	}
	return nil
}
