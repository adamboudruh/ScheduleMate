<script>
  /**
   * StatsModal — displays computed stats for the current schedule:
   *  - payroll (pre-tax, shift-by-shift)
   *  - soft-score breakdown (mirrors the optimizer's score)
   *  - validity by category, with coverage gaps / issues listed
   *
   * All values are passed in already-computed (see lib/scheduleStats.js).
   */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let show = false;
  export let payroll = null;
  export let score = null;
  export let validity = null;

  function close() { dispatch('close'); }
  function onKeydown(e) { if (e.key === 'Escape') close(); }

  const CATEGORY_LABELS = {
    coverage: 'Coverage (demand met)',
    shiftLength: 'Shift lengths',
    maxHours: 'Max weekly hours',
    availability: 'Within availability',
    storeHours: 'Within store hours',
  };
</script>

{#if show}
  <div class="overlay" role="presentation" on:click|self={close} on:keydown={onKeydown}>
    <div class="modal" role="dialog">
      <div class="modal-header">
        <h3 class="modal-title">Schedule Stats</h3>
        {#if validity}
          <span class="verdict" class:ok={validity.valid} class:bad={!validity.valid}>
            {validity.valid ? '✓ Valid' : '✕ Invalid'}
          </span>
        {/if}
        <button class="close-btn" on:click={close} title="Close">✕</button>
      </div>

      <div class="modal-body">
        <!-- Payroll -->
        {#if payroll}
          <section>
            <div class="section-head">
              <h4>Payroll <span class="muted">· pre-tax</span></h4>
              <span class="big-num">${payroll.total.toFixed(2)}</span>
            </div>
            {#if payroll.perEmployee.length > 0}
              <div class="rows">
                {#each payroll.perEmployee as p}
                  <div class="row">
                    <span class="row-name">{p.name}</span>
                    <span class="row-detail">{p.hours}h × ${p.wage.toFixed(2)}</span>
                    <span class="row-val">${p.cost.toFixed(2)}</span>
                  </div>
                {/each}
              </div>
            {/if}
          </section>
        {/if}

        <!-- Soft score -->
        {#if score}
          <section>
            <div class="section-head">
              <h4>Optimization score <span class="muted">· lower is better</span></h4>
              <span class="big-num">{score.total.toFixed(1)}</span>
            </div>
            <div class="score-grid">
              <div class="score-cell">
                <span class="score-label">Hours gap</span>
                <span class="score-val">{score.hoursGap.toFixed(1)}</span>
                <span class="score-weight">×{score.weights.hoursGap}</span>
              </div>
              <div class="score-cell">
                <span class="score-label">Fairness</span>
                <span class="score-val">{score.fairness.toFixed(1)}</span>
                <span class="score-weight">×{score.weights.fairness}</span>
              </div>
              <div class="score-cell">
                <span class="score-label">Clopen</span>
                <span class="score-val">{score.clopenPenalty.toFixed(1)}</span>
                <span class="score-weight">×{score.weights.clopen}</span>
              </div>
              <div class="score-cell">
                <span class="score-label">Overstaff</span>
                <span class="score-val">{score.overstaffPenalty.toFixed(1)}</span>
                <span class="score-weight">×{score.weights.overstaff}</span>
              </div>
            </div>
            <div class="rows tight">
              {#each score.perEmployeeGap as g}
                <div class="row">
                  <span class="row-name">{g.name}</span>
                  <span class="row-detail">{g.actual.toFixed(0)}h / {g.desired}h wanted</span>
                  <span class="row-val" class:warn={g.gap > 0}>
                    {g.gap > 0 ? `${g.gap.toFixed(0)}h off` : 'on target'}
                  </span>
                </div>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Validity -->
        {#if validity}
          <section>
            <h4>Validity</h4>
            <div class="cat-grid">
              {#each Object.entries(validity.categories) as [key, ok]}
                <div class="cat" class:ok class:bad={!ok}>
                  <span class="cat-dot">{ok ? '✓' : '✕'}</span>
                  <span class="cat-label">{CATEGORY_LABELS[key] ?? key}</span>
                </div>
              {/each}
            </div>

            {#if validity.gaps.length > 0}
              <div class="warn-block">
                <span class="warn-head">Coverage gaps</span>
                <ul>
                  {#each validity.gaps.slice(0, 8) as g}<li>{g}</li>{/each}
                  {#if validity.gaps.length > 8}<li class="more">+{validity.gaps.length - 8} more…</li>{/if}
                </ul>
              </div>
            {/if}

            {#if validity.issues.length > 0}
              <div class="warn-block">
                <span class="warn-head">Other issues</span>
                <ul>
                  {#each validity.issues.slice(0, 8) as iss}<li>{iss}</li>{/each}
                  {#if validity.issues.length > 8}<li class="more">+{validity.issues.length - 8} more…</li>{/if}
                </ul>
              </div>
            {/if}
          </section>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed; inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(3px);
    display: flex; align-items: center; justify-content: center;
    z-index: 450;
    animation: fade 0.16s ease-out;
  }
  @keyframes fade { from { opacity: 0; } to { opacity: 1; } }

  .modal {
    width: 480px;
    max-width: calc(100vw - 2rem);
    max-height: calc(100vh - 3rem);
    display: flex; flex-direction: column;
    background: #0f1e30;
    border-radius: 10px;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.1);
    animation: pop 0.2s cubic-bezier(0.34, 1.4, 0.64, 1);
  }
  @keyframes pop { from { transform: scale(0.96); opacity: 0; } to { transform: scale(1); opacity: 1; } }

  .modal-header {
    display: flex; align-items: center; gap: 0.65rem;
    padding: 1rem 1.2rem 0.8rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  }
  .modal-title {
    margin: 0; font-size: 0.95rem; font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
    text-transform: none; letter-spacing: normal;
  }
  .verdict {
    font-size: 0.72rem; font-weight: 600;
    padding: 0.15rem 0.55rem; border-radius: 20px;
  }
  .verdict.ok  { background: rgba(34, 197, 94, 0.15); color: #4ade80; }
  .verdict.bad { background: rgba(248, 113, 113, 0.15); color: #f87171; }

  .close-btn {
    margin-left: auto;
    background: none; border: none; color: rgba(255, 255, 255, 0.35);
    font-size: 0.8rem; padding: 0.2rem 0.4rem; border-radius: 4px;
  }
  .close-btn:hover {
    background: rgba(255, 255, 255, 0.07) !important;
    color: rgba(255, 255, 255, 0.7) !important;
    border-color: transparent !important;
  }

  .modal-body { padding: 0.5rem 1.2rem 1.1rem; overflow-y: auto; }

  section {
    padding: 0.9rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  }
  section:last-child { border-bottom: none; }

  h4 {
    margin: 0 0 0.5rem;
    font-size: 0.82rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.75);
    text-transform: none;
    letter-spacing: normal;
  }
  .muted { color: rgba(255, 255, 255, 0.3); font-weight: 400; font-size: 0.75rem; }

  .section-head {
    display: flex; align-items: baseline; justify-content: space-between;
    margin-bottom: 0.6rem;
  }
  .section-head h4 { margin: 0; }

  .big-num {
    font-size: 1.25rem;
    font-weight: 700;
    color: #fff;
    font-variant-numeric: tabular-nums;
  }

  .rows { display: flex; flex-direction: column; gap: 0.2rem; }
  .rows.tight { margin-top: 0.6rem; }

  .row {
    display: flex; align-items: center;
    font-size: 0.8rem;
    padding: 0.25rem 0;
  }
  .row-name { flex: 0 0 30%; color: rgba(255, 255, 255, 0.75); font-weight: 500; }
  .row-detail { flex: 1; color: rgba(255, 255, 255, 0.4); }
  .row-val { color: rgba(255, 255, 255, 0.7); font-variant-numeric: tabular-nums; }
  .row-val.warn { color: #fbbf24; }

  .score-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
  }
  .score-cell {
    background: rgba(255, 255, 255, 0.04);
    border-radius: 6px;
    padding: 0.5rem;
    display: flex; flex-direction: column; align-items: center; gap: 0.1rem;
  }
  .score-label { font-size: 0.66rem; color: rgba(255, 255, 255, 0.4); }
  .score-val { font-size: 1.05rem; font-weight: 600; color: rgba(255, 255, 255, 0.85); }
  .score-weight { font-size: 0.6rem; color: rgba(255, 255, 255, 0.3); }

  .cat-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.35rem;
  }
  .cat {
    display: flex; align-items: center; gap: 0.4rem;
    font-size: 0.78rem;
    padding: 0.35rem 0.5rem;
    border-radius: 5px;
    background: rgba(255, 255, 255, 0.03);
  }
  .cat-dot { font-size: 0.72rem; }
  .cat.ok .cat-dot { color: #4ade80; }
  .cat.bad .cat-dot { color: #f87171; }
  .cat.ok .cat-label { color: rgba(255, 255, 255, 0.6); }
  .cat.bad .cat-label { color: rgba(255, 255, 255, 0.8); }

  .warn-block { margin-top: 0.7rem; }
  .warn-head {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: rgba(251, 191, 36, 0.8);
  }
  .warn-block ul {
    margin: 0.3rem 0 0;
    padding-left: 1.1rem;
  }
  .warn-block li {
    font-size: 0.76rem;
    color: rgba(255, 255, 255, 0.55);
    line-height: 1.5;
  }
  .warn-block li.more { color: rgba(255, 255, 255, 0.35); list-style: none; margin-left: -0.5rem; }
</style>