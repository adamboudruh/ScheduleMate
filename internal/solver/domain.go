package solver

import "schedulemate/internal/models"

// This file contains core domain logic of the solver.
// Functions:
//	- availabilityMap: returns a map of each slot to the employee's availability entries for that day.
//	- generateDomain: returns all valid shifts for an employee on a given day, intersecting their availability
//    with store hours and shift-length constraints. "00:00"–"00:00" availability entries (day-off encoding) produce an empty domain.

// availabilityMap returns a map of each slot to the employee's availability entries for that day.
func availabilityMap(avails []models.Availability) map[slot][]models.Availability {
	m := map[slot][]models.Availability{}
	for _, a := range avails {
		key := slot{EmployeeID: a.EmployeeID, DayOfWeek: a.DayOfWeek}
		m[key] = append(m[key], a)
	}
	return m
}

// generateDomain returns all valid shifts for an employee on a given day,
// intersecting their availability with store hours and shift-length constraints.
// "00:00"–"00:00" availability entries (day-off encoding) produce an empty domain.
func generateDomain(s slot, avails []models.Availability) []shiftOption {
	ds := cfg.daySettings[s.DayOfWeek]
	storeOpen := timeToMinutes(ds.SchedulableOpen)
	storeClose := timeToMinutes(ds.SchedulableClose)

	// Store is closed this day — no shifts possible
	if storeOpen == 0 && storeClose == 0 {
		return nil
	}

	minShift := cfg.settings.MinShiftLength * 60
	maxShift := cfg.settings.MaxShiftLength * 60

	var domain []shiftOption
	for _, avail := range avails {
		availStart := timeToMinutes(avail.StartTime)
		availEnd := timeToMinutes(avail.EndTime)

		// "00:00"–"00:00" means unavailable
		if availStart == 0 && availEnd == 0 {
			continue
		}

		windowStart := max(availStart, storeOpen)
		windowEnd := min(availEnd, storeClose)

		for start := windowStart; start < windowEnd; start += stepMinutes {
			for end := start + minShift; end <= start+maxShift && end <= windowEnd; end += stepMinutes {
				domain = append(domain, shiftOption{minutesToTime(start), minutesToTime(end)})
			}
		}
	}
	return domain
}
