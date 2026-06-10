<script>
  import { fade } from 'svelte/transition';
  import {
    GetAllSchedules, GetScheduleByWeek, CreateSchedule, DeleteSchedule,
    GetShiftsBySchedule, GenerateSchedule, GetAllEmployees,
    DeleteShiftsBySchedule, CreateShift, DeleteShift, CancelGeneration,
    GetSettings, GetAllDaySettings, GetAllAvailabilities,
  } from '../../wailsjs/go/main/App.js';
  import WeekGrid from '../components/WeekGrid.svelte';
  import AddShiftModal from '../components/AddShiftModal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';
  import InfoModal from '../components/InfoModal.svelte';
  import StatsModal from '../components/StatsModal.svelte';
  import { csvFilename, toGridCSV, downloadCSV } from '../lib/csvExport.js';
  import { computePayroll, computeScore, computeValidity } from '../lib/scheduleStats.js';

  const DAY_LONG = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const ROW_H = 44;

  let schedules = [];
  let employees = [];
  let availabilities = [];
  let settings = null;
  let daySettings = {};
  let currentSchedule = null;
  let shifts = [];
  let weekOf = defaultWeek();
  let generating = false;
  let solverMessage = '';
  let solverSuccess = false;
  let viewMode = 'grid';        // 'grid' | 'list'
  let prevSelectValue = '';     // keeps the "View previous schedules" dropdown showing its label
  let genTimer = null;          // client-side solver timeout

  // Modals
  let showAddShift = false;
  let showInfo = false;
  let showStats = false;
  let confirmState = { show: false, title: '', message: '', confirmLabel: 'Confirm', danger: false, action: null };

  // helper functions
  function tmin(t) {
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  function sundayOf(date) {
    const d = new Date(date);
    d.setDate(d.getDate() - d.getDay());
    return d.toISOString().split('T')[0];
  }

  // Default to NEXT week
  function defaultWeek() {
    const d = new Date();
    d.setDate(d.getDate() + 7);
    return sundayOf(d);
  }

  function formatWeekLabel(dateStr) {
    const d = new Date(dateStr + 'T00:00:00');
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function dayDate(weekStr, dayOfWeek) {
    const d = new Date(weekStr + 'T00:00:00');
    d.setDate(d.getDate() + (dayOfWeek - 1));
    return d.getDate();
  }

  function shiftWeek(delta) {
    const d = new Date(weekOf + 'T00:00:00');
    d.setDate(d.getDate() + delta * 7);
    weekOf = d.toISOString().split('T')[0];
    loadWeek();
  }

  // data actions
  async function loadAll() {
    [schedules, employees, settings, daySettings, availabilities] = await Promise.all([
      GetAllSchedules(), GetAllEmployees(), GetSettings(), GetAllDaySettings(), GetAllAvailabilities(),
    ]);
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

  // generate (with confirm if shifts already exist)
  function requestGenerate() {
    if (shifts.length > 0) {
      askConfirm({
        title: 'Regenerate schedule?',
        message: `This replaces the ${shifts.length} existing shift${shifts.length !== 1 ? 's' : ''} for this week.`,
        confirmLabel: 'Regenerate',
        danger: true,
        action: generate,
      });
    } else {
      generate();
    }
  }

  async function generate() {
    await ensureSchedule();
    generating = true;
    solverMessage = 'Running solver…';
    solverSuccess = false;

    // Client-side safety timeout: if the solver hasn't returned in 15s, ask the
    // backend to cancel the run (it bails out and returns a timed-out result).
    clearTimeout(genTimer);
    genTimer = setTimeout(() => { CancelGeneration().catch(() => {}); }, 15000);

    try {
      const result = await GenerateSchedule(currentSchedule.id);
      solverMessage = result.message;
      solverSuccess = result.solved;
      if (result.solved) shifts = await GetShiftsBySchedule(currentSchedule.id);
    } catch (e) {
      solverMessage = 'Error: ' + String(e);
    } finally {
      clearTimeout(genTimer);
      genTimer = null;
      generating = false;
    }
  }

  // Manual stop button while generating.
  function stopGenerate() {
    CancelGeneration().catch(() => {});
  }

  // clear shifts (with confirm)
  function requestClearShifts() {
    if (!currentSchedule || shifts.length === 0) return;
    askConfirm({
      title: 'Clear all shifts?',
      message: 'Removes every shift from this week. The schedule itself is kept.',
      confirmLabel: 'Clear shifts',
      danger: true,
      action: clearShifts,
    });
  }

  async function clearShifts() {
    await DeleteShiftsBySchedule(currentSchedule.id);
    shifts = [];
    solverMessage = '';
  }

  // delete schedule (with confirm)
  function requestDeleteSchedule() {
    if (!currentSchedule) return;
    askConfirm({
      title: 'Delete this schedule?',
      message: `Permanently deletes the week of ${formatWeekLabel(weekOf)} and all its shifts. This cannot be undone.`,
      confirmLabel: 'Delete schedule',
      danger: true,
      action: deleteSchedule,
    });
  }

  async function deleteSchedule() {
    const id = currentSchedule.id;
    await DeleteShiftsBySchedule(id); // explicit; FK cascade would also handle this
    await DeleteSchedule(id);
    currentSchedule = null;
    shifts = [];
    solverMessage = '';
    await loadAll();
  }

  // add shift
  async function openAddShift() {
    await ensureSchedule();
    showAddShift = true;
  }

  async function handleAddShift(event) {
    const s = event.detail;
    try {
      await CreateShift({
        scheduleId: currentSchedule.id,
        employeeId: s.employeeId,
        dayOfWeek: s.dayOfWeek,
        startTime: s.startTime,
        endTime: s.endTime,
      });
      shifts = await GetShiftsBySchedule(currentSchedule.id);
    } catch (e) {
      console.error('Failed to add shift:', e);
    }
    showAddShift = false;
  }

  // Delete a single shift (X on a blob / chip). Optimistically drop it locally.
  async function deleteShift(id) {
    if (id == null) return;
    try {
      await DeleteShift(id);
      shifts = shifts.filter(s => s.id !== id);
    } catch (e) {
      console.error('Failed to delete shift:', e);
    }
  }

  // export
  function exportCSV() {
    if (shifts.length === 0) return;
    const csv = toGridCSV(shifts, employees);
    downloadCSV(csvFilename(weekOf), csv);
  }

  // confirm modal
  function askConfirm(opts) {
    confirmState = { show: true, confirmLabel: 'Confirm', danger: false, action: null, message: '', ...opts };
  }
  function onConfirmYes() {
    const action = confirmState.action;
    confirmState = { ...confirmState, show: false };
    if (action) action();
  }
  function onConfirmNo() {
    confirmState = { ...confirmState, show: false };
  }

  // input handlers
  function onDateInput(e) {
    weekOf = sundayOf(new Date(e.target.value + 'T00:00:00'));
    loadWeek();
  }

  function onScheduleSelect(e) {
    const id = parseInt(e.target.value);
    prevSelectValue = ''; // snap the dropdown back to its label
    if (!id) return;
    const s = schedules.find(s => s.id === id);
    if (s) {
      weekOf = s.weekOf;
      loadWeek(); // set new schedule and shifts
    }
  }

  // create employee map for easy lookup
  $: employeeMap = Object.fromEntries(employees.map(e => [e.id, e]));

  // Two palettes that share hues: solid blobs (grid) and translucent chips (list).
  const EMP_BLOB = [
    { bg: 'rgba(59,130,246,0.85)',  text: '#fff' },
    { bg: 'rgba(168,85,247,0.85)',  text: '#fff' },
    { bg: 'rgba(34,197,94,0.8)',    text: '#06281a' },
    { bg: 'rgba(251,191,36,0.85)',  text: '#3a2e05' },
    { bg: 'rgba(248,113,113,0.85)', text: '#fff' },
    { bg: 'rgba(20,184,166,0.8)',   text: '#03201d' },
  ];
  const EMP_CHIP = [
    { bg: 'rgba(59,130,246,0.16)',  border: 'rgba(59,130,246,0.4)',  text: '#93c5fd' },
    { bg: 'rgba(168,85,247,0.16)',  border: 'rgba(168,85,247,0.4)',  text: '#c084fc' },
    { bg: 'rgba(34,197,94,0.15)',   border: 'rgba(34,197,94,0.4)',   text: '#4ade80' },
    { bg: 'rgba(251,191,36,0.15)',  border: 'rgba(251,191,36,0.4)',  text: '#fbbf24' },
    { bg: 'rgba(248,113,113,0.15)', border: 'rgba(248,113,113,0.4)', text: '#fca5a5' },
    { bg: 'rgba(20,184,166,0.15)',  border: 'rgba(20,184,166,0.4)',  text: '#5eead4' },
  ];
  $: empIndex = Object.fromEntries(employees.map((e, i) => [e.id, i]));
  $: blobColor = (id) => EMP_BLOB[(empIndex[id] ?? 0) % EMP_BLOB.length];
  $: chipColor = (id) => EMP_CHIP[(empIndex[id] ?? 0) % EMP_CHIP.length];

  // group shifts by day for easy lookup when rendering the grid
  $: shiftsByDay = (() => {
    const groups = {};
    for (let d = 1; d <= 7; d++) groups[d] = [];
    for (const s of shifts) groups[s.dayOfWeek].push(s);
    for (let d = 1; d <= 7; d++) groups[d].sort((a, b) => tmin(a.startTime) - tmin(b.startTime));
    return groups;
  })();

  /**
   * A shift can span several hour-rows (e.g. 09:00–13:00 = 4 rows). The grid
   * gives us one slot per (day, hour) cell, so I render each shift's blob ONCE
   * in its starting cell and let it overflow downward, augmenting its height using
   * the constant ROW_H * the shift's duration in hours.
   *
   * To stop overlapping shifts on the same day from sitting on top of each
   * other, we assign each shift a "lane" (a column within the day) using a
   * simple greedy interval-packing pass: walk the day's shifts in start order,
   * and drop each into the first lane whose previous shift has already ended;
   * if none is free, open a new lane. The number of lanes a day ends up using
   * sets how wide each blob is (1 / laneCount of that day).
   *
   */
  $: startMap = (() => {
    const map = {};
    const byDay = {};
    for (let d = 1; d <= 7; d++) byDay[d] = [];
    for (const s of shifts) byDay[s.dayOfWeek].push(s);

    // greedy lane assignment within each day
    for (let d = 1; d <= 7; d++) {
      const sorted = [...byDay[d]].sort((a, b) => tmin(a.startTime) - tmin(b.startTime)); // sort by start time
      const laneEndMin = []; // when the last shift in each lane ends
      const placed = [];
      for (const s of sorted) {
        const start = tmin(s.startTime);
        const end = tmin(s.endTime);
        let lane = laneEndMin.findIndex(endMin => endMin <= start);
        if (lane === -1) { lane = laneEndMin.length; laneEndMin.push(end); }
        else laneEndMin[lane] = end;
        placed.push({ shift: s, lane, start, end });
      }
      const laneCount = Math.max(1, laneEndMin.length);
      for (const p of placed) {
        const key = `${d}-${p.start}`;
        (map[key] ??= []).push({ ...p, laneCount });
      }
    }
    return map;
  })();

  $: totalShifts = shifts.length;
  $: hoursScheduled = shifts.reduce((sum, s) => sum + (tmin(s.endTime) - tmin(s.startTime)) / 60, 0);

  // Reactive stats that are recomputed whenever shifts / employees / settings change.
  $: payroll  = settings ? computePayroll(shifts, employees) : null;
  $: score    = settings ? computeScore(shifts, employees, daySettings, settings) : null;
  $: validity = settings ? computeValidity(shifts, employees, daySettings, settings, availabilities) : null;

  loadAll().then(loadWeek);
</script>

<div class="page">
  <!-- Header -->
  <div class="page-header">
    <div class="title-row">
      <h2>Schedule</h2>
      {#if currentSchedule}
        <button class="stats-pill" on:click={() => (showStats = true)} title="View stats">
          <span class="stat">{totalShifts} shift{totalShifts !== 1 ? 's' : ''}</span>
          <span class="stat-sep">·</span>
          <span class="stat">{hoursScheduled.toFixed(0)}h</span>
          <span class="stat-sep">·</span>
          {#if validity}
            <span class="verdict" class:ok={validity.valid} class:bad={!validity.valid}>
              {validity.valid ? 'Valid' : 'Invalid'}
            </span>
          {/if}
        </button>
      {/if}
    </div>
    <button class="info-btn" on:click={() => (showInfo = true)} title="About scheduling">ⓘ</button>
  </div>

  <!-- Toolbar: date control (left) · view toggle (center) · previous schedules (right) -->
  <div class="toolbar">
    <div class="date-control">
      <label class="cal-btn" title="Jump to a date">
        📅
        <input type="date" value={weekOf} min="2023-01-01" on:change={onDateInput} />
      </label>
      <button class="nav-btn" on:click={() => shiftWeek(-1)} title="Previous week">‹</button>
      <span class="week-label">Week of {formatWeekLabel(weekOf)}</span>
      <button class="nav-btn" on:click={() => shiftWeek(1)} title="Next week">›</button>
    </div>

    <div class="view-toggle">
      <button class:active={viewMode === 'grid'} on:click={() => (viewMode = 'grid')}>Grid</button>
      <button class:active={viewMode === 'list'} on:click={() => (viewMode = 'list')}>List</button>
    </div>

    {#if schedules.length > 0}
      <select class="prev-select" bind:value={prevSelectValue} on:change={onScheduleSelect}>
        <option value="">View previous schedules…</option>
        {#each schedules as s}
          <option value={s.id}>{formatWeekLabel(s.weekOf)}</option>
        {/each}
      </select>
    {/if}
  </div>

  <!-- Actions -->
  <div class="actions">
    <button class="primary" on:click={requestGenerate} disabled={generating}>
      {#if generating}<span class="spinner">⟳</span> Generating…{:else}✦ Generate{/if}
    </button>
    {#if generating}
      <button class="danger" on:click={stopGenerate}>Stop</button>
    {/if}
    <button on:click={openAddShift}>+</button>
    <button on:click={exportCSV} disabled={shifts.length === 0}>⭳</button>
    <button on:click={requestClearShifts} disabled={!currentSchedule || shifts.length === 0}>Clear shifts</button>
    <button class="danger" on:click={requestDeleteSchedule} disabled={!currentSchedule}>Delete schedule</button>
  </div>

  <!-- Solver message -->
  {#if solverMessage}
    <div class="solver-msg" class:success={solverSuccess} in:fade>
      <span class="msg-icon">{solverSuccess ? '✓' : 'ℹ'}</span>
      {solverMessage}
    </div>
  {/if}

  <!-- Empty -->
  {#if shifts.length === 0 && !generating}
    <div class="empty-state" in:fade>
      <span class="empty-icon">🗓️</span>
      <p>{currentSchedule ? 'No shifts yet — hit Generate or add one manually.' : 'No schedule for this week yet.'}</p>
    </div>
  {:else if viewMode === 'grid'}
    <!-- Grid view -->
    {#if settings}
      <WeekGrid {daySettings} allowOutsideHours={settings.allowOutsideHours} rowHeight={ROW_H} let:day let:hour>
        {#each startMap[`${day}-${hour}`] ?? [] as item (item.shift.id ?? `${item.shift.employeeId}-${item.start}`)}
          {@const col = blobColor(item.shift.employeeId)}
          {@const spanHours = (item.end - item.start) / 60}
          <div
            class="blob"
            style="
              height: {spanHours * ROW_H - 3}px;
              left: calc({(item.lane / item.laneCount) * 100}% + 1px);
              width: calc({100 / item.laneCount}% - 2px);
              background: {col.bg};
              color: {col.text};
            "
            title="{employeeMap[item.shift.employeeId]?.name ?? '#' + item.shift.employeeId}: {item.shift.startTime}–{item.shift.endTime}"
          >
            <button class="blob-x" on:click|stopPropagation={() => deleteShift(item.shift.id)} title="Delete shift">✕</button>
            <span class="blob-name">{employeeMap[item.shift.employeeId]?.name ?? `#${item.shift.employeeId}`}</span>
            <span class="blob-time">{item.shift.startTime}–{item.shift.endTime}</span>
          </div>
        {/each}
      </WeekGrid>
    {/if}
  {:else}
    <!-- List view -->
    <div class="days-grid">
      {#each [1, 2, 3, 4, 5, 6, 7] as day}
        {#if shiftsByDay[day]?.length > 0}
          <div class="day-card" in:fade={{ delay: (day - 1) * 25 }}>
            <div class="day-header">
              <span class="day-name">{DAY_LONG[day]}</span>
              <span class="day-date">{dayDate(weekOf, day)}</span>
            </div>
            <div class="shift-list">
              {#each shiftsByDay[day] as shift}
                {@const col = chipColor(shift.employeeId)}
                <div class="shift-chip" style="background:{col.bg}; border-color:{col.border}">
                  <span class="chip-name" style="color:{col.text}">
                    {employeeMap[shift.employeeId]?.name ?? `#${shift.employeeId}`}
                  </span>
                  <span class="chip-time">{shift.startTime}–{shift.endTime}</span>
                  <button class="chip-x" on:click={() => deleteShift(shift.id)} title="Delete shift">✕</button>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<!-- Modals -->
<AddShiftModal show={showAddShift} {employees} on:save={handleAddShift} on:close={() => (showAddShift = false)} />
<InfoModal show={showInfo} on:close={() => (showInfo = false)} />
<StatsModal show={showStats} {payroll} {score} {validity} on:close={() => (showStats = false)} />
<ConfirmModal
  show={confirmState.show}
  title={confirmState.title}
  message={confirmState.message}
  confirmLabel={confirmState.confirmLabel}
  danger={confirmState.danger}
  on:confirm={onConfirmYes}
  on:cancel={onConfirmNo}
/>

<style>
  .page { max-width: 880px; }

  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1.1rem;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 0.85rem;
  }
  .title-row h2 { margin: 0; }

  .stats-pill {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.8rem;
    padding: 0.3rem 0.7rem;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.45);
  }
  .stats-pill:hover {
    background: rgba(255, 255, 255, 0.08) !important;
    color: rgba(255, 255, 255, 0.7) !important;
  }
  .stat { color: inherit; }
  .stat-sep { opacity: 0.35; }
  .verdict { font-weight: 600; }
  .verdict.ok  { color: #4ade80; }
  .verdict.bad { color: #f87171; }

  .info-btn {
    width: 30px; height: 30px; padding: 0;
    border-radius: 50%;
    font-size: 0.9rem;
    color: rgba(255, 255, 255, 0.5);
  }

  /* Toolbar */
  .toolbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1rem;
    padding: 0.55rem 0.8rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
  }

  .date-control {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .cal-btn {
    position: relative;
    width: 30px; height: 28px;
    display: flex; align-items: center; justify-content: center;
    border-radius: 5px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.07);
    cursor: pointer;
    font-size: 0.85rem;
    overflow: hidden;
  }
  .cal-btn:hover { background: rgba(255, 255, 255, 0.12); }
  /* The real date input sits transparently over the icon so a click opens the
     native picker without needing the (less-supported) showPicker() API. */
  .cal-btn input[type="date"] {
    position: absolute;
    inset: 0;
    opacity: 0;
    border: none;
    padding: 0;
    cursor: pointer;
  }

  .nav-btn {
    width: 28px; height: 28px; padding: 0;
    display: flex; align-items: center; justify-content: center;
    font-size: 1.1rem; border-radius: 5px;
  }

  .week-label {
    font-size: 0.88rem;
    font-weight: 500;
    color: rgba(255, 255, 255, 0.78);
    min-width: 168px;
    text-align: center;
  }

  .view-toggle {
    display: flex;
    margin-left: auto;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 6px;
    overflow: hidden;
  }
  .view-toggle button {
    border: none;
    border-radius: 0;
    background: transparent;
    padding: 0.3rem 0.75rem;
    font-size: 0.8rem;
    color: rgba(255, 255, 255, 0.5);
  }
  .view-toggle button.active {
    background: rgba(37, 99, 235, 0.25) !important;
    color: #93c5fd !important;
  }

  .prev-select { font-size: 0.82rem; max-width: 200px; }

  /* Actions */
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .spinner { display: inline-block; animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .solver-msg {
    display: flex; align-items: center; gap: 0.5rem;
    font-size: 0.82rem;
    color: rgba(255, 255, 255, 0.5);
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    padding: 0.5rem 0.75rem;
    margin-bottom: 1rem;
  }
  .solver-msg.success {
    color: #4ade80;
    background: rgba(34, 197, 94, 0.08);
    border-color: rgba(34, 197, 94, 0.2);
  }
  .msg-icon { font-size: 0.9rem; }

  .empty-state {
    display: flex; flex-direction: column; align-items: center; gap: 0.6rem;
    padding: 4rem 1rem;
    text-align: center;
    color: rgba(255, 255, 255, 0.28);
  }
  .empty-icon { font-size: 2.5rem; opacity: 0.35; }
  .empty-state p { margin: 0; font-size: 0.875rem; }

  /* Shift blobs (grid view) */
  .blob {
    position: absolute;
    top: 2px;
    z-index: 3;
    border-radius: 5px;
    padding: 0.2rem 0.4rem;
    display: flex; flex-direction: column; justify-content: flex-start; gap: 1px;
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
  }
  .blob-name {
    font-size: 0.72rem; font-weight: 600; line-height: 1.1;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  .blob-time {
    font-size: 0.62rem; opacity: 0.8;
    font-variant-numeric: tabular-nums; white-space: nowrap;
  }
  .blob-x {
    position: absolute;
    top: 1px;
    right: 1px;
    width: 16px;
    height: 16px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.6rem;
    line-height: 1;
    border-radius: 50%;
    border: none;
    background: rgba(0, 0, 0, 0.35);
    color: rgba(255, 255, 255, 0.9);
    opacity: 0;
    transition: opacity 0.12s;
  }
  .blob:hover .blob-x { opacity: 1; }
  .blob-x:hover {
    background: rgba(239, 68, 68, 0.85) !important;
    color: #fff !important;
  }

  /* Day cards (list view) */
  .days-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 0.65rem;
  }
  .day-card {
    background: rgba(255, 255, 255, 0.025);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    overflow: hidden;
  }
  .day-header {
    display: flex; align-items: center; justify-content: space-between;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    background: rgba(255, 255, 255, 0.025);
  }
  .day-name {
    font-size: 0.8rem; font-weight: 600;
    color: rgba(255, 255, 255, 0.7);
    text-transform: uppercase; letter-spacing: 0.05em;
  }
  .day-date { font-size: 0.8rem; color: rgba(255, 255, 255, 0.28); }
  .shift-list {
    display: flex; flex-direction: column; gap: 0.35rem;
    padding: 0.55rem 0.6rem;
  }
  .shift-chip {
    position: relative;
    display: flex; align-items: center; justify-content: space-between;
    padding: 0.35rem 0.6rem;
    border-radius: 5px;
    border: 1px solid;
  }
  .chip-name { font-size: 0.82rem; font-weight: 500; }
  .chip-time {
    font-size: 0.76rem; color: rgba(255, 255, 255, 0.4);
    font-variant-numeric: tabular-nums;
  }
  .chip-x {
    position: absolute;
    top: 50%;
    right: 4px;
    transform: translateY(-50%);
    width: 16px;
    height: 16px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.6rem;
    line-height: 1;
    border-radius: 50%;
    border: none;
    background: rgba(0, 0, 0, 0.3);
    color: rgba(255, 255, 255, 0.85);
    opacity: 0;
    transition: opacity 0.12s;
  }
  .shift-chip:hover .chip-x { opacity: 1; }
  .shift-chip:hover .chip-time { opacity: 0; }
  .chip-x:hover {
    background: rgba(239, 68, 68, 0.85) !important;
    color: #fff !important;
  }
</style>