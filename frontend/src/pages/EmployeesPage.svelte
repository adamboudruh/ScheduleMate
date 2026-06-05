<script>
  import { slide } from 'svelte/transition';
  import { GetAllEmployees, CreateEmployee, DeleteEmployee } from '../../wailsjs/go/main/App.js';
  import { Seed, ClearData } from '../../wailsjs/go/main/App.js';
  import CreateEmployeeModal from '../components/CreateEmployeeModal.svelte';

  let employees = [];
  let showModal = false;   // 
  let expandedId = null;   // which employee's details are expanded
  let search = '';

  $: filtered = employees.filter(e => {
    const q = search.toLowerCase().trim();
    if (!q) return true;
    const parts = e.name.toLowerCase().split(' ');
    return parts.some(part => part.startsWith(q));
  });

  async function load() {
    employees = await GetAllEmployees();
  }

  async function handleCreate(event) {
    await CreateEmployee(event.detail);
    showModal = false;
    load();
  }

  async function handleDelete(id) {
    await DeleteEmployee(id);
    expandedId = null;
    load();
  }

  async function seed() {
    await Seed();
    load();
  }

  async function clear() {
    await ClearData();
    load();
  }

  load();
</script>

<div class="page">

<h2>Employees</h2>

<div class="toolbar">
  <input bind:value={search} placeholder="Search by name..." />
  <div class="btn-row">
    <button on:click={() => (showModal = true)}>Add Employee</button>
    <button on:click={seed}>Seed</button>
    <button on:click={clear}>Clear</button>
  </div>
</div>

<div class="list">
  {#each filtered as emp (emp.id)}
    <div class="row">
      <div class="row-header" on:click={() => (expandedId = expandedId === emp.id ? null : emp.id)}>
        <span>{emp.name}</span>
        <span class="chevron">{expandedId === emp.id ? '▲' : '▼'}</span>
      </div>

      {#if expandedId === emp.id}
        <div class="details" transition:slide={{ duration: 200 }}>
          <p><strong>Role:</strong> {emp.role || '—'}</p>
          <p><strong>Desired Hours:</strong> {emp.desiredHours}</p>
          <p><strong>Max Hours:</strong> {emp.maxHours}</p>
          <p><strong>Wage:</strong> ${emp.wage?.toFixed(2)}</p>
          <div class="actions">
            <button on:click={() => handleDelete(emp.id)}>Delete</button>
          </div>
        </div>
      {/if}
    </div>
  {/each}

  {#if filtered.length === 0}
    <p style="opacity: 0.5;">No employees found.</p>
  {/if}
</div>

</div>

<CreateEmployeeModal
  show={showModal}
  title="Add Employee"
  on:submit={handleCreate}
  on:close={() => (showModal = false)}
/>

<style>
  .page {
    max-width: 680px;
    margin: 0 auto;
    padding: 1rem 2rem;
  }

  .toolbar {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    margin-bottom: 1.25rem;
  }

  .toolbar input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.55rem 0.9rem;
    border-radius: 4px;
    border: 1px solid rgba(255,255,255,0.2);
    background: rgba(255,255,255,0.07);
    color: white;
    font-size: 1.05rem;
  }

  .btn-row {
    display: flex;
    gap: 0.5rem;
    justify-content: center;
  }

  .btn-row button {
    padding: 0.35rem 1.1rem;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    width: 100%;
  }

  .row {
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: 6px;
    overflow: hidden;
  }

  .row-header {
    display: flex;
    justify-content: space-between;
    padding: 0.6rem 0.9rem;
    cursor: pointer;
    user-select: none;
  }

  .row-header:hover {
    background: rgba(255,255,255,0.05);
  }

  .chevron {
    opacity: 0.5;
    font-size: 0.75rem;
  }

  .details {
    padding: 0.6rem 0.9rem 0.8rem;
    border-top: 1px solid rgba(255,255,255,0.08);
    font-size: 0.9rem;
  }

  .details p {
    margin: 0.25rem 0;
  }

  .actions {
    margin-top: 0.75rem;
  }
</style>
