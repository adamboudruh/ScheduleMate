package database

import "schedulemate/internal/models"

func GetSettings() (models.Settings, error) {
	var s models.Settings
	err := DB.QueryRow(`
        SELECT id, min_shift_length, max_shift_length, allow_outside_hours, time_between_shifts
        FROM settings WHERE id = 1`,
	).Scan(&s.ID, &s.MinShiftLength, &s.MaxShiftLength, &s.AllowOutsideHours, &s.TimeBetweenShifts)
	if err != nil {
		return models.Settings{}, err
	}
	return s, nil
}

func UpdateSettings(s models.Settings) error {
	_, err := DB.Exec(`
        UPDATE settings
        SET min_shift_length=?, max_shift_length=?, allow_outside_hours=?, time_between_shifts=?
        WHERE id = 1`,
		s.MinShiftLength, s.MaxShiftLength, s.AllowOutsideHours, s.TimeBetweenShifts,
	)
	return err
}
