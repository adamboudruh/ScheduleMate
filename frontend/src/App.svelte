<script>
  import { GetAllEmployees, CreateEmployee, UpdateEmployee, DeleteEmployee } from '../wailsjs/go/main/App.js'
  import { Seed, ClearData } from '../wailsjs/go/main/App.js'
  import CreateEmployeeModal from './CreateEmployeeModal.svelte';

  let employees = [];
  let showCreateEmployeeModal = false;
  
  async function handleSave(event) {
    await CreateEmployee(event.detail);
    showCreateEmployeeModal = false;
    loadEmployees();
  }

  async function loadEmployees() {
    employees = await GetAllEmployees();
  }

  async function seed() {
    await Seed();
    loadEmployees();
  }

  async function clear() {
    await ClearData();
    loadEmployees();
  }

  loadEmployees();
</script>

<main>
  <h2>Employees</h2>
  <button class="btn" on:click={seed}>Seed</button>
  <button class="btn" on:click={clear}>Clear</button>
  <button class="btn" on:click={() => showCreateEmployeeModal = true}>Add Employee</button>

  <table>
    <thead>
      <tr>
        <th>ID</th>
        <th>Name</th>
        <th>Role</th>
        <th>Desired Hours</th>
        <th>Wage</th>
      </tr>
    </thead>
    <tbody>
      {#each employees as emp}
        <tr>
          <td>{emp.id}</td>
          <td>{emp.name}</td>
          <td>{emp.role}</td>
          <td>{emp.desiredHours}</td>
          <td>{emp.wage}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</main>

<CreateEmployeeModal
  show={showCreateEmployeeModal}
  title="Create Employee"
  on:submit={handleSave}
  on:close={() => showCreateEmployeeModal = false}
/>

<style>

</style>