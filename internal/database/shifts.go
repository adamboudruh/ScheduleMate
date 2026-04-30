package database

import (
	"fmt"
	"schedulemate/internal/models"
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

func CreateShift(s models.Shift) (int64, error) {
	// input validation
	if s.ScheduleID == 0 {
		return 0, fmt.Errorf("schedule_id is required")
	}
	if s.EmployeeID == 0 {
		return 0, fmt.Errorf("employee_id is required")
	}
	if s.DayOfWeek < 1 || s.DayOfWeek > 7 {
		return 0, fmt.Errorf("day_of_week must be between 1 and 7")
	}
	if s.StartTime == "" || s.EndTime == "" {
		return 0, fmt.Errorf("start_time and end_time are required")
	}
	if s.StartTime >= s.EndTime {
		return 0, fmt.Errorf("start_time must be before end_time")
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
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateShift(s models.Shift) error {
	// Layer 1: input validation
	if s.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if s.DayOfWeek < 1 || s.DayOfWeek > 7 {
		return fmt.Errorf("day_of_week must be between 1 and 7")
	}
	if s.StartTime == "" || s.EndTime == "" {
		return fmt.Errorf("start_time and end_time are required")
	}
	if s.StartTime >= s.EndTime {
		return fmt.Errorf("start_time must be before end_time")
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
	return err
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
