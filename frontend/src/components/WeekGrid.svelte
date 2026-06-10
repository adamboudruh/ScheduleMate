<script>
  /**
   * WeekGrid, shared 7-column week layout used by both the Availability and
   * Schedule pages. It owns ONLY the layout: the time axis, the day columns,
   * closed-day shading, and the store open/close boundary markers.
   * 
   * Looks complicated but at it's core, it's a CSS grid. First row is dedicated to day labels,
   * first column is dedicaated to time labels (shifted up so it looks like it's on the line).
   * The amount of rows (hours on the grid) is determined just by the day of the week with the most
   * schedulable hours. Each cell (a combination of day and start hour) is given a state based on 
   * whether it its respective day is closed, and whether the cell is outside store hours, within 
   * store hours but outside schedulable hours, or fully schedulable.
   * 
   * Slot object is passed in from parent and will contain what's actually displayed there.
   * References: https://gridbyexample.com/examples/
   */

  export let daySettings = {};        // { 1: {dayOfWeek, openTime, closeTime, schedulableOpen, schedulableClose, employeesNeeded}, ... }
  export let allowOutsideHours = false;
  export let rowHeight = 40;          // px per hour row

  const DAY_SHORT = ['', 'Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  const DAYS = [1, 2, 3, 4, 5, 6, 7];

  function hourToMinutes(t) {
    if (!t) return null;
    const [h, m] = t.split(':').map(Number);
    return h * 60 + m;
  }

  // Resolve each day's effective schedulable window, mirroring
  // resolveSchedulableHours: when allowOutsideHours is off (or schedulable
  // times are blank) the schedulable window falls back to open/close.
  // Effectively creates definitive "store hours" for the grid to render
  $: resolved = (() => {
    const out = {}; // output -> map of dayOfWeek to { closed (bool), openMin, closeMin, schedOpenMin, schedCloseMin }
    for (const d of DAYS) {
      const ds = daySettings[d]; // grab day settings for this day
      if (!ds) { out[d] = null; continue; } // no settings means its closed
      let storeOpen = ds.schedulableOpen;
      let storeClose = ds.schedulableClose;
      if (!allowOutsideHours || !storeOpen) storeOpen = ds.openTime;
      if (!allowOutsideHours || !storeClose) storeClose = ds.closeTime;
      out[d] = {
        closed: ds.employeesNeeded === 0,
        openMin: hourToMinutes(ds.openTime),
        closeMin: hourToMinutes(ds.closeTime),
        schedOpenMin: hourToMinutes(storeOpen),
        schedCloseMin:hourToMinutes(storeClose),
      };
    }
    return out;
  })();

  // Rows go from earliest sched open to latest sched close across all days
  $: range = (() => { // { start, end } in minutes after midnight
    let start = 100000, end = -100000; // crazy number
    for (const d of DAYS) {
      const r = resolved[d];
      if (!r || r.closed) continue;
      if (r.schedOpenMin < start) start = r.schedOpenMin;
      if (r.schedCloseMin > end)  end = r.schedCloseMin;
    }
    if (start === Infinity) { start = 9 * 60; end = 21 * 60; } // all closed → sane default
    return { start, end };
  })();

  $: hours = (() => { // an array of each hour mark that should be rendered in the grid, used for iterating rows
    const arr = [];
    for (let t = range.start; t < range.end; t += 60) arr.push(t);
    return arr;
  })();

  function cellState(day, hour) { // returns what part of the day this cell is in
    const resolvedHours = resolved[day];
    let cellstate = "";
    if (!resolvedHours || resolvedHours.closed) cellstate = 'closed';
    if (hour < resolvedHours.schedOpenMin || hour >= resolvedHours.schedCloseMin) cellstate = 'outside'; // outside of schedulable hours for this day, happens if another day is extending the grid's range
    if (hour < resolvedHours.openMin)  cellstate = 'preopen';
    if (hour >= resolvedHours.closeMin) cellstate = 'postclose';
    let state = cellstate || 'open';
    console.log(`cellState for day ${day}, hour ${hour}: ${state}`);
    return state;
  }

  function openBoundary(day, hour) { // returns true if this cell is the first hour of the schedulable window
    const resolvedHours = resolved[day];
    if (!resolvedHours || resolvedHours.closed) return false;
    return hour === resolvedHours.openMin && resolvedHours.schedOpenMin < resolvedHours.openMin;
  }
  function closeBoundary(day, hour) { // returns true if this cell is the last hour of the schedulable window
    const resolvedHours = resolved[day];
    if (!resolvedHours || resolvedHours.closed) return false;
    return hour === resolvedHours.closeMin && resolvedHours.closeMin < resolvedHours.schedCloseMin;
  }

  function minutesToHourString(min) { // formats minutes after midnight to readable hour
    let h = Math.floor(min / 60) % 24;
    const period = h < 12 ? 'AM' : 'PM';
    let h12 = h % 12;
    if (h12 === 0) h12 = 12;
    return `${h12} ${period}`;
  }
</script>

{#if hours.length === 0}
  <p class="grid-loading">Loading grid…</p>
{:else}
  <div
    class="week-grid"
    style="--row-h:{rowHeight}px; grid-template-columns: 54px repeat(7, minmax(0, 1fr));"
  >
    <!-- header row -->
    <div class="corner"></div> <!-- empty corner cell -->
    {#each DAYS as day}
      <div class="day-head" class:closed={resolved[day]?.closed}> 
        <span class="day-name">{DAY_SHORT[day]}</span>
        {#if resolved[day]?.closed}
          <span class="closed-tag">Closed</span>
        {/if}
      </div>
    {/each}

    <!-- Hour rows -->
    {#each hours as hour, i} <!-- go hour by hour of the day, save a cell for the time, and then make 7 day cells-->
      <div class="time-axis">
        <span class="time-label">{minutesToHourString(hour)}</span>
        {#if i === hours.length - 1}
          <span class="time-label end">{minutesToHourString(range.end)}</span>
        {/if}
      </div>

      {#each DAYS as day}
        {@const st = cellState(day, hour)}
        <div
          class="cell state-{st}"
          class:open-bound={openBoundary(day, hour)}
          class:close-bound={closeBoundary(day, hour)}
          style="height: var(--row-h);"
        >
          {#if st === 'open' || st === 'preopen' || st === 'postclose'}
            <slot {day} {hour} state={st} /> <!-- slot is defined in the parent component -->
          {/if}

          {#if openBoundary(day, hour)}
            <span class="bound-label">open</span>
          {/if}
          {#if closeBoundary(day, hour)}
            <span class="bound-label">close</span>
          {/if}
        </div>
      {/each}
    {/each}
  </div>
{/if}

<style>
  .grid-loading {
    color: rgba(255, 255, 255, 0.35);
    font-size: 0.875rem;
  }

  .week-grid {
    display: grid;
    width: 100%;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.015);
    padding-top: 6px; /* room for the first axis label that straddles the top line */
  }

  /* Header */
  .corner {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .day-head {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1px;
    padding: 0.4rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    border-left: 1px solid rgba(255, 255, 255, 0.06);
  }

  .day-name {
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.65);
  }

  .day-head.closed .day-name {
    color: rgba(255, 255, 255, 0.3);
  }

  .closed-tag {
    font-size: 0.58rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: rgba(248, 113, 113, 0.7);
  }

  /* Time axis */
  .time-axis {
    position: relative;
    border-right: 1px solid rgba(255, 255, 255, 0.08);
  }

  /* little time label all the way to the left */
  .time-label {
    position: absolute;
    top: -0.55em; /* make it sit right on the light */
    right: 7px;
    font-size: 0.62rem;
    color: rgba(255, 255, 255, 0.3);
    white-space: nowrap;
  }

  .time-label.end {
    top: auto;
    bottom: -0.55em;
    z-index: 10;
  }

  /* Cells */
  .cell {
    position: relative;
    border-left: 1px solid rgba(255, 255, 255, 0.06);
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    overflow: visible; /* lets spanning blobs overflow downward */
  }

  /* Schedulable + within store hours: the normal interactive area */
  .cell.state-open {
    background: transparent;
  }

  /* schedulable but outside store hours (only with allowOutsideHours) */
  /* do a cool stripe thing */
  .cell.state-preopen,
  .cell.state-postclose {
    background: repeating-linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.015),
      rgba(255, 255, 255, 0.015) 5px,
      rgba(255, 255, 255, 0.04) 5px,
      rgba(255, 255, 255, 0.04) 10px
    );
  }

  /* Not schedulable this day (day's window is narrower than the global range) */
  .cell.state-outside {
    background: rgba(0, 0, 0, 0.22);
  }

  /* Whole day closed */
  .cell.state-closed {
    background: repeating-linear-gradient(
      45deg,
      rgba(0, 0, 0, 0.25),
      rgba(0, 0, 0, 0.25) 6px,
      rgba(248, 113, 113, 0.04) 6px,
      rgba(248, 113, 113, 0.04) 12px
    );
  }

  /* Boundary markers */
  .cell.open-bound { border-top: 2px solid rgba(74, 222, 128, 0.5); }
  .cell.close-bound { border-top: 2px solid rgba(248, 113, 113, 0.5); }

  .bound-label {
    position: absolute;
    top: 1px;
    left: 2px;
    z-index: 4;
    font-size: 0.52rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 0 3px;
    border-radius: 3px;
    background: rgba(13, 27, 42, 0.85);
    color: rgba(255, 255, 255, 0.5);
    pointer-events: none;
  }
</style>