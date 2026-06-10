<script>
  /** Small reusable confirmation dialog. Driven by props; emits confirm/cancel.
   *  Set requireText to force the user to type a phrase before confirming. */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let show = false;
  export let title = 'Are you sure?';
  export let message = '';
  export let confirmLabel = 'Confirm';
  export let cancelLabel = 'Cancel';
  export let danger = false;
  export let requireText = ''; // if set, user must type this exact phrase to enable confirm

  let typed = '';
  let wasShown = false;
  // Reset the typed phrase each time the dialog opens.
  $: if (show && !wasShown) { typed = ''; wasShown = true; }
  $: if (!show) wasShown = false;

  $: canConfirm = !requireText || typed.trim().toLowerCase() === requireText.toLowerCase();

  function confirm() { if (canConfirm) dispatch('confirm'); }
  function cancel()  { dispatch('cancel'); }

  function onKeydown(e) {
    if (e.key === 'Escape') cancel();
    if (e.key === 'Enter' && canConfirm && !requireText) confirm();
  }
</script>

{#if show}
  <div class="overlay" role="presentation" on:click|self={cancel} on:keydown={onKeydown}>
    <div class="modal" role="alertdialog">
      <h3 class="title">{title}</h3>
      {#if message}<p class="message">{message}</p>{/if}
      {#if requireText}
        <input
          class="confirm-input"
          bind:value={typed}
          placeholder={`Type "${requireText}" to confirm`}
          autocomplete="off"
          spellcheck="false"
        />
      {/if}
      <div class="actions">
        <button on:click={cancel}>{cancelLabel}</button>
        <button class:primary={!danger} class:danger on:click={confirm} disabled={!canConfirm}>{confirmLabel}</button>
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
    width: 360px;
    max-width: calc(100vw - 2rem);
    background: #0f1e30;
    border-radius: 10px;
    padding: 1.2rem 1.3rem 1.1rem;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.1);
    animation: pop 0.2s cubic-bezier(0.34, 1.4, 0.64, 1);
  }
  @keyframes pop { from { transform: scale(0.95); opacity: 0; } to { transform: scale(1); opacity: 1; } }

  .title {
    margin: 0 0 0.5rem;
    font-size: 0.98rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
    text-transform: none;
    letter-spacing: normal;
  }

  .message {
    margin: 0 0 1.1rem;
    font-size: 0.85rem;
    line-height: 1.5;
    color: rgba(255, 255, 255, 0.6);
  }

  .confirm-input {
    width: 100%;
    margin: 0 0 1.1rem;
    box-sizing: border-box;
    font-size: 0.85rem;
    padding: 0.45rem 0.6rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
</style>