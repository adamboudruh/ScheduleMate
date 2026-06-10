<script>
  import { fly, fade } from 'svelte/transition';
  import EmployeesPage from './pages/EmployeesPage.svelte';
  import AvailabilityPage from './pages/AvailabilityPage.svelte';
  import SchedulePage from './pages/SchedulePage.svelte';
  import SettingsPage from './pages/SettingsPage.svelte';

  let currentPage = 'employees';

  const navItems = [
    { id: 'employees',   label: 'Employees',   icon: '👥' },
    { id: 'availability',label: 'Availability', icon: '📅' },
    { id: 'schedule',    label: 'Schedule',     icon: '🗓️' },
    { id: 'settings',    label: 'Settings',     icon: '⚙️' },
  ];
</script>

<div class="app-shell">
  <!-- Persistent sidebar -->
  <aside class="sidebar">
    <div class="brand">
      <span class="brand-icon">⏱</span>
      <span class="brand-name">ScheduleMate</span>
    </div>
    <nav>
      {#each navItems as item}
        <button
          class="nav-item"
          class:active={currentPage === item.id}
          on:click={() => (currentPage = item.id)}
        >
          <span class="nav-icon">{item.icon}</span>
          <span class="nav-label">{item.label}</span>
          {#if currentPage === item.id}
            <span class="active-bar" in:fade={{ duration: 150 }}></span>
          {/if}
        </button>
      {/each}
    </nav>
  </aside>

  <!-- Main content -->
  <main class="content">
    {#if currentPage === 'employees'}
      <EmployeesPage />
    {:else if currentPage === 'availability'}
      <AvailabilityPage />
    {:else if currentPage === 'schedule'}
      <SchedulePage />
    {:else if currentPage === 'settings'}
      <SettingsPage />
    {/if}
  </main>
</div>

<style>
  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    background: #0d1b2a;
    color: #e8edf2;
    font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
    font-size: 14px;
    line-height: 1.5;
  }

  :global(h2) {
    margin: 0 0 1.5rem 0;
    font-size: 1.4rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: #f0f4f8;
  }

  :global(h3) {
    margin: 0 0 0.85rem 0;
    font-size: 0.7rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: rgba(255,255,255,0.35);
  }

  :global(button) {
    cursor: pointer;
    font-family: inherit;
    font-size: 0.85rem;
    padding: 0.4rem 0.9rem;
    border-radius: 6px;
    border: 1px solid transparent;
    background: rgba(255,255,255,0.045);
    color: rgba(255,255,255,0.68);
    transition: background 0.15s, border-color 0.15s, color 0.15s;
  }

  :global(button:hover:not(:disabled)) {
    background: rgba(255,255,255,0.09);
    border-color: transparent;
    color: rgba(255,255,255,0.92);
  }

  :global(button:disabled) {
    opacity: 0.3;
    cursor: not-allowed;
  }

  :global(button.primary) {
    background: rgba(59,130,246,0.22);
    border-color: transparent;
    color: #93c5fd;
  }

  :global(button.primary:hover:not(:disabled)) {
    background: rgba(59,130,246,0.34);
    color: #bfdbfe;
  }

  :global(button.danger) {
    background: rgba(239, 68, 68, 0.1);
    border-color: transparent;
    color: #f87171;
  }

  :global(button.danger:hover:not(:disabled)) {
    background: rgba(239, 68, 68, 0.2);
    color: #fca5a5;
  }

  :global(input[type="text"]),
  :global(input[type="number"]),
  :global(input[type="time"]),
  :global(input[type="date"]),
  :global(select) {
    font-family: inherit;
    font-size: 0.875rem;
    background: rgba(255,255,255,0.06);
    border: 1px solid rgba(255,255,255,0.12);
    border-radius: 5px;
    color: #e8edf2;
    padding: 0.4rem 0.65rem;
    outline: none;
    transition: border-color 0.15s, background 0.15s;
  }

  :global(input:focus),
  :global(select:focus) {
    border-color: rgba(96,165,250,0.6);
    background: rgba(255,255,255,0.09);
  }

  :global(input:disabled),
  :global(select:disabled) {
    opacity: 0.3;
    cursor: not-allowed;
  }

  :global(input[type="checkbox"]) {
    width: 1rem;
    height: 1rem;
    accent-color: #2563eb;
    cursor: pointer;
  }

  /* The native dropdown popup inherits OS colors unless the options are styled,
     which on some platforms renders as white text on a white background. Force
     a dark, readable palette for the option list. */
  :global(select option),
  :global(select optgroup) {
    background: #0f1e30;
    color: #e8edf2;
  }

  .app-shell {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .sidebar {
    width: 188px;
    flex-shrink: 0;
    background: rgba(8, 18, 32, 0.9);
    border-right: 1px solid rgba(255,255,255,0.06);
    display: flex;
    flex-direction: column;
    padding: 1.25rem 0.75rem;
    gap: 0.25rem;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.25rem 0.6rem 1.25rem;
    border-bottom: 1px solid rgba(255,255,255,0.07);
    margin-bottom: 0.5rem;
  }

  .brand-icon {
    font-size: 1.2rem;
  }

  .brand-name {
    font-size: 0.95rem;
    font-weight: 600;
    color: rgba(255,255,255,0.88);
    letter-spacing: 0.01em;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .nav-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.65rem;
    width: 100%;
    text-align: left;
    padding: 0.55rem 0.75rem;
    border-radius: 6px;
    background: none;
    border: none;
    color: rgba(255,255,255,0.5);
    font-size: 0.875rem;
    font-weight: 400;
    transition: background 0.12s, color 0.12s;
    overflow: hidden;
  }

  .nav-item:hover {
    background: rgba(255,255,255,0.06) !important;
    border-color: transparent !important;
    color: rgba(255,255,255,0.82) !important;
  }

  .nav-item.active {
    background: rgba(37,99,235,0.18) !important;
    border-color: transparent !important;
    color: #93c5fd !important;
    font-weight: 500;
  }

  .nav-icon {
    font-size: 1rem;
    flex-shrink: 0;
  }

  .active-bar {
    position: absolute;
    left: 0;
    top: 20%;
    bottom: 20%;
    width: 3px;
    background: #3b82f6;
    border-radius: 0 2px 2px 0;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 2rem 2.5rem;
  }
</style>