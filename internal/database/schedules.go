package database

import (
	"fmt"
	"schedulemate/internal/models"
)

func GetAllSchedules() ([]models.Schedule, error) {
	rows, err := DB.Query("SELECT id, week_of, notes FROM schedule")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := []models.Schedule{}
	for rows.Next() {
		var s models.Schedule
		if err := rows.Scan(&s.ID, &s.WeekOf, &s.Notes); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func GetScheduleByWeek(weekOf string) (models.Schedule, error) {
	var s models.Schedule
	err := DB.QueryRow("SELECT id, week_of, notes FROM schedule WHERE week_of = ?", weekOf).
		Scan(&s.ID, &s.WeekOf, &s.Notes)
	if err != nil {
		return models.Schedule{}, fmt.Errorf("schedule for week %s not found", weekOf)
	}
	return s, nil
}

func CreateSchedule(s models.Schedule) (int64, error) {
	if s.WeekOf == "" {
		return 0, fmt.Errorf("week_of is required")
	}

	result, err := DB.Exec(
		"INSERT INTO schedule (week_of, notes) VALUES (?, ?)",
		s.WeekOf, s.Notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateSchedule(s models.Schedule) error {
	if s.ID == 0 {
		return fmt.Errorf("id is required")
	}

	_, err := DB.Exec(
		"UPDATE schedule SET notes=? WHERE id=?",
		s.Notes, s.ID,
	)
	return err
}

func DeleteSchedule(id int) error {
	result, err := DB.Exec("DELETE FROM schedule WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("schedule with id %d not found", id)
	}
	return nil
}
