// scheduleStats.js
// Frontend computations for the schedule Stats modal: payroll, soft-score
// breakdown, and validity.
//
// computeScore mirrors the weights + formulas in scoring.go,
// and computeValidity mirrors isValidSchedule / isDemandMet. These are
// duplicated here so the modal works without a round-trip
//
// Note to self, there could be some drift if one side's logic is changed
// and the other's isn't, so keep that in mind

const STEP = 60;
const DAY_LONG = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

// Mirrors scoring.go weights.
const WEIGHTS = { hoursGap: 1.0, fairness: 3.0, clopen: 2.0, overstaff: 4.0 };

function tmin(t) {
  if (!t) return 0;
  const [h, m] = t.split(':').map(Number);
  return h * 60 + m;
}

function fmt(min) {
  const h = Math.floor(min / 60), m = min % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

// Resolve each day's effective schedulable window, mirroring go resolveSchedulableHours
function resolveDays(daySettings, allowOutsideHours) {
  const out = {};
  for (let d = 1; d <= 7; d++) {
    const ds = daySettings[d];
    if (!ds) { out[d] = null; continue; }
    let so = ds.schedulableOpen, sc = ds.schedulableClose;
    if (!allowOutsideHours || !so) so = ds.openTime;
    if (!allowOutsideHours || !sc) sc = ds.closeTime;
    out[d] = {
      closed: ds.employeesNeeded === 0,
      needed: ds.employeesNeeded,
      openMin: tmin(ds.openTime),
      closeMin: tmin(ds.closeTime),
      schedOpenMin: tmin(so),
      schedCloseMin: tmin(sc),
    };
  }
  return out;
}

function weeklyMinutes(shifts) {
  const m = {};
  for (const s of shifts) {
    m[s.employeeId] = (m[s.employeeId] || 0) + (tmin(s.endTime) - tmin(s.startTime));
  }
  return m;
}

// Does shift cover the t -> t+STEP window? same logic as solver's coversWindow
function coversWindow(s, day, t) {
  return s.dayOfWeek === day && tmin(s.startTime) <= t && tmin(s.endTime) >= t + STEP;
}

// calculate payroll
export function computePayroll(shifts, employees) {
  const empMap = Object.fromEntries(employees.map(e => [e.id, e]));
  let total = 0;
  const perEmp = {};
  for (const s of shifts) {
    const e = empMap[s.employeeId];
    const hrs = (tmin(s.endTime) - tmin(s.startTime)) / 60;
    const wage = e?.wage ?? 0;
    const cost = hrs * wage;
    total += cost;
    if (!perEmp[s.employeeId]) {
      perEmp[s.employeeId] = { name: e?.name ?? `#${s.employeeId}`, hours: 0, cost: 0, wage };
    }
    perEmp[s.employeeId].hours += hrs;
    perEmp[s.employeeId].cost += cost;
  }
  return { total, perEmployee: Object.values(perEmp) };
}

// soft score (mirrors scoring.go score())
export function computeScore(shifts, employees, daySettings, settings) {
  const resolved = resolveDays(daySettings, settings.allowOutsideHours);
  const totals = weeklyMinutes(shifts);

  let hoursGap = 0;
  const gaps = {};
  for (const e of employees) {
    const actual = (totals[e.id] || 0) / 60;
    const gap = Math.abs(actual - e.desiredHours);
    hoursGap += gap;
    gaps[e.id] = gap;
  }

  // Fairness: spread of the per-employee gaps (max − min).
  let maxGap = 0, minGap = Infinity;
  for (const e of employees) {
    const g = gaps[e.id];
    if (g > maxGap) maxGap = g;
    if (g < minGap) minGap = g;
  }
  if (!isFinite(minGap)) minGap = 0;
  const fairness = maxGap - minGap;

  // Clopen: penalize <timeBetweenShifts rest between consecutive-day shifts.
  // Like the solver, this assumes at most one shift per employee per day
  // (keyed by emp+day); a manual 2nd shift on a day is overwritten here.
  const byEmpDay = {};
  for (const s of shifts) ((byEmpDay[s.employeeId] ??= {})[s.dayOfWeek] = s);
  let clopen = 0;
  const between = settings.timeBetweenShifts;
  for (const e of employees) {
    for (let day = 1; day <= 6; day++) {
      const t1 = byEmpDay[e.id]?.[day];
      const t2 = byEmpDay[e.id]?.[day + 1];
      if (!t1 || !t2) continue;
      const endToday = tmin(t1.endTime) / 60;
      const startTomorrow = tmin(t2.startTime) / 60;
      const gap = (24 - endToday) + startTomorrow;
      if (gap < between) clopen += between - gap;
    }
  }

  // Overstaff: Σ excess coverage over employees-needed across all windows.
  let overstaff = 0;
  for (let day = 1; day <= 7; day++) {
    const r = resolved[day];
    if (!r || r.closed) continue;
    for (let t = r.schedOpenMin; t < r.schedCloseMin; t += STEP) {
      let covered = 0;
      for (const s of shifts) if (coversWindow(s, day, t)) covered++;
      if (covered > r.needed) overstaff += covered - r.needed;
    }
  }

  const total =
    WEIGHTS.hoursGap * hoursGap +
    WEIGHTS.fairness * fairness +
    WEIGHTS.clopen * clopen +
    WEIGHTS.overstaff * overstaff;

  const perEmployeeGap = employees.map(e => ({
    name: e.name,
    actual: (totals[e.id] || 0) / 60,
    desired: e.desiredHours,
    gap: gaps[e.id],
  }));

  return {
    hoursGap, fairness, clopenPenalty: clopen, overstaffPenalty: overstaff, total,
    weights: WEIGHTS, perEmployeeGap,
  };
}

// mirrors isValidSchedule / isDemandMet)
// Returns a per-category pass/fail plus human-readable gap & issue lists.
export function computeValidity(shifts, employees, daySettings, settings, availabilities) {
  const resolved = resolveDays(daySettings, settings.allowOutsideHours);
  const empMap = Object.fromEntries(employees.map(e => [e.id, e]));

  // availability: emp -> day -> [rows]
  const availMap = {};
  for (const a of (availabilities || [])) {
    ((availMap[a.employeeId] ??= {})[a.dayOfWeek] ??= []).push(a);
  }

  const gaps = [];   // coverage shortfalls (the headline "valid" check)
  const issues = []; // other hard-constraint violations

  let coverage = true, shiftLength = true, maxHours = true, availability = true, storeHours = true;
  const minLen = settings.minShiftLength * 60;
  const maxLen = settings.maxShiftLength * 60;

  // Coverage / demand met
  for (let day = 1; day <= 7; day++) {
    const r = resolved[day];
    if (!r || r.closed) continue;
    for (let t = r.schedOpenMin; t < r.schedCloseMin; t += STEP) {
      let covered = 0;
      for (const s of shifts) if (coversWindow(s, day, t)) covered++;
      if (covered < r.needed) {
        coverage = false;
        gaps.push(`${DAY_LONG[day]} ${fmt(t)}–${fmt(t + STEP)}: ${covered}/${r.needed} staffed`);
      }
    }
  }

  // checks per shift
  for (const s of shifts) {
    const name = empMap[s.employeeId]?.name ?? `#${s.employeeId}`;
    const len = tmin(s.endTime) - tmin(s.startTime);
    if (len < minLen || len > maxLen) {
      shiftLength = false;
      issues.push(`${name}: ${DAY_LONG[s.dayOfWeek]} shift is ${len / 60}h (allowed ${settings.minShiftLength}–${settings.maxShiftLength}h)`);
    }
    const r = resolved[s.dayOfWeek];
    if (r && !r.closed && (tmin(s.startTime) < r.schedOpenMin || tmin(s.endTime) > r.schedCloseMin)) {
      storeHours = false;
      issues.push(`${name}: ${DAY_LONG[s.dayOfWeek]} shift falls outside schedulable hours`);
    }
    const avs = availMap[s.employeeId]?.[s.dayOfWeek] || [];
    const within = avs.some(a => tmin(s.startTime) >= tmin(a.startTime) && tmin(s.endTime) <= tmin(a.endTime));
    if (!within) {
      availability = false;
      issues.push(`${name}: ${DAY_LONG[s.dayOfWeek]} shift is outside their availability`);
    }
  }

  // weekly max hours
  const totals = weeklyMinutes(shifts);
  for (const e of employees) {
    if ((totals[e.id] || 0) > e.maxHours * 60) {
      maxHours = false;
      issues.push(`${e.name}: ${((totals[e.id] || 0) / 60).toFixed(1)}h exceeds max ${e.maxHours}h`);
    }
  }

  const categories = { coverage, shiftLength, maxHours, availability, storeHours };
  const valid = Object.values(categories).every(Boolean);
  return { valid, categories, gaps, issues };
}