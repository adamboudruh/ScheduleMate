<script>
  import { fade } from 'svelte/transition';
  import {
    GetAllEmployees, GetAllAvailabilities,
    GetSettings, GetAllDaySettings,
    CreateAvailability, DeleteAvailability,
  } from '../../wailsjs/go/main/App.js';
  import WeekGrid from '../components/WeekGrid.svelte';
  import EditAvailabilityModal from '../components/EditAvailabilityModal.svelte';

  // state vars containing all employees, their availabilities, and the settings
  let employees = [];
  let allAvail = [];
  let settings = null;
  let daySettings = {};

  let selected = 'general';   // 'general' | employeeId
  let showModal = false;
  let saving = false;

  function hoursToMinutes(t) { // converts HH:MM to minutes after midnight
    if (!t) return null;
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  async function loadData() {
    [employees, allAvail, settings, daySettings] = await Promise.all([
      GetAllEmployees(),
      GetAllAvailabilities(),
      GetSettings(),
      GetAllDaySettings(),
    ]);
  }

  // True if an hour block is covered by at least one availability in the list, used for employee and general
  function isCovered(availabilities, day, hour) {
    return availabilities.some(a =>
      a.dayOfWeek === day &&
      hoursToMinutes(a.startTime) <= hour &&
      hoursToMinutes(a.endTime) >= hour + 60
    );
  }

  $: selectedEmployee = selected === 'general'
    ? null
    : employees.find(e => e.id === Number(selected)) ?? null;

  $: selectedAvail = selectedEmployee
    ? allAvail.filter(a => a.employeeId === selectedEmployee.id)
    : [];

  // For the general view: how many employees are available per (day, hour).
  function generalCount(day, hour) {
    let n = 0;
    for (const e of employees) {
      const rows = allAvail.filter(a => a.employeeId === e.id);
      if (isCovered(rows, day, hour)) n++;
    }
    return n;
  }

  // Color scale for the general heatmap (0 --> none, more --> deeper green).
  function heatColor(count, total) {
    if (count === 0) return 'rgba(248, 113, 113, 0.12)';
    const ratio = total > 0 ? count / total : 0;
    const alpha = 0.15 + ratio * 0.45;
    return `rgba(34, 197, 94, ${alpha.toFixed(3)})`;
  }

  async function handleSave(event) {
    const { employeeId, rows } = event.detail;
    saving = true;
    try {
      // No bulk endpoint: clear this employee's existing rows, then recreate.
      const existing = allAvail.filter(a => a.employeeId === employeeId);
      for (const a of existing) {
        await DeleteAvailability(a.id);
      }
      for (const r of rows) {
        await CreateAvailability({
          employeeId,
          dayOfWeek: r.dayOfWeek,
          startTime: r.startTime,
          endTime: r.endTime,
        });
      }
      await loadData();
    } catch (e) {
      console.error('Failed to save availability:', e);
    }
    saving = false;
    showModal = false;
  }

  loadData();
</script>

<div class="page">
  <h2>Availability</h2>

  <!-- Control bar -->
  <div class="control-bar">
    <div class="bar-left">
      {#if selectedEmployee}
        <button class="primary" on:click={() => (showModal = true)} disabled={saving}>
          {saving ? 'Saving…' : '✎ Edit Availability'}
        </button>
      {:else}
        <span class="view-hint">Showing combined availability across all employees</span>
      {/if}
    </div>

    <div class="bar-right">
      <label class="view-label" for="view-select">View</label>
      <select id="view-select" bind:value={selected}>
        <option value="general">General (all employees)</option>
        {#each employees as e}
          <option value={e.id}>{e.name}</option>
        {/each}
      </select>
    </div>
  </div>

  <!-- Legend -->
  <!-- <div class="legend" in:fade>
    {#if selectedEmployee}
      <span class="legend-item"><span class="swatch avail"></span> Available</span>
      <span class="legend-item"><span class="swatch unavail"></span> Unavailable</span>
    {:else}
      <span class="legend-label">Fewer</span>
      <span class="scale">
        {#each [0, 1, 2, 3, 4] as step}
          <span class="scale-cell" style="background: {heatColor(step, 4)}"></span>
        {/each}
      </span>
      <span class="legend-label">More available</span>
    {/if}
  </div> -->

  <!-- Grid -->
  {#if settings}
    <WeekGrid {daySettings} allowOutsideHours={settings.allowOutsideHours} rowHeight={32} let:day let:hour>
      {#if selectedEmployee}
        <div
          class="avail-fill"
          class:on={isCovered(selectedAvail, day, hour)}
          class:off={!isCovered(selectedAvail, day, hour)}
        ></div>
      {:else}
        {@const count = generalCount(day, hour)}
        <div
          class="heat-fill"
          style="background: {heatColor(count, employees.length)}"
          title="{count} available"
        >
          {#if count > 0}<span class="heat-num">{count}</span>{/if}
        </div>
      {/if}
    </WeekGrid>
  {/if}
</div>

<EditAvailabilityModal
  show={showModal}
  employee={selectedEmployee}
  availabilities={selectedAvail}
  on:save={handleSave}
  on:close={() => (showModal = false)}
/>

<style>
  .page { max-width: 880px; }

  .control-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.6rem 0.9rem;
    margin-bottom: 0.85rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
  }

  .bar-left { min-height: 32px; display: flex; align-items: center; }

  .view-hint {
    font-size: 0.8rem;
    color: rgba(255, 255, 255, 0.4);
  }

  .bar-right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .view-label {
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: rgba(255, 255, 255, 0.35);
  }

  /* Legend */
  .legend {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
    font-size: 0.74rem;
    color: rgba(255, 255, 255, 0.4);
  }

  .legend-item { display: flex; align-items: center; gap: 0.35rem; }

  .swatch {
    width: 12px;
    height: 12px;
    border-radius: 3px;
    display: inline-block;
  }
  .swatch.avail   { background: rgba(34, 197, 94, 0.55); }
  .swatch.unavail { background: rgba(248, 113, 113, 0.3); }

  .scale { display: flex; gap: 2px; }
  .scale-cell {
    width: 16px;
    height: 12px;
    border-radius: 2px;
  }

  .legend-label { font-size: 0.7rem; }

  /* Cell fills */
  .avail-fill {
    position: absolute;
    inset: 1px;
    border-radius: 2px;
  }
  .avail-fill.on  { background: rgba(34, 197, 94, 0.4); }
  .avail-fill.off { background: rgba(248, 113, 113, 0.12); }

  .heat-fill {
    position: absolute;
    inset: 1px;
    border-radius: 2px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .heat-num {
    font-size: 0.6rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.55);
  }
</style>