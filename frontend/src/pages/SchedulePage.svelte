<script>
  import {
    GetAllSchedules, GetScheduleByWeek, CreateSchedule,
    GetShiftsBySchedule, GenerateSchedule, GetAllEmployees,
    DeleteShiftsBySchedule
  } from '../../wailsjs/go/main/App.js';

  const DAY_NAMES = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

  let schedules = [];
  let employees = [];
  let currentSchedule = null;
  let shifts = [];
  let weekOf = sundayOf(new Date());
  let generating = false;
  let solverMessage = '';

  // -- date helpers --

  function sundayOf(date) {
    const d = new Date(date);
    d.setDate(d.getDate() - d.getDay());
    return d.toISOString().split('T')[0];
  }

  function formatWeekLabel(dateStr) {
    const d = new Date(dateStr + 'T00:00:00');
    return d.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });
  }

  function shiftWeek(delta) {
    const d = new Date(weekOf + 'T00:00:00');
    d.setDate(d.getDate() + delta * 7);
    weekOf = d.toISOString().split('T')[0];
    loadWeek();
  }

  // -- data loading --

  async function loadAll() {
    [schedules, employees] = await Promise.all([GetAllSchedules(), GetAllEmployees()]);
  }

  async function loadWeek() {
    solverMessage = '';
    try {
      currentSchedule = await GetScheduleByWeek(weekOf);
      shifts = await GetShiftsBySchedule(currentSchedule.id);
    } catch {
      currentSchedule = null;
      shifts = [];
    }
  }

  async function ensureSchedule() {
    if (!currentSchedule) {
      const id = await CreateSchedule({ weekOf, notes: '' });
      currentSchedule = { id: Number(id), weekOf, notes: '' };
      await loadAll();
    }
  }

  async function generate() {
    await ensureSchedule();
    generating = true;
    solverMessage = 'Running solver...';
    try {
      const result = await GenerateSchedule(currentSchedule.id);
      solverMessage = result.message;
      if (result.solved) {
        shifts = await GetShiftsBySchedule(currentSchedule.id);
      }
    } catch (e) {
      solverMessage = 'Error: ' + String(e);
    }
    generating = false;
  }

  async function clearShifts() {
    if (!currentSchedule) return;
    await DeleteShiftsBySchedule(currentSchedule.id);
    shifts = [];
    solverMessage = '';
  }

  // -- derived --

  $: employeeMap = Object.fromEntries(employees.map(e => [e.id, e.name]));

  $: shiftsByDay = (() => {
    const groups = {};
    for (let d = 1; d <= 7; d++) groups[d] = [];
    for (const s of shifts) groups[s.dayOfWeek].push(s);
    return groups;
  })();

  // -- event handlers --

  function onDateInput(e) {
    weekOf = sundayOf(new Date(e.target.value + 'T00:00:00'));
    loadWeek();
  }

  function onScheduleSelect(e) {
    const id = parseInt(e.target.value);
    if (!id) return;
    const s = schedules.find(s => s.id === id);
    if (s) {
      weekOf = s.weekOf;
      currentSchedule = s;
      GetShiftsBySchedule(id).then(r => (shifts = r));
      solverMessage = '';
    }
  }

  loadAll().then(loadWeek);
</script>

<div class="page">
  <h2>Schedule</h2>

  <div class="week-nav">
    <button on:click={() => shiftWeek(-1)}>◀</button>
    <span class="week-label">Week of {formatWeekLabel(weekOf)}</span>
    <button on:click={() => shiftWeek(1)}>▶</button>
    <input type="date" value={weekOf} min="2023-01-01" step="7" on:change={onDateInput} title="Pick a Sunday" />
  </div>

  <div class="actions">
    <button on:click={generate} disabled={generating}>
      {generating ? 'Generating...' : 'Generate'}
    </button>
    <button on:click={clearShifts} disabled={!currentSchedule || shifts.length === 0}>Clear Shifts</button>

    {#if schedules.length > 0}
      <select on:change={onScheduleSelect}>
        <option value="">Past schedules...</option>
        {#each schedules as s}
          <option value={s.id} selected={currentSchedule?.id === s.id}>
            {formatWeekLabel(s.weekOf)}
          </option>
        {/each}
      </select>
    {/if}
  </div>

  {#if solverMessage}
    <p class="solver-msg">{solverMessage}</p>
  {/if}

  {#if shifts.length === 0 && currentSchedule}
    <p style="opacity: 0.5;">No shifts yet. Hit Generate to build a schedule.</p>
  {:else if !currentSchedule}
    <p style="opacity: 0.5;">No schedule for this week yet.</p>
  {/if}

  <div class="days">
    {#each [1,2,3,4,5,6,7] as day}
      {#if shiftsByDay[day]?.length > 0}
        <div class="day-block">
          <strong>{DAY_NAMES[day]}</strong>
          {#each shiftsByDay[day] as shift}
            <div class="shift-row">
              <span class="emp-name">{employeeMap[shift.employeeId] ?? `#${shift.employeeId}`}</span>
              <span class="shift-time">{shift.startTime} – {shift.endTime}</span>
            </div>
          {/each}
        </div>
      {/if}
    {/each}
  </div>
</div>

<style>
  .page {
    max-width: 680px;
    margin: 0 auto;
    padding: 1rem 2rem;
    text-align: left;
  }

  .week-nav {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .week-label {
    font-size: 1.05rem;
    min-width: 220px;
    text-align: center;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-bottom: 0.75rem;
  }

  .solver-msg {
    opacity: 0.75;
    font-size: 0.9rem;
    margin-bottom: 0.75rem;
  }

  .days {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: 1rem;
  }

  .day-block {
    border-left: 2px solid rgba(255,255,255,0.15);
    padding-left: 0.75rem;
  }

  .day-block strong {
    display: block;
    margin-bottom: 0.3rem;
    opacity: 0.85;
  }

  .shift-row {
    display: flex;
    justify-content: space-between;
    font-size: 0.9rem;
    padding: 0.15rem 0;
    max-width: 320px;
  }

  .emp-name {
    opacity: 0.9;
  }

  .shift-time {
    opacity: 0.6;
    font-variant-numeric: tabular-nums;
  }

  select {
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.15);
    border-radius: 4px;
    color: white;
    padding: 0.25rem 0.4rem;
    font-size: 0.9rem;
  }
</style>
