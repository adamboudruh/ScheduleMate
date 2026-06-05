<script>
  import { fly } from 'svelte/transition';
  import EmployeesPage from './pages/EmployeesPage.svelte';
  import AvailabilityPage from './pages/AvailabilityPage.svelte';
  import SchedulePage from './pages/SchedulePage.svelte';
  import SettingsPage from './pages/SettingsPage.svelte';

  let currentPage = 'employees';
  let drawerOpen = false;

  const navItems = [
    { id: 'employees', label: 'Employees' },
    { id: 'availability', label: 'Availability' },
    { id: 'schedule', label: 'Schedule' },
    { id: 'settings', label: 'Settings' },
  ];
</script>

<button class="hamburger" on:click={() => (drawerOpen = !drawerOpen)}>☰</button>

{#if drawerOpen}
  <div class="drawer" transition:fly={{ x: -220, duration: 200 }}>
    <strong>ScheduleMate</strong>
    <nav>
      {#each navItems as item}
        <button
          class:active={currentPage === item.id}
          on:click={() => { currentPage = item.id; drawerOpen = false; }}
        >
          {item.label}
        </button>
      {/each}
    </nav>
  </div>
{/if}

{#if currentPage === 'employees'}
  <EmployeesPage />
{:else if currentPage === 'availability'}
  <AvailabilityPage />
{:else if currentPage === 'schedule'}
  <SchedulePage />
{:else if currentPage === 'settings'}
  <SettingsPage />
{/if}

<style>
  .hamburger {
    position: fixed;
    top: 0.75rem;
    left: 0.75rem;
    background: none;
    border: none;
    color: white;
    font-size: 1.5rem;
    cursor: pointer;
    z-index: 300;
  }

  .drawer {
    position: fixed;
    top: 0;
    left: 0;
    width: 200px;
    height: 100vh;
    background: rgba(15, 25, 40, 0.88);
    backdrop-filter: blur(8px);
    z-index: 200;
    display: flex;
    flex-direction: column;
    padding: 1.5rem 1rem;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    margin-top: 50px;
  }

  nav button {
    background: none;
    border: none;
    color: rgba(255, 255, 255, 0.65);
    font-size: 1.3rem;
    cursor: pointer;
    text-align: left;
    padding: 0.6rem 0.9rem;
    border-radius: 4px;
    margin-top: 5px;
    margin-bottom: 30px;
    width: 100%;
  }

  nav button:hover {
    color: white;
    background: rgba(255, 255, 255, 0.06);
  }

  nav button.active {
    color: white;
    font-weight: bold;
    background: rgba(255, 255, 255, 0.12);
  }
</style>