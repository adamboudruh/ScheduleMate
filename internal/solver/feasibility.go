package solver

import (
	"fmt"
	"schedulemate/internal/models"
)

// feasibilityCheck is a fast pre-pass that catches obviously impossible inputs
// before wasting time in the backtracker.
//
// Check 1: weekly supply vs demand using actual domain output
// Check 2: per-window coverage using actual producible shifts
// Check 3: per-day person-hour supply vs demand (capped at MaxShiftLength per employee)

func feasibilityCheck(employees []models.Employee, availMap map[slot][]models.Availability) bool {
	// Pre-compute each employee's domain once, used by all checks
	allDomains := map[int]map[int][]shiftOption{}
	for _, emp := range employees {
		allDomains[emp.ID] = map[int][]shiftOption{}
		for day := 1; day <= 7; day++ {
			s := slot{EmployeeID: emp.ID, DayOfWeek: day}
			allDomains[emp.ID][day] = generateDomain(s, availMap[s])
		}
	}

	// Check 1: weekly supply vs demand
	weeklyDemandMinutes := 0
	for _, ds := range cfg.daySettings {
		weeklyDemandMinutes += ds.EmployeesNeeded * (timeToMinutes(ds.SchedulableClose) - timeToMinutes(ds.SchedulableOpen))
	}

	weeklySupplyMinutes := 0
	for _, emp := range employees {
		for day := 1; day <= 7; day++ {
			longest := 0
			for _, opt := range allDomains[emp.ID][day] {
				m := timeToMinutes(opt.EndTime) - timeToMinutes(opt.StartTime)
				if m > longest {
					longest = m
				}
			}
			weeklySupplyMinutes += longest
		}
	}

	if weeklySupplyMinutes < weeklyDemandMinutes {
		fmt.Printf("Infeasible: %dh of valid shifts available vs %dh demanded\n",
			weeklySupplyMinutes/60, weeklyDemandMinutes/60)
		return false
	}

	// Check 2: per-window coverage using actual producible shifts
	for day, ds := range cfg.daySettings {
		if ds.EmployeesNeeded == 0 {
			continue
		}
		storeOpen := timeToMinutes(ds.SchedulableOpen)
		storeClose := timeToMinutes(ds.SchedulableClose)
		for t := storeOpen; t < storeClose; t += stepMinutes {
			coveredBy := 0
			for _, emp := range employees {
				for _, opt := range allDomains[emp.ID][day] {
					if timeToMinutes(opt.StartTime) <= t && timeToMinutes(opt.EndTime) >= t+stepMinutes {
						coveredBy++
						break
					}
				}
			}
			if coveredBy < ds.EmployeesNeeded {
				fmt.Printf("Infeasible: day %d at %s covered by %d/%d employees\n",
					day, minutesToTime(t), coveredBy, ds.EmployeesNeeded)
				return false
			}
		}
	}

	// Check 3: per-day person-hour supply vs demand
	maxShiftMinutes := cfg.settings.MaxShiftLength * 60
	for day, ds := range cfg.daySettings {
		if ds.EmployeesNeeded == 0 {
			continue
		}
		demandMinutes := ds.EmployeesNeeded *
			(timeToMinutes(ds.SchedulableClose) - timeToMinutes(ds.SchedulableOpen))

		supplyMinutes := 0
		for _, emp := range employees {
			longest := 0
			for _, opt := range allDomains[emp.ID][day] {
				m := timeToMinutes(opt.EndTime) - timeToMinutes(opt.StartTime)
				if m > longest {
					longest = m
				}
			}
			if longest > maxShiftMinutes {
				longest = maxShiftMinutes
			}
			supplyMinutes += longest
		}

		if supplyMinutes < demandMinutes {
			fmt.Printf("Infeasible: day %d needs %dh of person-coverage but employees can only contribute %dh\n",
				day, demandMinutes/60, supplyMinutes/60)
			return false
		}
	}

	return true
}
