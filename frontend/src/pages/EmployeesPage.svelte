<script>
  import { slide, fade } from 'svelte/transition';
  import { GetAllEmployees, CreateEmployee, DeleteEmployee, UpdateEmployee } from '../../wailsjs/go/main/App.js';
  import { Seed, ClearData } from '../../wailsjs/go/main/App.js';
  import CreateEmployeeModal from '../components/CreateEmployeeModal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';

  let employees = [];
  let showModal = false;
  let expandedId = null; // id of employee whose card is expanded
  let search = ''; // search query for filtering employees by name
  let isEditing = false; // whether the currently expanded card is in edit mode
  let editData = {};
  let pendingDelete = null; // employee is queued for deletion and awaiting confirmation from modal

  function initials(name) {
    return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase();
  }

  function avatarHue(name) {
    let h = 0;
    for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) & 0xffff;
    return h % 360;
  }

  $: filtered = employees.filter(e => {
    const q = search.toLowerCase().trim();
    if (!q) return true;
    return e.name.toLowerCase().split(' ').some(part => part.startsWith(q));
  });

  async function load() { employees = await GetAllEmployees(); }

  async function handleCreate(event) {
    await CreateEmployee(event.detail);
    showModal = false;
    load();
  }

  function toggleExpand(emp) {
  if (expandedId === emp.id) {
    expandedId = null;
    isEditing = false;
    editData = {};
  } else {
    expandedId = emp.id;
    isEditing = false;
    editData = { ...emp }; // prime inputs with current values on expand
  }
}

  function startEdit() {
    isEditing = true;
  }

  function cancelEdit() {
    isEditing = false;
    const emp = employees.find(e => e.id === expandedId);
    if (emp) editData = { ...emp }; // reset to saved data
  }

  async function submitEdit() {
    const newEmployeeData = editData;
    const updatedEmployee = {
      ...newEmployeeData,
      desiredHours: Number(newEmployeeData.desiredHours),
      maxHours: Number(newEmployeeData.maxHours),
      wage: Number(newEmployeeData.wage),
    };

    await UpdateEmployee(updatedEmployee);

    // set employees to 
    employees = employees.map(e => (e.id === newEmployeeData.id ? { ...e, ...updatedEmployee } : e));
    if (expandedId === newEmployeeData.id) {
      editData = { ...updatedEmployee };
    }

    isEditing = false;
    await load();
  }

  function requestDelete(emp) {
    pendingDelete = emp;
  }
 
  async function confirmDelete() {
    const emp = pendingDelete;
    pendingDelete = null;
    if (!emp) return;
    await DeleteEmployee(emp.id);
    expandedId = null;
    load();
  }

  async function seed()  { await Seed();      load(); }

  load();
</script>

<div class="page">
  <div class="page-header">
    <h2>Employees</h2>
    <span class="count-badge">{filtered.length}</span>
  </div>

  <div class="toolbar">
    <div class="search-wrap">
      <span class="search-icon">🔍</span>
      <input bind:value={search} placeholder="Search by name…" />
    </div>
    <div class="btn-row">
      <button class="primary" on:click={() => (showModal = true)}>+ Add Employee</button>
      <button on:click={seed} title="Load sample data">Seed</button>
    </div>
  </div>

  <div class="list">
    {#each filtered as emp (emp.id)}
      {@const hue = avatarHue(emp.name)}
      {@const hoursPercent = Math.min(100, (emp.desiredHours / emp.maxHours) * 100)}
      <div class="card" class:expanded={expandedId === emp.id}>
        <button
          class="card-header"
          on:click={() => toggleExpand(emp)}
          aria-expanded={expandedId === emp.id}
        >
          <div class="avatar" style="--hue:{hue}">
            {initials(emp.name)}
          </div>
          <div class="card-info">
            <span class="emp-name">{emp.name}</span>
          </div>
          <span class="hours-summary">{emp.desiredHours}h / wk</span>
          <span class="chevron" class:open={expandedId === emp.id}>›</span>
        </button>

        {#if expandedId === emp.id}
          <div class="card-body" class:editing={isEditing} transition:slide={{ duration: 180 }}>
            <div class="detail-grid">
              {#if isEditing}
                <div class="detail-item" style="grid-column: 1 / -1;">
                  <span class="detail-label">Name</span>
                  <input class="inline-input" type="text" readonly={!isEditing} bind:value={editData.name} />
                </div>
              {/if}
              <div class="detail-item">
                <span class="detail-label">Desired hours</span>
                <input class="inline-input" type="number" min="0" readonly={!isEditing} bind:value={editData.desiredHours} />
              </div>
              <div class="detail-item">
                <span class="detail-label">Max hours</span>
                <input class="inline-input" type="number" min="0" readonly={!isEditing} bind:value={editData.maxHours} />
              </div>
              <div class="detail-item">
                <span class="detail-label">Wage ($/hr)</span>
                <input class="inline-input" type="number" min="0" step="0.01" readonly={!isEditing} bind:value={editData.wage} />
              </div>
              <div class="detail-item">
                <span class="detail-label">Role</span>
                <input class="inline-input" type="text" readonly={!isEditing} bind:value={editData.role} />
              </div>
            </div>
            <div class="card-actions">
              {#if isEditing}
                <button class="icon-action" on:click={cancelEdit} title="Cancel editing">✕</button>
                <button class="icon-action confirm" on:click={submitEdit} title="Save changes">✓</button>
              {:else}
                <button class="danger" on:click={() => requestDelete(emp)}>Delete employee</button>
                <button class="icon-action edit-btn" on:click={startEdit} title="Edit employee">✏️</button>
              {/if}
            </div>
          </div>
        {/if}
      </div>
    {/each}

    {#if filtered.length === 0}
      <div class="empty-state" in:fade>
        <span class="empty-icon">👤</span>
        <p>{search ? 'No employees match your search.' : 'No employees yet. Add one or seed sample data.'}</p>
      </div>
    {/if}
  </div>
</div>

<CreateEmployeeModal
  show={showModal}
  title="Add Employee"
  on:submit={handleCreate}
  on:close={() => (showModal = false)}
/>

<ConfirmModal
  show={pendingDelete !== null}
  title="Delete {pendingDelete?.name ?? 'employee'}?"
  message="This permanently removes {pendingDelete?.name ?? 'this employee'}, along with all of their scheduled shifts and saved availability. This cannot be undone."
  confirmLabel="Delete employee"
  danger={true}
  on:confirm={confirmDelete}
  on:cancel={() => (pendingDelete = null)}
/>

<style>
  .page { max-width: 620px; }

  .page-header {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    margin-bottom: 1.5rem;
  }

  .count-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: rgba(255,255,255,0.08);
    border-radius: 20px;
    font-size: 0.75rem;
    font-weight: 600;
    color: rgba(255,255,255,0.45);
    padding: 0.1rem 0.55rem;
    min-width: 1.5rem;
    margin-top: 0.1rem;
  }

  .toolbar {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    margin-bottom: 1.4rem;
  }

  .search-wrap {
    position: relative;
  }

  .search-icon {
    position: absolute;
    left: 0.65rem;
    top: 50%;
    transform: translateY(-50%);
    font-size: 0.85rem;
    pointer-events: none;
    opacity: 0.45;
  }

  .search-wrap input {
    width: 100%;
    padding-left: 2rem;
    font-size: 0.9rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
  }

  .btn-row {
    display: flex;
    gap: 0.5rem;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
  }

  .card {
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 8px;
    overflow: hidden;
    background: rgba(255,255,255,0.025);
    transition: border-color 0.15s;
  }

  .card.expanded {
    border-color: rgba(255,255,255,0.14);
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    text-align: left;
    padding: 0.65rem 0.9rem;
    background: none;
    border: none;
    color: inherit;
    font-size: inherit;
    cursor: pointer;
    border-radius: 0;
  }

  .card-header:hover {
    background: rgba(255,255,255,0.04) !important;
    border-color: transparent !important;
  }

  .avatar {
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    background: hsl(var(--hue), 45%, 30%);
    border: 1px solid hsl(var(--hue), 45%, 45%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
    color: hsl(var(--hue), 65%, 85%);
    letter-spacing: 0.02em;
  }

  .card-info {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
  }

  .emp-name {
    font-weight: 500;
    color: rgba(255,255,255,0.88);
    font-size: 0.9rem;
  }

  .role-badge {
    font-size: 0.7rem;
    font-weight: 500;
    padding: 0.1rem 0.45rem;
    border-radius: 20px;
    border: 1px solid;
    text-transform: capitalize;
    letter-spacing: 0.02em;
  }

  .hours-summary {
    font-size: 0.8rem;
    color: rgba(255,255,255,0.35);
    font-variant-numeric: tabular-nums;
    margin-right: 0.25rem;
  }

  .chevron {
    font-size: 1.1rem;
    color: rgba(255,255,255,0.3);
    transition: transform 0.18s;
    display: inline-block;
  }

  .chevron.open {
    transform: rotate(90deg);
  }

  .card-body {
    padding: 0.75rem 0.9rem 0.9rem;
    border-top: 1px solid rgba(255,255,255,0.07);
    background: rgba(0,0,0,0.12);
  }

  .detail-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.6rem 0rem;
    margin-bottom: 0.9rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .detail-label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: rgba(255,255,255,0.3);
  }

  .detail-value {
    font-size: 0.88rem;
    color: rgba(255,255,255,0.8);
    font-weight: 500;
  }

  .inline-input {
  width: 100%;
  font-size: 0.88rem;
  font-weight: 500;
  color: rgba(255,255,255,0.8);
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 0.15rem 0;
  font-family: inherit;
  cursor: default;
  pointer-events: none;
  -moz-appearance: textfield; /* hide number spinners at rest */
  text-align: center;
}

.inline-input::-webkit-outer-spin-button,
.inline-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.card-body.editing .inline-input {
  background: rgba(255,255,255,0.07);
  border-color: rgba(255,255,255,0.15);
  padding: 0.3rem 0.5rem;
  cursor: text;
  pointer-events: auto;
  -moz-appearance: auto;
  transition: background 0.15s, border-color 0.15s;
}

.card-body.editing .inline-input::-webkit-outer-spin-button,
.card-body.editing .inline-input::-webkit-inner-spin-button {
  -webkit-appearance: auto;
}

.card-body.editing .inline-input:focus {
  outline: none;
  border-color: rgba(255,255,255,0.35);
}

  .hours-bar-wrap {
    margin-bottom: 0.9rem;
  }

  .hours-bar-label {
    display: flex;
    justify-content: space-between;
    font-size: 0.7rem;
    color: rgba(255,255,255,0.35);
    margin-bottom: 0.3rem;
  }

  .hours-bar-track {
    height: 4px;
    background: rgba(255,255,255,0.08);
    border-radius: 2px;
    overflow: hidden;
  }

  .hours-bar-fill {
    height: 100%;
    background: hsl(var(--hue), 55%, 55%);
    border-radius: 2px;
    transition: width 0.3s;
  }

    .card-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .edit-btn {
    margin-left: auto;
  }

  .edit-input {
    width: 100%;
    font-size: 0.88rem;
    padding: 0.3rem 0.5rem;
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.15);
    border-radius: 4px;
    color: rgba(255,255,255,0.9);
  }

  .edit-input:focus {
    outline: none;
    border-color: rgba(255,255,255,0.35);
  }

  .icon-action {
    padding: 0.3rem 0.65rem;
    border-radius: 4px;
    font-size: 0.9rem;
    font-weight: 600;
    border: 1px solid rgba(255,255,255,0.12);
    background: rgba(255,255,255,0.05);
    color: rgba(255,255,255,0.55);
    cursor: pointer;
  }

  .icon-action:hover {
    background: rgba(255,255,255,0.09);
    color: rgba(255,255,255,0.85);
  }

  .icon-action.confirm {
    border-color: rgba(100,210,100,0.35);
    color: rgba(120,220,120,0.9);
  }

  .icon-action.confirm:hover {
    background: rgba(100,210,100,0.1);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 3rem 1rem;
    text-align: center;
    color: rgba(255,255,255,0.3);
  }

  .empty-icon {
    font-size: 2rem;
    opacity: 0.4;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.875rem;
  }
</style>