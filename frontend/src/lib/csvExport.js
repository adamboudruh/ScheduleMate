// csvExport.js
// Builds a CSV for one week's schedule and triggers a download.

const DAY_LONG = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

function tmin(t) {
  const [h, m] = t.split(':').map(Number);
  return h * 60 + m;
}

// CSV-escape a single field.
function esc(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s;
}

// 'YYYY-MM-DD' -> 'ScheduleMate-MM-DD-YYYY.csv'
export function csvFilename(weekOf) {
  const [y, m, d] = weekOf.split('-');
  return `ScheduleMate-${m}-${d}-${y}.csv`;
}

// GRID layout: one row per employee, one column per day. most printable format
export function toGridCSV(shifts, employees) {
  const header = ['Employee', 'Role', 'Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const byEmpDay = {}; // map of shifts by employee day for when building rows
  for (const s of shifts) {
    (byEmpDay[s.employeeId] ??= {});
    (byEmpDay[s.employeeId][s.dayOfWeek] ??= []).push(`${s.startTime}-${s.endTime}`);
  }
  const rows = [header];
  for (const e of employees) {
    const row = [e.name, e.role || ''];
    for (let d = 1; d <= 7; d++) {
      row.push((byEmpDay[e.id]?.[d] || []).join(' / '));
    }
    rows.push(row);
  }
  return rows.map(r => r.map(esc).join(',')).join('\n');
}

export function downloadCSV(filename, csv) {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}