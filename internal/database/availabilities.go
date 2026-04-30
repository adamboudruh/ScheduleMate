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

func CreateAvailability(a models.Availability) (int64, error) {
	// input validation
	if a.EmployeeID == 0 {
		return 0, fmt.Errorf("employee_id is required")
	}
	if a.DayOfWeek < 1 || a.DayOfWeek > 7 {
		return 0, fmt.Errorf("day_of_week must be between 1 and 7")
	}
	if a.StartTime == "" || a.EndTime == "" {
		return 0, fmt.Errorf("start_time and end_time are required")
	}
	if a.StartTime >= a.EndTime {
		return 0, fmt.Errorf("start_time must be before end_time")
	}

	// conflict detection:
	// result := CheckAvailabilityConflict(a)
	// if len(result.Errors) > 0 {
	// 	return 0, fmt.Errorf(strings.Join(result.Errors, "; "))
	// }
	// pass result.Warnings back up to caller when wired in

	res, err := DB.Exec(
		"INSERT INTO availability (employee_id, day_of_week, start_time, end_time) VALUES (?, ?, ?, ?)",
		a.EmployeeID, a.DayOfWeek, a.StartTime, a.EndTime,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateAvailability(a models.Availability) error {
	// input validation
	if a.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if a.DayOfWeek < 1 || a.DayOfWeek > 7 {
		return fmt.Errorf("day_of_week must be between 1 and 7")
	}
	if a.StartTime == "" || a.EndTime == "" {
		return fmt.Errorf("start_time and end_time are required")
	}
	if a.StartTime >= a.EndTime {
		return fmt.Errorf("start_time must be before end_time")
	}

	// conflict detection
	// result := CheckAvailabilityConflict(a)
	// if len(result.Errors) > 0 {
	// 	return fmt.Errorf(strings.Join(result.Errors, "; "))
	// }

	_, err := DB.Exec(
		"UPDATE availability SET day_of_week=?, start_time=?, end_time=? WHERE id=?",
		a.DayOfWeek, a.StartTime, a.EndTime, a.ID,
	)
	return err
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
