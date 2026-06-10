<script>
  import { GetSettings, UpdateSettings, GetAllDaySettings, UpdateDaySettings, ClearData } from '../../wailsjs/go/main/App.js';
  import ConfirmModal from '../components/ConfirmModal.svelte';

  const DAY_NAMES = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const DAY_SHORT = ['', 'Sun',    'Mon',    'Tue',     'Wed',       'Thu',      'Fri',    'Sat'];

  let settings = null;
  let days = [];
  let prevNeeded = {};
  let saveStatus = ''; // '' | 'saving' | 'saved' | 'error'
  let showClearDb = false;
  let clearStatus = '';

  async function clearDatabase() {
    showClearDb = false;
    clearStatus = 'Clearing…';
    try {
      await ClearData();
      clearStatus = 'Database cleared.';
    } catch (e) {
      clearStatus = 'Error clearing database.';
    }
    setTimeout(() => (clearStatus = ''), 3000);
  }

  async function load() {
    settings = await GetSettings();
    const map = await GetAllDaySettings();
    days = Object.values(map).sort((a, b) => a.dayOfWeek - b.dayOfWeek);
    for (const d of days) {
      if (d.employeesNeeded > 0) prevNeeded[d.dayOfWeek] = d.employeesNeeded;
    }
  }

  function toggleDay(day) {
    if (day.employeesNeeded === 0) {
      day.employeesNeeded = prevNeeded[day.dayOfWeek] ?? 2;
    } else {
      prevNeeded[day.dayOfWeek] = day.employeesNeeded;
      day.employeesNeeded = 0;
    }
    days = days;
  }

  async function save() {
    saveStatus = 'saving';
    try {
      await UpdateSettings(settings);
      for (const day of days) await UpdateDaySettings(day);
      saveStatus = 'saved';
    } catch (e) {
      saveStatus = 'error';
    }
    setTimeout(() => (saveStatus = ''), 2200);
  }

  load();
</script>

<div class="page">
  <h2>Settings</h2>

  {#if settings}
    <!-- Shift constraints -->
    <section>
      <h3>Shift constraints</h3>
      <div class="fields-card">
        <div class="field-row">
          <div class="field-label">
            <span>Min shift length</span>
            <span class="field-unit">hours</span>
          </div>
          <input type="number" bind:value={settings.minShiftLength} min="1" max="12" />
        </div>
        <div class="divider" />
        <div class="field-row">
          <div class="field-label">
            <span>Max shift length</span>
            <span class="field-unit">hours</span>
          </div>
          <input type="number" bind:value={settings.maxShiftLength} min="1" max="16" />
        </div>
        <div class="divider" />
        <div class="field-row">
          <div class="field-label">
            <span>Min time between shifts</span>
            <span class="field-unit">hours</span>
          </div>
          <input type="number" bind:value={settings.timeBetweenShifts} min="0" max="24" />
        </div>
        <div class="divider" />
        <div class="field-row">
          <div class="field-label">
            <span>Allow scheduling outside store hours</span>
          </div>
          <input type="checkbox" bind:checked={settings.allowOutsideHours} />
        </div>
      </div>
    </section>

    <!-- Per-day settings -->
    <section>
      <h3>Store hours per day</h3>
      <div class="days-list">
        {#each days as day}
          {@const closed = day.employeesNeeded === 0}
          <div class="day-row" class:closed>
            <div class="day-left">
              <button
                class="toggle-pill"
                class:open={!closed}
                on:click={() => toggleDay(day)}
                title={closed ? 'Open this day' : 'Close this day'}
              >
                {closed ? 'Closed' : 'Open'}
              </button>
              <span class="day-name">{DAY_NAMES[day.dayOfWeek]}</span>
            </div>
            <div class="day-fields" class:disabled={closed}>
              <div class="time-pair">
                <span class="time-label">Open</span>
                <input type="time" bind:value={day.openTime} disabled={closed} />
              </div>
              <span class="time-dash">–</span>
              <div class="time-pair">
                <span class="time-label">Close</span>
                <input type="time" bind:value={day.closeTime} disabled={closed} />
              </div>
              {#if settings.allowOutsideHours}
                <div class="time-pair">
                  <span class="time-label">Sched. open</span>
                  <input type="time" bind:value={day.schedulableOpen} disabled={closed} />
                </div>
                <span class="time-dash">–</span>
                <div class="time-pair">
                  <span class="time-label">Sched. close</span>
                  <input type="time" bind:value={day.schedulableClose} disabled={closed} />
                </div>
              {/if}
              <div class="time-pair">
                <span class="time-label">Staff needed</span>
                <input type="number" bind:value={day.employeesNeeded} min="0" max="20" disabled={closed} style="width:3.5rem" />
              </div>
            </div>
          </div>
        {/each}
      </div>
    </section>

    <!-- Save -->
    <div class="save-row">
      <button class="primary" on:click={save} disabled={saveStatus === 'saving'}>
        {saveStatus === 'saving' ? 'Saving…' : 'Save changes'}
      </button>
      {#if saveStatus === 'saved'}
        <span class="status-ok">✓ Saved</span>
      {:else if saveStatus === 'error'}
        <span class="status-err">⚠ Error saving</span>
      {/if}
    </div>

    <!-- Danger zone -->
    <section class="danger-zone">
      <h3>Danger zone</h3>
      <div class="danger-card">
        <div class="danger-text">
          <strong>Clear database</strong>
          <span>Permanently deletes all employees, availability, schedules, and shifts. Store settings are kept. This cannot be undone.</span>
        </div>
        <button class="danger" on:click={() => (showClearDb = true)}>Clear database</button>
      </div>
      {#if clearStatus}<span class="status-err" style="margin-left:0.1rem;">{clearStatus}</span>{/if}
    </section>
  {:else}
    <p class="loading">Loading…</p>
  {/if}
</div>

<ConfirmModal
  show={showClearDb}
  title="Clear the entire database?"
  message="This permanently deletes ALL employees, availability, schedules, and shifts. There is no undo."
  requireText="clear database"
  confirmLabel="Permanently clear"
  danger={true}
  on:confirm={clearDatabase}
  on:cancel={() => (showClearDb = false)}
/>

<style>
  .page { max-width: 680px; }

  section { margin-bottom: 2rem; }

  .danger-zone h3 { color: rgba(248, 113, 113, 0.65); }

  .danger-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.85rem 1rem;
    border: 1px solid rgba(248, 113, 113, 0.22);
    background: rgba(248, 113, 113, 0.05);
    border-radius: 8px;
  }

  .danger-text {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }
  .danger-text strong {
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.82);
    font-weight: 600;
  }
  .danger-text span {
    font-size: 0.76rem;
    color: rgba(255, 255, 255, 0.45);
    line-height: 1.4;
  }

  .fields-card {
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 8px;
    overflow: hidden;
  }

  .field-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.7rem 1rem;
  }

  .field-label {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    font-size: 0.875rem;
    color: rgba(255,255,255,0.78);
  }

  .field-unit {
    font-size: 0.7rem;
    color: rgba(255,255,255,0.3);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .field-row input[type="number"] { width: 5rem; text-align: right; }

  .divider {
    height: 1px;
    background: rgba(255,255,255,0.06);
    margin: 0;
  }

  /* Days list */
  .days-list {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .day-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 7px;
    padding: 0.55rem 0.85rem;
    transition: opacity 0.15s;
  }

  .day-row.closed {
    opacity: 0.5;
  }

  .day-left {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    min-width: 130px;
  }

  .day-name {
    font-size: 0.875rem;
    color: rgba(255,255,255,0.75);
    font-weight: 500;
  }

  .toggle-pill {
    font-size: 0.7rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    padding: 0.2rem 0.6rem;
    border-radius: 20px;
    border: 1px solid rgba(239,68,68,0.35);
    background: rgba(239,68,68,0.1);
    color: #fca5a5;
    min-width: 54px;
    text-align: center;
    transition: all 0.15s;
  }

  .toggle-pill.open {
    border-color: rgba(34,197,94,0.35);
    background: rgba(34,197,94,0.1);
    color: #4ade80;
  }

  .toggle-pill:hover {
    filter: brightness(1.15);
  }

  .day-fields {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    flex: 1;
  }

  .day-fields.disabled { opacity: 0.35; pointer-events: none; }

  .time-pair {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    align-items: flex-start;
  }

  .time-label {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: rgba(255,255,255,0.28);
  }

  .time-dash {
    color: rgba(255,255,255,0.2);
    font-size: 0.9rem;
    margin-top: 0.9rem;
  }

  /* Save */
  .save-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding-top: 0.25rem;
  }

  .status-ok  { font-size: 0.82rem; color: #4ade80; }
  .status-err { font-size: 0.82rem; color: #f87171; }

  .loading { color: rgba(255,255,255,0.35); font-size: 0.875rem; }
</style>