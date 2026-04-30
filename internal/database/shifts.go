package database

import (
	"fmt"
	"schedulemate/internal/models"
	// "schedulemate/internal/services"
)

func GetShiftsBySchedule(scheduleID int) ([]models.Shift, error) {
	rows, err := DB.Query(
		"SELECT * FROM shift WHERE schedule_id = ?",
		scheduleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shifts := []models.Shift{}
	for rows.Next() {
		var s models.Shift
		if err := rows.Scan(&s.ID, &s.ScheduleID, &s.EmployeeID, &s.DayOfWeek, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, nil
}

func GetShiftsByEmployee(employeeID int) ([]models.Shift, error) {
	rows, err := DB.Query(
		"SELECT * FROM shift WHERE employee_id = ?",
		employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shifts := []models.Shift{}
	for rows.Next() {
		var s models.Shift
		if err := rows.Scan(&s.ID, &s.ScheduleID, &s.EmployeeID, &s.DayOfWeek, &s.StartTime, &s.EndTime); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, nil
}

func CreateShift(s models.Shift) (int64, models.ValidationResult) {
	// input validation
	if s.ScheduleID == 0 {
		return 0, models.ValidationResult{Errors: []string{"Schedule ID is required"}}
	}
	if s.EmployeeID == 0 {
		return 0, models.ValidationResult{Errors: []string{"Employee ID is required"}}
	}
	if s.DayOfWeek < 1 || s.DayOfWeek > 7 {
		return 0, models.ValidationResult{Errors: []string{"Day of week must be between 1 and 7"}}
	}
	if s.StartTime == "" || s.EndTime == "" {
		return 0, models.ValidationResult{Errors: []string{"Start time and end time are required"}}
	}
	if s.StartTime >= s.EndTime {
		return 0, models.ValidationResult{Errors: []string{"Start time must be before end time"}}
	}

	// conflict detection
	// result := CheckShiftConflict(s)
	// if len(result.Errors) > 0 {
	// 	return 0, fmt.Errorf(strings.Join(result.Errors, "; "))
	// }
	// pass result.Warnings back up to caller when wired in

	res, err := DB.Exec(
		"INSERT INTO shift (schedule_id, employee_id, day_of_week, start_time, end_time) VALUES (?, ?, ?, ?, ?)",
		s.ScheduleID, s.EmployeeID, s.DayOfWeek, s.StartTime, s.EndTime,
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

func UpdateShift(s models.Shift) (models.ValidationResult) {
	// Layer 1: input validation
	if s.ID == 0 {
		return models.ValidationResult{Errors: []string{"ID is required"}}
	}
	if s.DayOfWeek < 1 || s.DayOfWeek > 7 {
		return models.ValidationResult{Errors: []string{"Day of week must be between 1 and 7"}}
	}
	if s.StartTime == "" || s.EndTime == "" {
		return models.ValidationResult{Errors: []string{"Start time and end time are required"}}
	}
	if s.StartTime >= s.EndTime {
		return models.ValidationResult{Errors: []string{"Start time must be before end time"}}
	}

	// conflict detection:
	// result := CheckShiftConflict(s)
	// if len(result.Errors) > 0 {
	// 	return fmt.Errorf(strings.Join(result.Errors, "; "))
	// }

	_, err := DB.Exec(
		"UPDATE shift SET schedule_id=?, employee_id=?, day_of_week=?, start_time=?, end_time=? WHERE id=?",
		s.ScheduleID, s.EmployeeID, s.DayOfWeek, s.StartTime, s.EndTime, s.ID,
	)
	if err != nil {
		return models.ValidationResult{Fatal: err}
	}
	return models.ValidationResult{}
}

func DeleteShift(id int) error {
	result, err := DB.Exec("DELETE FROM shift WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("shift with id %d not found", id)
	}
	return nil
}
