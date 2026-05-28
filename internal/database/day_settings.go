package database

import (
	"fmt"
	"schedulemate/internal/models"
)

func GetAllDaySettings() (map[int]models.DaySettings, error) {
	rows, err := DB.Query(`
        SELECT day_of_week, open_time, close_time,
               COALESCE(schedulable_open, ''), COALESCE(schedulable_close, ''),
               employees_needed
        FROM day_settings
        ORDER BY day_of_week`,
	) // coalesce schedulable times to empty string for easier handling
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int]models.DaySettings{} // return a map of 7 entries keyed by day_of_week
	for rows.Next() {
		var ds models.DaySettings
		if err := rows.Scan(
			&ds.DayOfWeek, &ds.OpenTime, &ds.CloseTime,
			&ds.SchedulableOpen, &ds.SchedulableClose,
			&ds.EmployeesNeeded,
		); err != nil {
			return nil, err
		}
		result[ds.DayOfWeek] = ds
	}
	return result, nil
}

func GetDaySettings(dayOfWeek int) (models.DaySettings, error) {
	var ds models.DaySettings
	err := DB.QueryRow(`
        SELECT day_of_week, open_time, close_time,
               COALESCE(schedulable_open, ''), COALESCE(schedulable_close, ''),
               employees_needed
        FROM day_settings WHERE day_of_week = ?`, dayOfWeek,
	).Scan(
		&ds.DayOfWeek, &ds.OpenTime, &ds.CloseTime,
		&ds.SchedulableOpen, &ds.SchedulableClose,
		&ds.EmployeesNeeded,
	)
	if err != nil {
		return models.DaySettings{}, fmt.Errorf("day settings for day %d not found", dayOfWeek)
	}
	return ds, nil
}

func UpdateDaySettings(ds models.DaySettings) error {
	if ds.DayOfWeek < 1 || ds.DayOfWeek > 7 {
		return fmt.Errorf("day_of_week must be between 1 and 7")
	}

	var schedulableOpen, schedulableClose = ds.OpenTime, ds.CloseTime // default to open/close if schedulable times are not provided
	if ds.SchedulableOpen != "" {
		schedulableOpen = ds.SchedulableOpen
	}
	if ds.SchedulableClose != "" {
		schedulableClose = ds.SchedulableClose
	}

	_, err := DB.Exec(`
        UPDATE day_settings
        SET open_time=?, close_time=?, schedulable_open=?, schedulable_close=?, employees_needed=?
        WHERE day_of_week=?`,
		ds.OpenTime, ds.CloseTime, schedulableOpen, schedulableClose,
		ds.EmployeesNeeded, ds.DayOfWeek,
	)
	return err
}
