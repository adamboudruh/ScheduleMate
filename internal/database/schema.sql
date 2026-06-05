CREATE TABLE IF NOT EXISTS employee (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    role          TEXT,
    desired_hours INTEGER NOT NULL,
    max_hours     INTEGER NOT NULL,
    wage          REAL
);

CREATE TABLE IF NOT EXISTS availability (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    employee_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,  -- 1=Sunday, 7=Saturday
    start_time  TEXT NOT NULL,
    end_time    TEXT NOT NULL,
    FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schedule (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    week_of TEXT NOT NULL UNIQUE,
    notes   TEXT
);

CREATE TABLE IF NOT EXISTS shift (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    employee_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,
    start_time  TEXT NOT NULL,
    end_time    TEXT NOT NULL,
    FOREIGN KEY (schedule_id) REFERENCES schedule(id) ON DELETE CASCADE,
    FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE
);

-- replaces store_hours
CREATE TABLE IF NOT EXISTS day_settings (
    day_of_week          INTEGER PRIMARY KEY,
    open_time            TEXT NOT NULL,
    close_time           TEXT NOT NULL,
    schedulable_open     TEXT, -- nullable, used when allow_outside_hours = true
    schedulable_close    TEXT, -- nullable, used when allow_outside_hours = true
    employees_needed     INTEGER NOT NULL DEFAULT 2
);

-- always exactly one row
CREATE TABLE IF NOT EXISTS settings (
    id                      INTEGER PRIMARY KEY DEFAULT 1,
    min_shift_length        INTEGER NOT NULL DEFAULT 4, -- hours
    max_shift_length        INTEGER NOT NULL DEFAULT 8, -- hours
    allow_outside_hours     INTEGER NOT NULL DEFAULT 0, -- boolean
    time_between_shifts     INTEGER NOT NULL DEFAULT 10  -- hours
);

-- Insert default settings if they don't exist

INSERT OR IGNORE INTO settings (id, min_shift_length, max_shift_length, allow_outside_hours, time_between_shifts)
VALUES (1, 4, 8, 0, 10); -- default settings: 4-8 hour shifts, no scheduling outside of open hours, 10 hours between shifts

INSERT OR IGNORE INTO day_settings (day_of_week, open_time, close_time, employees_needed)
VALUES
    (1, '09:00', '21:00', 2),
    (2, '09:00', '21:00', 2),
    (3, '09:00', '21:00', 2),
    (4, '09:00', '21:00', 2),
    (5, '09:00', '21:00', 2),
    (6, '09:00', '21:00', 2),
    (7, '09:00', '21:00', 2); -- default day settings: 9 AM to 9 PM, 2 employees needed