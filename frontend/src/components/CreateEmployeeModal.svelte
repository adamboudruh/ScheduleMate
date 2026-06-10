<script>
  export let show = false;
  export let title = '';
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let name = '';
  let role = '';
  let desiredHours = 20;
  let maxHours = 40;
  let wage = 15.00;

  function showModal(node) {
    node.showModal();
    return { destroy() { node.close(); } };
  }

  function submit() {
    if (!name.trim()) return;
    dispatch('submit', { name, role, desiredHours: Number(desiredHours), maxHours: Number(maxHours), wage: Number(wage) });
    reset();
  }

  function cancel() {
    dispatch('close');
    reset();
  }

  function reset() {
    show = false;
    name = ''; role = ''; desiredHours = 20; maxHours = 40; wage = 15.00;
  }

  function onKeydown(e) {
    if (e.key === 'Enter') submit();
    if (e.key === 'Escape') cancel();
  }
</script>

{#if show}
  <dialog use:showModal on:keydown={onKeydown}>
    <div class="modal-inner">
      <div class="modal-header">
        <h3 class="modal-title">{title}</h3>
        <button class="close-btn" on:click={cancel} title="Close">✕</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label for="emp-name">Name <span class="req">*</span></label>
          <input id="emp-name" bind:value={name} placeholder="Employee name" autocomplete="off" />
        </div>

        <div class="form-group">
          <label for="emp-role">Role</label>
          <input id="emp-role" bind:value={role} placeholder="e.g. Cashier, Stocker" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="emp-desired">Desired hours</label>
            <div class="input-suffix">
              <input id="emp-desired" type="number" bind:value={desiredHours} min="1" max="40" />
              <span>hrs</span>
            </div>
          </div>
          <div class="form-group">
            <label for="emp-max">Max hours</label>
            <div class="input-suffix">
              <input id="emp-max" type="number" bind:value={maxHours} min="1" max="80" />
              <span>hrs</span>
            </div>
          </div>
        </div>

        <div class="form-group">
          <label for="emp-wage">Hourly wage</label>
          <div class="input-prefix">
            <span>$</span>
            <input id="emp-wage" type="number" bind:value={wage} min="0" step="0.25" />
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button on:click={cancel}>Cancel</button>
        <button class="primary" on:click={submit} disabled={!name.trim()}>Add Employee</button>
      </div>
    </div>
  </dialog>
{/if}

<style>
  dialog {
    border: none;
    border-radius: 10px;
    padding: 0;
    background: #0f1e30;
    color: #e8edf2;
    box-shadow: 0 24px 60px rgba(0,0,0,0.6), 0 0 0 1px rgba(255,255,255,0.1);
    max-width: 400px;
    width: calc(100vw - 2rem);
    font-family: inherit;
    font-size: 14px;
  }

  dialog::backdrop {
    background: rgba(0,0,0,0.55);
    backdrop-filter: blur(3px);
  }

  dialog[open] {
    animation: modal-in 0.22s cubic-bezier(0.34, 1.4, 0.64, 1);
  }

  @keyframes modal-in {
    from { transform: scale(0.94) translateY(-8px); opacity: 0; }
    to   { transform: scale(1) translateY(0);       opacity: 1; }
  }

  .modal-inner {
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.2rem 0.75rem;
    border-bottom: 1px solid rgba(255,255,255,0.07);
  }

  .modal-title {
    margin: 0;
    font-size: 1.4rem;
    font-weight: 600;
    color: rgba(255,255,255,0.88);
    letter-spacing: -0.01em;
  }

  .close-btn {
    background: none;
    border: none;
    color: rgba(255,255,255,0.35);
    font-size: 1.4rem;
    padding: 0.2rem 0.4rem;
    cursor: pointer;
    border-radius: 4px;
  }

  .close-btn:hover {
    background: rgba(255,255,255,0.07) !important;
    color: rgba(255,255,255,0.7) !important;
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
    text-align: left;
  }

  .form-row {
    display: flex;
    gap: 0.85rem;
  }

  label {
    font-size: 0.75rem;
    color: rgba(255,255,255,0.45);
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .req { color: #f87171; }

  input {
    width: 100%;
  }

  .input-suffix,
  .input-prefix {
    display: flex;
    align-items: center;
    background: rgba(255,255,255,0.06);
    border: 1px solid rgba(255,255,255,0.12);
    border-radius: 5px;
    overflow: hidden;
  }

  .input-suffix input,
  .input-prefix input {
    background: none;
    border: none;
    flex: 1;
    padding: 0.4rem 0.5rem;
    color: #e8edf2;
    font-size: 0.875rem;
    outline: none;
    min-width: 0;
  }

  .input-suffix span,
  .input-prefix span {
    padding: 0 0.5rem;
    color: rgba(255,255,255,0.3);
    font-size: 0.8rem;
    white-space: nowrap;
    background: rgba(255,255,255,0.04);
    border-left: 1px solid rgba(255,255,255,0.08);
    align-self: stretch;
    display: flex;
    align-items: center;
  }

  .input-prefix span {
    border-left: none;
    border-right: 1px solid rgba(255,255,255,0.08);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.75rem 1.2rem 1rem;
    border-top: 1px solid rgba(255,255,255,0.07);
  }
</style>