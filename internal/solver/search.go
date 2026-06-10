package solver

import (
	"context"
	"fmt"
	"schedulemate/internal/models"
	"sort"
)

// This file contains the core backtracking search algorithm and its helper functions:
//	- pickNextSlot: MRV heuristic to choose the next slot to assign
//	- isDemandMet: checks if current schedule meets demand constraints
//	- isConsistent: checks if assigning a shift to a slot violates any hard constraints
//				1. Employee does not exceed their weekly max hours
//				2. No 30-min window exceeds EmployeesNeeded+1 (allows one handoff overlap)
//	- canStillCoverDemand: checks if remaining domains can still potentially cover demand. This is important
//    for pruning early when we make an assignment that doesn't immediately violate constraints but dooms us later.
//	- forwardCheck: Meat and potatoes. After assigning a shift to a slot, this prunes the domains of remaining unassigned
//    slots based on the new assignment. Implements three rules:
//				1. Remove options from same-day slots that would cause overstaffing (constraint 2 from isConsistent)
//				2. Remove options for this employee that exceed their remaining weekly hours (constraint 1 from isConsistent)
//				3. Remove options that would leave demand windows uncovered (calls canStillCoverDemand)
//    If a domain becomes empty, we know this branch is doomed and can go back immediately instead of delving deeper.
//	- restorePruned: helper to restore domains after backtracking
//	- lcv: least constraining value heuristic to order a slot's domain options by how many options they would eliminate from other slots
//	- backtrack: the core recursive CSP solver that uses all of the above. It also checks for timeout via context and prints debug info every 10k calls.

// pickNextSlot uses the MRV (minimum remaining values) heuristic:
// choose the unassigned slot with the fewest options left.
func pickNextSlot(slots []slot, domains map[slot][]shiftOption, sched schedule) (slot, bool) {
	var bestSlot slot
	smallestDomain := -1
	found := false
	for _, s := range slots {
		if _, assigned := sched[s]; assigned {
			continue
		}
		domainSize := len(domains[s])
		if smallestDomain == -1 || domainSize < smallestDomain {
			smallestDomain = domainSize
			bestSlot = s
			found = true
		}
	}
	return bestSlot, found
}

// isDemandMet checks whether every demand window is covered by the current schedule.
func isDemandMet(sched schedule) bool {
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
			if coveredBy < ds.EmployeesNeeded {
				return false
			}
		}
	}
	return true
}

// isConsistent checks whether assigning shift to s violates any hard constraints
// given the shifts already in sched.
//
// Constraints checked here (depend on other assignments):
//  1. Employee does not exceed their weekly max hours
//  2. No 30-min window exceeds EmployeesNeeded+1 (allows one handoff overlap)
func isConsistent(s slot, shift shiftOption, sched schedule, employees []models.Employee) bool {
	// Check 1: weekly max hours
	totalMinutes := 0
	for assigned, assignedShift := range sched {
		if assigned.EmployeeID == s.EmployeeID {
			totalMinutes += timeToMinutes(assignedShift.EndTime) - timeToMinutes(assignedShift.StartTime)
		}
	}
	candidateMinutes := timeToMinutes(shift.EndTime) - timeToMinutes(shift.StartTime)
	for _, emp := range employees {
		if emp.ID == s.EmployeeID && totalMinutes+candidateMinutes > int(emp.MaxHours)*60 {
			consistencyFails++
			return false
		}
	}

	// Check 2: staffing cap (allow EmployeesNeeded+1 for handoffs)
	ds := cfg.daySettings[s.DayOfWeek]
	shiftStart := timeToMinutes(shift.StartTime)
	shiftEnd := timeToMinutes(shift.EndTime)
	for t := shiftStart; t < shiftEnd; t += stepMinutes {
		coveredBy := 0
		for assigned, assignedShift := range sched {
			if assigned.DayOfWeek == s.DayOfWeek &&
				timeToMinutes(assignedShift.StartTime) <= t &&
				timeToMinutes(assignedShift.EndTime) >= t+stepMinutes {
				coveredBy++
			}
		}
		if coveredBy >= ds.EmployeesNeeded+1 {
			consistencyFails++
			return false
		}
	}
	return true
}

// canStillCoverDemand checks whether every uncovered demand window can still
// potentially be covered by at least one remaining domain option.
func canStillCoverDemand(domains map[slot][]shiftOption, sched schedule, affectedDays map[int]bool) bool {
	for day := range affectedDays {
		ds := cfg.daySettings[day]
		if ds.EmployeesNeeded == 0 {
			continue
		}
		storeOpen := timeToMinutes(ds.SchedulableOpen)
		storeClose := timeToMinutes(ds.SchedulableClose)

		for t := storeOpen; t < storeClose; t += stepMinutes {
			covered := 0
			for s, shift := range sched {
				if s.DayOfWeek == day &&
					timeToMinutes(shift.StartTime) <= t &&
					timeToMinutes(shift.EndTime) >= t+stepMinutes {
					covered++
				}
			}
			if covered >= ds.EmployeesNeeded {
				continue
			}

			potential := 0
			for otherSlot, options := range domains {
				if otherSlot.DayOfWeek != day {
					continue
				}
				if _, assigned := sched[otherSlot]; assigned {
					continue
				}
				for _, opt := range options {
					if timeToMinutes(opt.StartTime) <= t && timeToMinutes(opt.EndTime) >= t+stepMinutes {
						potential++
						break
					}
				}
			}
			if covered+potential < ds.EmployeesNeeded {
				return false
			}
		}
	}
	return true
}

// forwardCheck prunes domains of unassigned slots after a new assignment.
//
// Rule 1: remove options from same-day slots that would cause overstaffing
// Rule 2: remove options for this employee that exceed their remaining weekly hours
// Rule 3: check demand reachability — if any window becomes uncoverable, fail early
//
// Returns the pruned options map for restoration on backtrack.
// Returns nil to signal a domain wipeout (branch is doomed).
func forwardCheck(
	domains map[slot][]shiftOption,
	sched schedule,
	s slot,
	shift shiftOption,
	employees []models.Employee,
) map[slot][]shiftOption {
	pruned := map[slot][]shiftOption{}
	ds := cfg.daySettings[s.DayOfWeek]

	// Track which days had domains modified — only these need demand rechecking
	affectedDays := map[int]bool{s.DayOfWeek: true}

	// Rule 1: prune overstaffing options on the same day
	for otherSlot, domain := range domains {
		if otherSlot.DayOfWeek != s.DayOfWeek {
			continue
		}
		if _, assigned := sched[otherSlot]; assigned {
			continue
		}

		var remaining []shiftOption
		for _, opt := range domain {
			optStart := timeToMinutes(opt.StartTime)
			optEnd := timeToMinutes(opt.EndTime)
			wouldOverstaff := false
			for t := optStart; t < optEnd; t += stepMinutes {
				coveredBy := 0
				for assigned, assignedShift := range sched {
					if assigned.DayOfWeek == s.DayOfWeek &&
						timeToMinutes(assignedShift.StartTime) <= t &&
						timeToMinutes(assignedShift.EndTime) >= t+stepMinutes {
						coveredBy++
					}
				}
				if coveredBy >= ds.EmployeesNeeded+1 {
					wouldOverstaff = true
					break
				}
			}
			if wouldOverstaff {
				pruned[otherSlot] = append(pruned[otherSlot], opt)
			} else {
				remaining = append(remaining, opt)
			}
		}

		if len(remaining) < len(domain) {
			domains[otherSlot] = remaining
		}
		if len(domains[otherSlot]) == 0 {
			forwardWipeouts++
			restorePruned(domains, pruned)
			return nil
		}
	}

	// Rule 2: prune this employee's future shifts that exceed remaining hours
	totalMinutes := 0
	for assigned, assignedShift := range sched {
		if assigned.EmployeeID == s.EmployeeID {
			totalMinutes += timeToMinutes(assignedShift.EndTime) - timeToMinutes(assignedShift.StartTime)
		}
	}
	var maxHours int
	for _, emp := range employees {
		if emp.ID == s.EmployeeID {
			maxHours = int(emp.MaxHours)
			break
		}
	}
	remainingMinutes := maxHours*60 - totalMinutes

	for otherSlot, options := range domains {
		if otherSlot.EmployeeID != s.EmployeeID {
			continue
		}
		if _, assigned := sched[otherSlot]; assigned {
			continue
		}

		var remaining []shiftOption
		for _, opt := range options {
			if timeToMinutes(opt.EndTime)-timeToMinutes(opt.StartTime) <= remainingMinutes {
				remaining = append(remaining, opt)
			} else {
				pruned[otherSlot] = append(pruned[otherSlot], opt)
			}
		}

		if len(remaining) < len(options) {
			domains[otherSlot] = remaining
			affectedDays[otherSlot.DayOfWeek] = true
		}
		if len(domains[otherSlot]) == 0 {
			forwardWipeouts++
			restorePruned(domains, pruned)
			return nil
		}
	}

	// Rule 3: check demand reachability on affected days only
	if !canStillCoverDemand(domains, sched, affectedDays) {
		forwardWipeouts++
		restorePruned(domains, pruned)
		return nil
	}

	return pruned
}

func restorePruned(domains map[slot][]shiftOption, pruned map[slot][]shiftOption) {
	for s, options := range pruned {
		domains[s] = append(domains[s], options...)
	}
}

// lcv orders a slot's domain by least constraining value — the shift that
// eliminates the fewest options from other slots is tried first.
func lcv(s slot, domains map[slot][]shiftOption, sched schedule, employees []models.Employee) []shiftOption {
	type scored struct {
		shift      shiftOption
		eliminated int
	}

	var results []scored
	for _, shift := range domains[s] {
		sched[s] = shift
		eliminated := 0
		for otherSlot, domain := range domains {
			if otherSlot == s {
				continue
			}
			if _, assigned := sched[otherSlot]; assigned {
				continue
			}
			for _, opt := range domain {
				if !isConsistent(otherSlot, opt, sched, employees) {
					eliminated++
				}
			}
		}
		delete(sched, s)
		results = append(results, scored{shift, eliminated})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].eliminated < results[j].eliminated
	})

	ordered := make([]shiftOption, len(results))
	for i, r := range results {
		ordered[i] = r.shift
	}
	return ordered
}

// backtrack is the core recursive CSP solver.
func backtrack(ctx context.Context, slots []slot, domains map[slot][]shiftOption, sched schedule, employees []models.Employee, minShiftsForDemand int) bool {
	// Cancellation check on every call. Critical: a cancelled context must make
	// the WHOLE search unwind, not just fail one branch. We return false here
	// AND re-check inside the value loop below so each level bails immediately
	// instead of moving on to the next candidate value.
	if ctx.Err() != nil {
		return false
	}

	callCount++
	if callCount%10000 == 0 {
		fmt.Printf("[DEBUG] calls=%d  fails=%d  wipeouts=%d  schedule_size=%d\n",
			callCount, consistencyFails, forwardWipeouts, len(sched))
	}

	if len(sched) >= minShiftsForDemand && isDemandMet(sched) {
		return true
	}

	s, found := pickNextSlot(slots, domains, sched)
	if !found {
		return false
	}

	domainSnapshot := make([]shiftOption, len(domains[s]))
	copy(domainSnapshot, domains[s])

	for _, shift := range domainSnapshot {
		// Hard stop on cancellation: don't keep trying other values.
		if ctx.Err() != nil {
			return false
		}
		if !isConsistent(s, shift, sched, employees) {
			continue
		}
		sched[s] = shift

		pruned := forwardCheck(domains, sched, s, shift, employees)
		if pruned != nil {
			if backtrack(ctx, slots, domains, sched, employees, minShiftsForDemand) {
				return true
			}
			restorePruned(domains, pruned)
		}

		delete(sched, s)
	}
	return false
}
