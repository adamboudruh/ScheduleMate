package solver

import (
	"fmt"
	"math"
	"schedulemate/internal/models"
)

// This file scores the schedule based on soft constraints and tries to optimize it with local search moves
// functions:
//	- score: calculates the score of a complete schedule based on hours gap, fairness, and clopen penalty
//	- isValidSchedule: checks all hard constraints on a complete schedule. Used by the optimizer to verify that a move didn't break anything.
//	- optimize: runs local search on a valid schedule to improve its soft score. Three move types: adjust shift length, swap shifts, slide shifts.
//        - adjust shift length: extend or shrink by stepMinutes at start or end to better match desired hours
//        - swap shifts: swap shifts between two employees on the same day to improve fairness
//        - slide shifts: move a shift earlier or later within the same day to better match desired hours
//     After trying each of these moves, optimizer checks if schedue is still valid and if the score improved. If it
//     did, it keeps the change and continues optimizing.

const ( // clopen is the most important soft constraint to optimize for, then fairness, then hours gap
	weightHoursGap  = 1.0
	weightFairness  = 3.0
	weightClopen    = 2.0
	weightOverstaff = 4.0
)

// employeeWeeklyMinutes returns a map of employeeID -> total scheduled minutes
func employeeWeeklyMinutes(sched schedule) map[int]int {
	totals := map[int]int{}
	for s, shift := range sched {
		totals[s.EmployeeID] += timeToMinutes(shift.EndTime) - timeToMinutes(shift.StartTime)
	}
	return totals
}

func score(sched schedule, employees []models.Employee) ScoreResult {
	totals := employeeWeeklyMinutes(sched)

	// Hours gap: sum of |actual - desired| across all employees
	hoursGap := 0.0
	gaps := map[int]float64{}
	for _, emp := range employees {
		actual := float64(totals[emp.ID]) / 60.0
		desired := emp.DesiredHours
		gap := math.Abs(actual - (float64)(desired))
		hoursGap += gap
		gaps[emp.ID] = gap
	}

	// Fairness: max(gap) - min(gap) across employees
	maxGap := 0.0
	minGap := math.MaxFloat64
	for _, g := range gaps {
		if g > maxGap {
			maxGap = g
		}
		if g < minGap {
			minGap = g
		}
	}
	fairness := maxGap - minGap

	// Clopen penalty: penalize short gaps between consecutive-day shifts
	timeBetweenShifts := float64(cfg.settings.TimeBetweenShifts)
	clopenPenalty := 0.0
	for _, emp := range employees {
		for day := 1; day <= 6; day++ {
			slotToday := slot{EmployeeID: emp.ID, DayOfWeek: day}
			slotTomorrow := slot{EmployeeID: emp.ID, DayOfWeek: day + 1}
			shiftToday, hasToday := sched[slotToday]
			shiftTomorrow, hasTomorrow := sched[slotTomorrow]
			if !hasToday || !hasTomorrow {
				continue
			}
			endToday := float64(timeToMinutes(shiftToday.EndTime)) / 60.0
			startTomorrow := float64(timeToMinutes(shiftTomorrow.StartTime)) / 60.0
			gap := (24.0 - endToday) + startTomorrow
			if gap < timeBetweenShifts {
				clopenPenalty += timeBetweenShifts - gap
			}
		}
	}

	// Overstaffing penalty: sum of excess coverage across all time windows
	overstaffPenalty := 0.0
	for day, ds := range cfg.daySettings {
		if ds.EmployeesNeeded == 0 {
			continue
		}
		storeOpen := timeToMinutes(ds.SchedulableOpen)
		storeClose := timeToMinutes(ds.SchedulableClose)
		for t := storeOpen; t < storeClose; t += stepMinutes {
			coveredBy := 0
			for s, shift := range sched {
				if s.DayOfWeek == day &&
					timeToMinutes(shift.StartTime) <= t &&
					timeToMinutes(shift.EndTime) >= t+stepMinutes {
					coveredBy++
				}
			}
			if coveredBy > ds.EmployeesNeeded {
				overstaffPenalty += float64(coveredBy - ds.EmployeesNeeded)
			}
		}
	}

	total := weightHoursGap*hoursGap + weightFairness*fairness + weightClopen*clopenPenalty + weightOverstaff*overstaffPenalty

	return ScoreResult{
		HoursGap:      hoursGap,
		Fairness:      fairness,
		ClopenPenalty: clopenPenalty,
		Total:         total,
	}
}

// isValidSchedule checks all hard constraints on a complete schedule.
// Used by the optimizer to verify that a move didn't break anything.
func isValidSchedule(sched schedule, employees []models.Employee, availMap map[slot][]models.Availability) bool {
	if !isDemandMet(sched) {
		return false
	}

	for s, shift := range sched {
		shiftStart := timeToMinutes(shift.StartTime)
		shiftEnd := timeToMinutes(shift.EndTime)
		shiftLen := shiftEnd - shiftStart

		// Shift length within bounds
		if shiftLen < cfg.settings.MinShiftLength*60 || shiftLen > cfg.settings.MaxShiftLength*60 {
			return false
		}

		// Shift within store hours
		ds := cfg.daySettings[s.DayOfWeek]
		if shiftStart < timeToMinutes(ds.SchedulableOpen) || shiftEnd > timeToMinutes(ds.SchedulableClose) {
			return false
		}

		// Shift within employee availability
		avails := availMap[s]
		withinAvail := false
		for _, avail := range avails {
			if shiftStart >= timeToMinutes(avail.StartTime) && shiftEnd <= timeToMinutes(avail.EndTime) {
				withinAvail = true
				break
			}
		}
		if !withinAvail {
			return false
		}
	}

	// No employee exceeds max hours
	totals := employeeWeeklyMinutes(sched)
	for _, emp := range employees {
		if totals[emp.ID] > int(emp.MaxHours)*60 {
			return false
		}
	}

	return true
}

// optimize runs local search on a valid schedule to improve its soft score.
// Three move types: adjust shift length, swap shifts, slide shifts.
func optimize(sched schedule, employees []models.Employee, availMap map[slot][]models.Availability, maxIterations int) schedule {
	bestScore := score(sched, employees)
	movesKept := 0

	for iter := 0; iter < maxIterations; iter++ {
		improved := false

		// --- Move 1: Adjust shift length (extend or shrink by stepMinutes at either end) ---
		for s, shift := range sched {
			original := shift
			shiftStart := timeToMinutes(shift.StartTime)
			shiftEnd := timeToMinutes(shift.EndTime)

			candidates := []shiftOption{
				{minutesToTime(shiftStart - stepMinutes), shift.EndTime},
				{minutesToTime(shiftStart + stepMinutes), shift.EndTime},
				{shift.StartTime, minutesToTime(shiftEnd + stepMinutes)},
				{shift.StartTime, minutesToTime(shiftEnd - stepMinutes)},
			}

			for _, candidate := range candidates {
				sched[s] = candidate
				if isValidSchedule(sched, employees, availMap) {
					newScore := score(sched, employees)
					if newScore.Total < bestScore.Total {
						movesKept++
						fmt.Printf("  [OPT #%d] adjust emp=%d day=%d: %s-%s -> %s-%s (score %.1f -> %.1f)\n",
							movesKept, s.EmployeeID, s.DayOfWeek,
							original.StartTime, original.EndTime,
							candidate.StartTime, candidate.EndTime,
							bestScore.Total, newScore.Total)
						bestScore = newScore
						improved = true
						break
					}
				}
				sched[s] = original
			}
		}

		// --- Move 2: Swap shifts between two employees on the same day ---
		for slot1, shift1 := range sched {
			for slot2, shift2 := range sched {
				if slot1.DayOfWeek != slot2.DayOfWeek {
					continue
				}
				if slot1.EmployeeID >= slot2.EmployeeID {
					continue
				}
				sched[slot1] = shift2
				sched[slot2] = shift1
				if isValidSchedule(sched, employees, availMap) {
					newScore := score(sched, employees)
					if newScore.Total < bestScore.Total {
						movesKept++
						fmt.Printf("  [OPT #%d] swap day=%d: emp %d (%s-%s) <-> emp %d (%s-%s) (score %.1f -> %.1f)\n",
							movesKept, slot1.DayOfWeek,
							slot1.EmployeeID, shift1.StartTime, shift1.EndTime,
							slot2.EmployeeID, shift2.StartTime, shift2.EndTime,
							bestScore.Total, newScore.Total)
						bestScore = newScore
						improved = true
						continue
					}
				}
				sched[slot1] = shift1
				sched[slot2] = shift2
			}
		}

		// --- Move 3: Slide shift earlier or later (same length) ---
		// Runs before remove/add so overlapping shifts get repositioned rather than cut.
		for s, shift := range sched {
			original := shift
			shiftStart := timeToMinutes(shift.StartTime)
			shiftEnd := timeToMinutes(shift.EndTime)

			candidates := []shiftOption{
				{minutesToTime(shiftStart - stepMinutes), minutesToTime(shiftEnd - stepMinutes)},
				{minutesToTime(shiftStart + stepMinutes), minutesToTime(shiftEnd + stepMinutes)},
			}

			for _, candidate := range candidates {
				candStart := timeToMinutes(candidate.StartTime)
				candEnd := timeToMinutes(candidate.EndTime)
				if candStart < 0 || candEnd > 24*60 {
					continue
				}
				sched[s] = candidate
				if isValidSchedule(sched, employees, availMap) {
					newScore := score(sched, employees)
					if newScore.Total < bestScore.Total {
						movesKept++
						fmt.Printf("  [OPT #%d] slide emp=%d day=%d: %s-%s -> %s-%s (score %.1f -> %.1f)\n",
							movesKept, s.EmployeeID, s.DayOfWeek,
							original.StartTime, original.EndTime,
							candidate.StartTime, candidate.EndTime,
							bestScore.Total, newScore.Total)
						bestScore = newScore
						improved = true
						break
					}
				}
				sched[s] = original
			}
		}

		// --- Move 4: Remove a shift entirely ---
		// No desired-hours guard — the score decides. Move 5 can add hours back
		// in a better position, making remove+add a relocation rather than a cut.
		for slotToRemove, shiftToRemove := range sched {
			delete(sched, slotToRemove)
			if isDemandMet(sched) {
				newScore := score(sched, employees)
				if newScore.Total < bestScore.Total {
					movesKept++
					fmt.Printf("  [OPT #%d] remove emp=%d day=%d: %s-%s (score %.1f -> %.1f)\n",
						movesKept, slotToRemove.EmployeeID, slotToRemove.DayOfWeek,
						shiftToRemove.StartTime, shiftToRemove.EndTime,
						bestScore.Total, newScore.Total)
					bestScore = newScore
					improved = true
					break
				}
			}
			sched[slotToRemove] = shiftToRemove
		}

		// --- Move 5: Add a shift for an under-scheduled employee ---
		// Makes the optimizer bidirectional — combined with Move 4, hours can be
		// relocated (removed from a bad position, added in a good one).
		totals := employeeWeeklyMinutes(sched)
	addLoop:
		for _, emp := range employees {
			if float64(totals[emp.ID])/60.0 >= (float64)(emp.DesiredHours) {
				continue
			}
			for day := 1; day <= 7; day++ {
				addSlot := slot{EmployeeID: emp.ID, DayOfWeek: day}
				if _, assigned := sched[addSlot]; assigned {
					continue
				}
				for _, opt := range generateDomain(addSlot, availMap[addSlot]) {
					sched[addSlot] = opt
					if isValidSchedule(sched, employees, availMap) {
						newScore := score(sched, employees)
						if newScore.Total < bestScore.Total {
							movesKept++
							fmt.Printf("  [OPT #%d] add emp=%d day=%d: %s-%s (score %.1f -> %.1f)\n",
								movesKept, emp.ID, day,
								opt.StartTime, opt.EndTime,
								bestScore.Total, newScore.Total)
							bestScore = newScore
							improved = true
							break addLoop
						}
					}
					delete(sched, addSlot)
				}
			}
		}

		if !improved {
			fmt.Printf("  Local optimum reached after %d iterations, %d moves kept.\n", iter+1, movesKept)
			break
		}
	}

	return sched
}

// Consolidate is a post-optimization step that merges fragmented shift pairs into longer single
// shifts provided they don't violdate any constrictions. Runs until no more consolidations are possible.
func consolidate(sched schedule, employees []models.Employee, availMap map[slot][]models.Availability) schedule {
	for {
		consolidated := false

		for _, emp1 := range employees {
			for _, emp2 := range employees {
				if emp1.ID >= emp2.ID {
					continue
				}

				type sharedDay struct {
					day         int
					shift1      shiftOption
					shift2      shiftOption
					mergedStart string
					mergedEnd   string
				}
				var sharedDays []sharedDay

				for day := 1; day <= 7; day++ {
					s1 := slot{EmployeeID: emp1.ID, DayOfWeek: day}
					s2 := slot{EmployeeID: emp2.ID, DayOfWeek: day}
					shift1, ok1 := sched[s1]
					shift2, ok2 := sched[s2]
					if !ok1 || !ok2 {
						continue
					}

					len1 := timeToMinutes(shift1.EndTime) - timeToMinutes(shift1.StartTime)
					len2 := timeToMinutes(shift2.EndTime) - timeToMinutes(shift2.StartTime)
					if len1 != 4*60 || len2 != 4*60 {
						continue
					}

					// Shifts must tile: one ends exactly where the other begins.
					if shift1.EndTime != shift2.StartTime && shift2.EndTime != shift1.StartTime {
						continue
					}

					mergedStart := shift1.StartTime
					if timeToMinutes(shift2.StartTime) < timeToMinutes(shift1.StartTime) {
						mergedStart = shift2.StartTime
					}
					mergedEnd := shift1.EndTime
					if timeToMinutes(shift2.EndTime) > timeToMinutes(shift1.EndTime) {
						mergedEnd = shift2.EndTime
					}

					sharedDays = append(sharedDays, sharedDay{day, shift1, shift2, mergedStart, mergedEnd})
				}

				if len(sharedDays) < 2 {
					continue
				}

				for di := 0; di < len(sharedDays); di++ {
					for dj := di + 1; dj < len(sharedDays); dj++ {
						dayA, dayB := sharedDays[di], sharedDays[dj]
						slotE1A := slot{EmployeeID: emp1.ID, DayOfWeek: dayA.day}
						slotE2A := slot{EmployeeID: emp2.ID, DayOfWeek: dayA.day}
						slotE1B := slot{EmployeeID: emp1.ID, DayOfWeek: dayB.day}
						slotE2B := slot{EmployeeID: emp2.ID, DayOfWeek: dayB.day}
						mergedA := shiftOption{dayA.mergedStart, dayA.mergedEnd}
						mergedB := shiftOption{dayB.mergedStart, dayB.mergedEnd}

						for _, opt := range []struct{ e1Day, e2Day int }{
							{dayA.day, dayB.day},
							{dayB.day, dayA.day},
						} {
							var e1Merged, e2Merged shiftOption
							var e1Keep, e1Remove, e2Keep, e2Remove slot
							if opt.e1Day == dayA.day {
								e1Merged, e2Merged = mergedA, mergedB
								e1Keep, e1Remove = slotE1A, slotE1B
								e2Keep, e2Remove = slotE2B, slotE2A
							} else {
								e1Merged, e2Merged = mergedB, mergedA
								e1Keep, e1Remove = slotE1B, slotE1A
								e2Keep, e2Remove = slotE2A, slotE2B
							}

							sched[e1Keep] = e1Merged
							delete(sched, e1Remove)
							sched[e2Keep] = e2Merged
							delete(sched, e2Remove)

							if isValidSchedule(sched, employees, availMap) {
								fmt.Printf("  [CONSOLIDATE] emp%d→day%d emp%d→day%d\n",
									emp1.ID, opt.e1Day, emp2.ID, opt.e2Day)
								consolidated = true
								goto nextConsolidation
							}

							// Revert
							sched[slotE1A] = dayA.shift1
							sched[slotE1B] = dayB.shift1
							sched[slotE2A] = dayA.shift2
							sched[slotE2B] = dayB.shift2
						}
					}
				}
			}
		}

	nextConsolidation:
		if !consolidated {
			break
		}
	}
	return sched
}
