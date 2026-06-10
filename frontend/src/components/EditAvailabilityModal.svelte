<script>
  /**
   * EditAvailabilityModal — edit one employee's weekly availability.
   *
   * Each of the 7 days holds zero or more availability rows. A row is either
   * "all day" (00:00–24:00, the agreed encoding so the solver does NOT read it
   * as a day-off) or a manual start/end window. Rows within a day may not
   * overlap. A day with no rows means the employee is off that day.
   *
   * On save it emits the full flat row set; the parent persists it (delete +
   * recreate, since there's no bulk endpoint).
   */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let show = false;
  export let employee = null;          // { id, name }
  export let availabilities = [];      // rows for this employee: {dayOfWeek, startTime, endTime}

  const DAY_NAMES = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const DAYS = [1, 2, 3, 4, 5, 6, 7];
  const ALL_DAY = { start: '00:00', end: '24:00' };

  // Internal editable state: { [day]: [ {start, end, allDay} ] }
  let rowsByDay = {};

  function tmin(t) {
    if (!t) return null;
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  function isAllDay(r) {
    return r.startTime === ALL_DAY.start && r.endTime === ALL_DAY.end;
  }

  // Build internal state whenever the modal opens for a (new) employee.
  function hydrate() {
    const next = {};
    for (const d of DAYS) next[d] = [];
    for (const a of availabilities) {
      if (!next[a.dayOfWeek]) next[a.dayOfWeek] = [];
      next[a.dayOfWeek].push({
        start: a.startTime,
        end: a.endTime,
        allDay: isAllDay(a),
      });
    }
    rowsByDay = next;
  }

  // Re-hydrate each time it becomes visible
  let wasShown = false;
  $: if (show && !wasShown) { hydrate(); wasShown = true; }
  $: if (!show) wasShown = false;

  function addRow(day) {
    rowsByDay[day] = [...rowsByDay[day], { start: '09:00', end: '17:00', allDay: false }];
    rowsByDay = rowsByDay;
  }

  function removeRow(day, idx) {
    rowsByDay[day] = rowsByDay[day].filter((_, i) => i !== idx);
    rowsByDay = rowsByDay;
  }

  function toggleAllDay(day, idx) {
    const row = rowsByDay[day][idx];
    row.allDay = !row.allDay;
    if (row.allDay) {
      row.start = ALL_DAY.start;
      row.end = ALL_DAY.end;
    } else {
      row.start = '09:00';
      row.end = '17:00';
    }
    rowsByDay = rowsByDay;
  }

  // Per-row validity. A row errors if start>=end, or if it overlaps another
  // row on the same day. Overlap: aStart < bEnd && bStart < aEnd.
  function rowError(day, idx) {
    const row = rowsByDay[day][idx];
    if (row.allDay) {
      // all-day is valid on its own but conflicts with any other row that day
      return rowsByDay[day].length > 1 ? 'Remove other rows to use All day' : null;
    }
    const s = tmin(row.start), e = tmin(row.end);
    if (s == null || e == null) return 'Set both times';
    if (s >= e) return 'Start must be before end';
    for (let j = 0; j < rowsByDay[day].length; j++) {
      if (j === idx) continue;
      const other = rowsByDay[day][j];
      const os = other.allDay ? 0 : tmin(other.start);
      const oe = other.allDay ? 24 * 60 : tmin(other.end);
      if (os == null || oe == null) continue;
      if (s < oe && os < e) return 'Overlaps another slot';
    }
    return null;
  }

  // Whole-modal validity
  $: hasErrors = DAYS.some(d =>
    rowsByDay[d]?.some((_, i) => rowError(d, i) != null)
  );

  $: totalRows = DAYS.reduce((n, d) => n + (rowsByDay[d]?.length ?? 0), 0);

  function save() {
    if (hasErrors) return;
    const flat = [];
    for (const d of DAYS) {
      for (const row of rowsByDay[d] ?? []) {
        flat.push({ dayOfWeek: d, startTime: row.start, endTime: row.end });
      }
    }
    dispatch('save', { employeeId: employee.id, rows: flat });
  }

  function close() {
    dispatch('close');
  }

  function fmt12(t) {
    if (t === '24:00') return '12:00 AM⁺';
    const [h, m] = t.split(':').map(Number);
    const p = h < 12 ? 'AM' : 'PM';
    let h12 = h % 12; if (h12 === 0) h12 = 12;
    return `${h12}:${String(m).padStart(2, '0')} ${p}`;
  }
</script>

{#if show && employee}
  <div class="overlay" role="presentation" on:click|self={close}>
    <div class="modal" role="dialog">
      <div class="modal-header">
        <div>
          <h3 class="modal-title">Edit Availability</h3>
          <span class="modal-sub">{employee.name}</span>
        </div>
        <button class="close-btn" on:click={close} title="Close">✕</button>
      </div>

      <div class="modal-body">
        {#each DAYS as day}
          <div class="day-block">
            <div class="day-row-head">
              <span class="day-label">{DAY_NAMES[day]}</span>
              {#if (rowsByDay[day]?.length ?? 0) === 0}
                <span class="off-tag">Off</span>
              {/if}
              <button class="add-btn" on:click={() => addRow(day)} title="Add a time slot">+</button>
            </div>

            {#each rowsByDay[day] ?? [] as row, idx (idx)}
              {@const err = rowError(day, idx)}
              <div class="slot-row" class:err={err != null}>
                <label class="allday-toggle">
                  <input type="checkbox" checked={row.allDay} on:change={() => toggleAllDay(day, idx)} />
                  <span>All day</span>
                </label>

                {#if row.allDay}
                  <span class="allday-readout">{fmt12(ALL_DAY.start)} – {fmt12(ALL_DAY.end)}</span>
                {:else}
                  <div class="time-inputs">
                    <input type="time" step="3600" bind:value={row.start} />
                    <span class="dash">–</span>
                    <input type="time" step="3600" bind:value={row.end} />
                  </div>
                {/if}

                <button class="remove-btn" on:click={() => removeRow(day, idx)} title="Remove slot">✕</button>

                {#if err}
                  <span class="err-msg">{err}</span>
                {/if}
              </div>
            {/each}
          </div>
        {/each}
      </div>

      <div class="modal-footer">
        <span class="footer-info">{totalRows} slot{totalRows !== 1 ? 's' : ''}</span>
        <div class="footer-actions">
          <button on:click={close}>Cancel</button>
          <button class="primary save-btn" on:click={save} disabled={hasErrors}>Save</button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(3px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 400;
    animation: fade 0.18s ease-out;
  }

  @keyframes fade { from { opacity: 0; } to { opacity: 1; } }

  .modal {
    width: 460px;
    max-width: calc(100vw - 2rem);
    max-height: calc(100vh - 3rem);
    display: flex;
    flex-direction: column;
    background: #0f1e30;
    border-radius: 10px;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.1);
    animation: modal-in 0.22s cubic-bezier(0.34, 1.4, 0.64, 1);
  }

  @keyframes modal-in {
    from { transform: scale(0.95) translateY(-8px); opacity: 0; }
    to   { transform: scale(1) translateY(0); opacity: 1; }
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 1rem 1.2rem 0.8rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  }

  .modal-title {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
  }

  .modal-sub {
    font-size: 0.78rem;
    color: rgba(255, 255, 255, 0.4);
  }

  .close-btn {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.35);
    font-size: 0.8rem;
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
  }
  .close-btn:hover {
    background: rgba(255, 255, 255, 0.07) !important;
    color: rgba(255, 255, 255, 0.7) !important;
    border-color: transparent !important;
  }

  .modal-body {
    padding: 0.6rem 1.2rem;
    overflow-y: auto;
    flex: 1;
  }

  .day-block {
    padding: 0.55rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }
  .day-block:last-child { border-bottom: none; }

  .day-row-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.3rem;
  }

  .day-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.7);
    min-width: 86px;
  }

  .off-tag {
    font-size: 0.66rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: rgba(255, 255, 255, 0.28);
  }

  .add-btn {
    margin-left: auto;
    width: 24px;
    height: 24px;
    padding: 0;
    font-size: 1rem;
    line-height: 1;
    border-radius: 5px;
  }

  .slot-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.3rem 0.1rem 0.3rem 0.5rem;
    margin: 0.2rem 0;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.03);
    flex-wrap: wrap;
  }

  .slot-row.err {
    background: rgba(248, 113, 113, 0.08);
    outline: 1px solid rgba(248, 113, 113, 0.25);
  }

  .allday-toggle {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.74rem;
    color: rgba(255, 255, 255, 0.55);
    cursor: pointer;
    white-space: nowrap;
  }

  .time-inputs {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .time-inputs input { padding: 0.25rem 0.4rem; font-size: 0.8rem; }

  .dash { color: rgba(255, 255, 255, 0.3); }

  .allday-readout {
    font-size: 0.78rem;
    color: rgba(255, 255, 255, 0.55);
    font-variant-numeric: tabular-nums;
  }

  .remove-btn {
    margin-left: auto;
    width: 22px;
    height: 22px;
    padding: 0;
    font-size: 0.68rem;
    border-radius: 5px;
    color: rgba(255, 255, 255, 0.4);
  }
  .remove-btn:hover {
    background: rgba(248, 113, 113, 0.15) !important;
    color: #f87171 !important;
    border-color: transparent !important;
  }

  .err-msg {
    flex-basis: 100%;
    font-size: 0.68rem;
    color: #f87171;
    padding-left: 0.5rem;
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.2rem;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
  }

  .footer-info {
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.35);
  }

  .footer-actions { display: flex; gap: 0.5rem; }

  .save-btn:not(:disabled) {
    background: #16a34a;
    border-color: #22c55e;
  }
  .save-btn:not(:disabled):hover {
    background: #15803d;
  }
</style>