<script>
  /**
   * AddShiftModal — manually add a single shift to the current schedule.
   * Just an employee, a day, and a start/end time. The parent supplies the
   * schedule id and persists via CreateShift.
   */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let show = false;
  export let employees = [];

  const DAY_NAMES = ['', 'Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const DAYS = [1, 2, 3, 4, 5, 6, 7];

  let employeeId = '';
  let dayOfWeek = 2;
  let startTime = '09:00';
  let endTime = '17:00';

  function tmin(t) {
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  $: error = (() => {
    if (!employeeId) return 'Pick an employee';
    if (!startTime || !endTime) return 'Set both times';
    if (tmin(startTime) >= tmin(endTime)) return 'Start must be before end';
    return null;
  })();

  function save() {
    if (error) return;
    dispatch('save', {
      employeeId: Number(employeeId),
      dayOfWeek,
      startTime,
      endTime,
    });
    reset();
  }

  function close() {
    dispatch('close');
    reset();
  }

  function reset() {
    employeeId = '';
    dayOfWeek = 2;
    startTime = '09:00';
    endTime = '17:00';
  }

  function onKeydown(e) {
    if (e.key === 'Escape') close();
    if (e.key === 'Enter') save();
  }
</script>

{#if show}
  <div class="overlay" role="presentation" on:click|self={close} on:keydown={onKeydown}>
    <div class="modal" role="dialog">
      <div class="modal-header">
        <h3 class="modal-title">Add Shift</h3>
        <button class="close-btn" on:click={close} title="Close">✕</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="shift-emp">Employee</label>
          <select id="shift-emp" bind:value={employeeId}>
            <option value="" disabled>Select employee…</option>
            {#each employees as e}
              <option value={e.id}>{e.name}</option>
            {/each}
          </select>
        </div>

        <div class="form-group">
          <label for="shift-day">Day</label>
          <select id="shift-day" bind:value={dayOfWeek}>
            {#each DAYS as d}
              <option value={d}>{DAY_NAMES[d]}</option>
            {/each}
          </select>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="shift-start">Start</label>
            <input id="shift-start" type="time" step="3600" bind:value={startTime} />
          </div>
          <div class="form-group">
            <label for="shift-end">End</label>
            <input id="shift-end" type="time" step="3600" bind:value={endTime} />
          </div>
        </div>

        {#if error && employeeId}
          <p class="error-line">{error}</p>
        {/if}
      </div>

      <div class="modal-footer">
        <button on:click={close}>Cancel</button>
        <button class="primary" on:click={save} disabled={error != null}>Add Shift</button>
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
    width: 380px;
    max-width: calc(100vw - 2rem);
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
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.2rem 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  }

  .modal-title {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.88);
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
    padding: 1rem 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    flex: 1;
  }

  .form-row { display: flex; gap: 0.85rem; }

  label {
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: rgba(255, 255, 255, 0.45);
    font-weight: 500;
  }

  select, input { width: 100%; }

  .error-line {
    margin: 0;
    font-size: 0.74rem;
    color: #f87171;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.75rem 1.2rem 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
  }
</style>