<script>
  import { GetSettings, UpdateSettings, GetAllDaySettings, UpdateDaySettings } from '../../wailsjs/go/main/App.js';

  const DAY_NAMES = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

  let settings = null;
  let days = [];
  let prevNeeded = {}; // remember employees_needed before a day is closed

  let darkMode = localStorage.getItem('darkMode') !== 'false';
  let saveStatus = '';

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
    saveStatus = 'Saving...';
    try {
      await UpdateSettings(settings);
      for (const day of days) {
        await UpdateDaySettings(day);
      }
      saveStatus = 'Saved.';
    } catch (e) {
      saveStatus = 'Error saving.';
    }
    setTimeout(() => (saveStatus = ''), 2000);
  }

  function toggleDarkMode() {
    darkMode = !darkMode;
    localStorage.setItem('darkMode', String(darkMode));
    document.body.classList.toggle('light-mode', !darkMode);
  }

  load();
</script>

<div class="page">
  <h2>Settings</h2>

  {#if settings}
    <section>
      <h3>Scheduling</h3>

      <div class="field-row">
        <label>Min shift length (hrs)</label>
        <input type="number" bind:value={settings.minShiftLength} min="1" max="12" />
      </div>
      <div class="field-row">
        <label>Max shift length (hrs)</label>
        <input type="number" bind:value={settings.maxShiftLength} min="1" max="16" />
      </div>
      <div class="field-row">
        <label>Min time between shifts (hrs)</label>
        <input type="number" bind:value={settings.timeBetweenShifts} min="0" max="24" />
      </div>
      <div class="field-row">
        <label>Allow scheduling outside store hours</label>
        <input type="checkbox" bind:checked={settings.allowOutsideHours} />
      </div>
    </section>

    <section>
      <h3>Per-Day Settings</h3>
      <table>
        <thead>
          <tr>
            <th>Day</th>
            <th>Open</th>
            <th>Close</th>
            {#if settings.allowOutsideHours}
              <th>Sched. Open</th>
              <th>Sched. Close</th>
            {/if}
            <th>Employees Needed</th>
            <th>Open</th>
          </tr>
        </thead>
        <tbody>
          {#each days as day}
            {@const closed = day.employeesNeeded === 0}
            <tr class:closed>
              <td>{DAY_NAMES[day.dayOfWeek]}</td>
              <td><input type="time" bind:value={day.openTime} disabled={closed} /></td>
              <td><input type="time" bind:value={day.closeTime} disabled={closed} /></td>
              {#if settings.allowOutsideHours}
                <td><input type="time" bind:value={day.schedulableOpen} disabled={closed} /></td>
                <td><input type="time" bind:value={day.schedulableClose} disabled={closed} /></td>
              {/if}
              <td><input type="number" bind:value={day.employeesNeeded} min="0" max="20" disabled={closed} style="width: 4rem;" /></td>
              <td><input type="checkbox" checked={!closed} on:change={() => toggleDay(day)} /></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </section>

    <button on:click={save}>Save</button>
    {#if saveStatus}<span style="margin-left: 0.75rem; opacity: 0.7;">{saveStatus}</span>{/if}
  {:else}
    <p>Loading...</p>
  {/if}

  <section>
    <h3>App</h3>
    <div class="field-row">
      <label>Dark mode</label>
      <input type="checkbox" checked={darkMode} on:change={toggleDarkMode} />
    </div>
  </section>
</div>

<style>
  .page {
    max-width: 780px;
    margin: 0 auto;
    padding: 1rem 2rem;
    text-align: left;
  }

  section {
    margin-bottom: 2rem;
  }

  .field-row {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 0.6rem;
  }

  .field-row label {
    min-width: 260px;
    opacity: 0.85;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  th, td {
    text-align: left;
    padding: 0.4rem 0.6rem;
    border-bottom: 1px solid rgba(255,255,255,0.07);
  }

  th {
    opacity: 0.6;
    font-weight: normal;
  }

  tr.closed td {
    opacity: 0.35;
  }

  input[type="number"],
  input[type="time"] {
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.15);
    border-radius: 4px;
    color: white;
    padding: 0.25rem 0.4rem;
    font-size: 0.9rem;
  }

  input:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
</style>
