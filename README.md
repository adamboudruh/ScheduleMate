# ScheduleMate

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Svelte](https://img.shields.io/badge/Svelte-4.x-FF3E00?style=flat&logo=svelte&logoColor=white)](https://svelte.dev/)
[![JavaScript](https://img.shields.io/badge/JavaScript-ES2022-F7DF1E?style=flat&logo=javascript&logoColor=black)](https://developer.mozilla.org/en-US/docs/Web/JavaScript)
[![SQLite](https://img.shields.io/badge/SQLite-3-003B57?style=flat&logo=sqlite&logoColor=white)](https://www.sqlite.org/)

Capstone project for **CS 406 Senior Capstone** at **Oregon State University**

ScheduleMate is a cross-platform desktop application that automates employee shift scheduling. Managers define their employees, availability windows, and store constraints, and ScheduleMate's constraint satisfaction engine generates a valid, optimized schedule in seconds.

**[Watch the demo](https://drive.google.com/file/d/1Eyo_pjbb4FDMJwGXdOSMRj0raKfDd2Qv/view?usp=sharing)**

---

## Features

- **Employee management**: Add employees with roles, hourly wages, desired weekly hours, and maximum weekly hours.
- **Availability tracking**: Record each employee's available windows per day of the week with precise start/end times.
- **Automated schedule generation**: Produce a full weekly schedule via a CSP solver that respects all constraints simultaneously.
- **Configurable store settings**: Set open/close times, minimum staff coverage per day, and optionally allow scheduling outside of core store hours.
- **Shift length enforcement**: Enforce minimum and maximum shift durations (default: 4–8 hours) globally across all generated schedules.
- **Rest period enforcement**: Guarantee a configurable minimum gap between any two shifts for the same employee (default: 10 hours).
- **Per-day staffing targets**: Specify how many employees must be on shift for each individual day of the week.
- **Weekly schedule history**: Schedules are stored by week with optional notes, so past schedules are always accessible.
- **Persistent local storage**: All data lives in a local SQLite database; no account or internet connection required.

---

## How the CSP Works

ScheduleMate models weekly scheduling as a **Constraint Satisfaction Problem (CSP)**, a classical AI technique where a solution must simultaneously satisfy a set of hard constraints.

**Variables:** Each potential shift assignment is a variable; which employee works, on which day, during which time window.

**Domains:** The domain of each variable is narrowed upfront by intersecting the employee's stated availability for that day with the store's schedulable hours, eliminating impossible assignments before search begins.

**Constraints enforced during generation:**

| Constraint | Description |
|---|---|
| Availability | A shift can only be assigned within an employee's declared available window |
| Shift length | Every shift must fall within the configured min/max duration |
| Rest period | An employee cannot be scheduled for another shift until the minimum rest gap has elapsed |
| Hour targets | Each employee's total scheduled hours must not exceed their maximum, and the solver tries to meet their desired hours |
| Minimum coverage | Each open day must have at least the configured number of employees on shift |
| Store hours | Shifts are bounded by the store's open and close times (or the extended schedulable window if outside-hours scheduling is enabled) |

The solver also applies constraint propagation after each assignment to prune the remaining search space before recursing, significantly reducing backtracking.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Desktop framework | [Wails v2](https://wails.io/), compiles Go + web frontend into a single native binary |
| Backend / solver | Go |
| Frontend | Svelte + JavaScript |
| Database | SQLite (via `modernc.org/sqlite`, pure Go, no CGO required) |

---

## Getting Started

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Run in development

```bash
git clone https://github.com/yourusername/schedulemate.git
cd schedulemate
wails dev
```

### Build a native binary

```bash
wails build
```

The output is a self-contained executable.

---

## Project Context

Built as a capstone project for **CS 406** at **Oregon State University**. The goal was to apply concepts from algorithms, databases, and software engineering to a real-world scheduling problem, a domain where manual approaches are error-prone and time-consuming for small business operators.
