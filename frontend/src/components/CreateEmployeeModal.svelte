<script>
    export let show = false;
    export let title = '';
    import { createEventDispatcher } from 'svelte'; // how child components can send events to parent components
    const dispatch = createEventDispatcher();

    let name = '';
    let role = '';
    let desiredHours = 0;
    let maxHours = 0;
    let wage = 0;

    function showModal(node) {
        node.showModal();
        return { destroy() { node.close(); } };
    }

    const submit = () => {
        dispatch('submit', { name, role, desiredHours, maxHours, wage });
        name = ""; role = ""; desiredHours = 0; maxHours = 0; wage = 0;
        cancel();
    }

    const cancel = () => {
        dispatch('close');
        show = false;
        name = '';
        role = '';
        desiredHours = 0;
        maxHours = 0;
        wage = 0;
    }

</script>


{#if show}
    <dialog use:showModal>
        <div>
            <h3>{title}</h3>

            <label>Name</label>
            <input bind:value={name} placeholder="Employee name" />

            <label>Role</label>
            <input bind:value={role} placeholder="e.g. Cashier" />

            <label>Desired Hours</label>
            <input bind:value={desiredHours} type="number" min="1" max="40" />

            <label>Max Hours</label>
            <input bind:value={maxHours} type="number" min="1" max="40" />

            <label>Wage</label>
            <input bind:value={wage} type="number" min="0" step="0.01" />

            <div class="actions">
                <button on:click={submit}>Save</button>
                <button on:click={cancel}>Cancel</button>
            </div>
        </div>
    </dialog>
{/if}

<style>
	dialog {
        position: fixed;
        inset: 0;
        /* background: rgba(0,0,0,0.5); */
		max-width: 32em;
		border-radius: 0.2em;
		border: none;
		padding: 0;
        display: flex;
        align-items: center;
        justify-content: center;
	}
	dialog::backdrop {
		background: rgba(0, 0, 0, 0.3);
	}
	dialog > div {
		padding: 1em;
	}
	dialog[open] {
		animation: zoom 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
	}
	@keyframes zoom {
		from {
			transform: scale(0.95);
		}
		to {
			transform: scale(1);
		}
	}
	dialog[open]::backdrop {
		animation: fade 0.2s ease-out;
	}
	@keyframes fade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	button {
		display: block;
	}
</style>