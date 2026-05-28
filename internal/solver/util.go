package solver

import (
	"fmt"
	"strconv"
)

// This file contains some general utility functions used across the solver package
// functions:
//	- timeToMinutes: converts "HH:MM" to total minutes like "01:30" -> 90
//	- minutesToTime: converts total minutes back to "HH:MM" like 90 -> "01:30"
//

func timeToMinutes(t string) int {
	hours, _ := strconv.Atoi(t[0:2])
	minutes, _ := strconv.Atoi(t[3:5])
	return hours*60 + minutes
}

func minutesToTime(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}
