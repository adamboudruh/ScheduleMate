<script>
  /** Static help/tips modal for the Schedule page. */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  export let show = false;
  function close() { dispatch('close'); }
  function onKeydown(e) { if (e.key === 'Escape') close(); }
</script>

{#if show}
  <div class="overlay" role="presentation" on:click|self={close} on:keydown={onKeydown}>
    <div class="modal" role="dialog">
      <div class="modal-header">
        <h3 class="modal-title">About scheduling</h3>
        <button class="close-btn" on:click={close} title="Close">✕</button>
      </div>
      <div class="modal-body">
        <div class="tip">
          <span class="tip-icon">✓</span>
          <p><strong>A valid schedule</strong> has no coverage gaps. The staffing demand you set per day in Settings is met for every open hour of the week.</p>
        </div>
        <div class="tip">
          <span class="tip-icon">⚠</span>
          <p><strong>If a valid schedule can't be generated</strong>, availability and demand can't be satisfied together. Try adding employee availability, lowering "employees needed" for a day, or widening store hours.</p>
        </div>
        <div class="tip">
          <span class="tip-icon">🎨</span>
          <p><strong>Each employee has a color.</strong> In grid view, overlapping shifts on the same day are shown side by side so you can see hand-offs.</p>
        </div>
        <div class="tip">
          <span class="tip-icon">📊</span>
          <p><strong>The Stats panel</strong> (click the summary line above the grid) shows weekly payroll, the optimizer's soft score, and a validity breakdown by category.</p>
        </div>
        <div class="tip">
          <span class="tip-icon">🌙</span>
          <p><strong>"Clopen"</strong> means too little rest between shifts on consecutive days (less than the minimum you set). The optimizer tries to avoid these.</p>
        </div>
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
    z-index: 500;
    animation: fade 0.16s ease-out;
  }
  @keyframes fade { from { opacity: 0; } to { opacity: 1; } }

  .modal {
    width: 440px;
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
    display: flex; align-items: center; justify-content: space-between;
    padding: 1rem 1.2rem 0.8rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  }
  .modal-title {
    margin: 0; font-size: 0.95rem; font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
    text-transform: none; letter-spacing: normal;
  }
  .close-btn {
    background: none; border: none; color: rgba(255, 255, 255, 0.35);
    font-size: 0.8rem; padding: 0.2rem 0.4rem; border-radius: 4px;
  }
  .close-btn:hover {
    background: rgba(255, 255, 255, 0.07) !important;
    color: rgba(255, 255, 255, 0.7) !important;
    border-color: transparent !important;
  }

  .modal-body { padding: 0.8rem 1.2rem 1.1rem; overflow-y: auto; }

  .tip {
    display: flex;
    gap: 0.65rem;
    padding: 0.55rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }
  .tip:last-child { border-bottom: none; }

  .tip-icon { font-size: 0.95rem; flex-shrink: 0; margin-top: 0.05rem; }

  .tip p {
    margin: 0;
    font-size: 0.82rem;
    line-height: 1.5;
    color: rgba(255, 255, 255, 0.6);
  }
  .tip strong { color: rgba(255, 255, 255, 0.82); font-weight: 600; }
</style>