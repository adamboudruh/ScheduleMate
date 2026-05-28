package solver

import (
	"context"
	"fmt"
	"schedulemate/internal/models"
	"time"
)

// This file contains the internal types used by the solver and acts as the entry point
// for the solver package
// types:
//  - shiftOption: a potential shift assignment for an employee on a given day
//  - slot: a specific employee+day combination that we need to assign a shiftOption to
//  - schedule: the current partial assignment of shiftOptions to slots during the search
//  - SolverInput: the input struct the caller (app.go) passes in with all the data needed to run the solver
//  - SolverResult: the output struct the solver returns with the generated schedule and metadata about the search
//  - cfg: package-level config that will hold settings and daySettings
// functions:
//	- Run: single entry point, app.go calls this with SolverInput and gets back SolverResult. Run
//	  sets up the package-level cfg, runs the feasibility pre-check, builds domains, and calls the backtracking search.
//  - resolveSchedulableHours: fills in SchedulableOpen/Close from OpenTime/CloseTime when allow_outside_hours is disabled.
// 	  This way the rest of the solver can always read SchedulableOpen/Close without caring about the setting.
//

// Step size for time-window iteration. 60 = 1 hour, 30 = 30 minutes.
const stepMinutes = 60

// package-level config
var cfg struct {
	settings    models.Settings
	daySettings map[int]models.DaySettings
}

// Solver-internal types

type shiftOption struct {
	StartTime string
	EndTime   string
}

type slot struct {
	EmployeeID int
	DayOfWeek  int
}

type schedule map[slot]shiftOption

// Debug counters
var (
	callCount        int
	consistencyFails int
	forwardWipeouts  int
)

func resetCounters() {
	callCount = 0
	consistencyFails = 0
	forwardWipeouts = 0
}

// Exported types for the caller, stuff that's passed in and out of the solver package

type SolverInput struct {
	Employees      []models.Employee
	Availabilities []models.Availability
	DaySettings    map[int]models.DaySettings
	Settings       models.Settings
}

type ScoreResult struct {
	HoursGap      float64 `json:"hoursGap"`
	Fairness      float64 `json:"fairness"`
	ClopenPenalty float64 `json:"clopenPenalty"`
	Total         float64 `json:"total"`
}

type SolverResult struct {
	Shifts    []models.Shift `json:"shifts"`
	Score     ScoreResult    `json:"score"`
	CallCount int            `json:"callCount"`
	Wipeouts  int            `json:"wipeouts"`
	Elapsed   time.Duration  `json:"elapsed"`
	Feasible  bool           `json:"feasible"`
	Solved    bool           `json:"solved"`
	TimedOut  bool           `json:"timedOut"`
	Message   string         `json:"message"`
}

// Single entry point for the solver. The caller (app.go) loads
// everything from the DB, hands it in as SolverInput, and gets back a
// SolverResult with shifts ready to insert.
//
//  1. app.go calls database.GetAllEmployees(), GetAllAvailabilities(), GetAllDaySettings(), GetSettings()
//  2. app.go passes all of that into solver.Run() through SolverInput
//  3. Run() stores settings in package-level cfg so they're accessible to all internal functions
//  4. Run() returns SolverResult with an array of shifts with ScheduleID as 0
func Run(input SolverInput) SolverResult {
	// store config so all functions can access it
	cfg.settings = input.Settings
	cfg.daySettings = resolveSchedulableHours(input.DaySettings, input.Settings)

	resetCounters()
	start := time.Now()

	availMap := availabilityMap(input.Availabilities)

	// feasibility pre-check
	if !feasibilityCheck(input.Employees, availMap) {
		return SolverResult{
			Feasible: false,
			Solved:   false,
			Elapsed:  time.Since(start),
			Message:  "Schedule is not feasible with current employee availability and demand.",
		}
	}

	// Build domains
	domains := map[slot][]shiftOption{}
	slots := []slot{}
	for _, emp := range input.Employees {
		for day := 1; day <= 7; day++ {
			s := slot{EmployeeID: emp.ID, DayOfWeek: day}
			domain := generateDomain(s, availMap[s])
			if len(domain) > 0 {
				domains[s] = domain
				slots = append(slots, s)
			}
		}
	}

	fmt.Printf("Solver: %d slots, feasibility passed. Searching...\n", len(slots))

	// backtracking search with timeout
	const solveTimeout = 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), solveTimeout)
	defer cancel()

	sched := schedule{}
	solved := backtrack(ctx, slots, domains, sched, input.Employees)

	elapsed := time.Since(start)

	if !solved {
		timedOut := ctx.Err() != nil
		msg := "Search exhausted all possibilities. The dataset is definitively infeasible."
		if timedOut {
			msg = fmt.Sprintf("Search timed out after %s. The dataset is likely infeasible in ways the pre-check didn't catch.", solveTimeout)
		}
		return SolverResult{
			Feasible:  true,
			Solved:    false,
			TimedOut:  timedOut,
			CallCount: callCount,
			Wipeouts:  forwardWipeouts,
			Elapsed:   elapsed,
			Message:   msg,
		}
	}

	// score before optimization
	initialScore := score(sched, input.Employees)
	fmt.Printf("Solver: solution found in %s (calls=%d). Score=%.1f. Optimizing...\n",
		elapsed, callCount, initialScore.Total)

	// optimize
	sched = optimize(sched, input.Employees, availMap, 50)
	finalScore := score(sched, input.Employees)

	fmt.Printf("Solver: optimization done. Score %.1f -> %.1f\n",
		initialScore.Total, finalScore.Total)

	// convert internal schedule to []models.Shift
	shifts := scheduleToShifts(sched)

	return SolverResult{
		Shifts:    shifts,
		Score:     finalScore,
		CallCount: callCount,
		Wipeouts:  forwardWipeouts,
		Elapsed:   time.Since(start),
		Feasible:  true,
		Solved:    true,
		Message:   "Schedule generated and optimized successfully.",
	}
}

// Fills in SchedulableOpen/Close from OpenTime/CloseTime when allow_outside_hours is disabled
// This way the rest of the solver can always read SchedulableOpen/Close without caring about the setting
func resolveSchedulableHours(daySettings map[int]models.DaySettings, settings models.Settings) map[int]models.DaySettings {
	resolved := make(map[int]models.DaySettings, len(daySettings))
	for day, ds := range daySettings {
		if !settings.AllowOutsideHours || ds.SchedulableOpen == "" {
			ds.SchedulableOpen = ds.OpenTime
		}
		if !settings.AllowOutsideHours || ds.SchedulableClose == "" {
			ds.SchedulableClose = ds.CloseTime
		}
		resolved[day] = ds
	}
	return resolved
}

// scheduleToShifts converts the internal schedule map to a slice of models.Shift
// ScheduleID is left as 0, the caller fills it in before writing to DB.
func scheduleToShifts(sched schedule) []models.Shift {
	shifts := make([]models.Shift, 0, len(sched))
	for s, opt := range sched {
		shifts = append(shifts, models.Shift{
			EmployeeID: s.EmployeeID,
			DayOfWeek:  s.DayOfWeek,
			StartTime:  opt.StartTime,
			EndTime:    opt.EndTime,
		})
	}
	return shifts
}
