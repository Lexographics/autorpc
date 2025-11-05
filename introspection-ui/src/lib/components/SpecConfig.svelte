<script>
	import { specStore, setSpecUrl } from '$lib/stores/spec.svelte.js';
	import '$lib/styles/components.css';

	let showModal = $state(false);
	let inputUrl = $state('');

	function openModal() {
		inputUrl = specStore.specUrl || '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		inputUrl = '';
	}

	function handleSave() {
		if (inputUrl.trim()) {
			setSpecUrl(inputUrl.trim());
			closeModal();
		}
	}

	function handleKeydown(event) {
		if (event.key === 'Escape') {
			closeModal();
		} else if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
			handleSave();
		}
	}

	function handleModalClick(event) {
		if (event.target === event.currentTarget) {
			closeModal();
		}
	}
</script>

<button class="config-button" onclick={openModal}>
	Configure Spec URL
</button>

{#if showModal}
	<div 
		class="config-modal" 
		onclick={handleModalClick} 
		onkeydown={handleKeydown} 
		role="dialog" 
		tabindex="0"
		aria-label="Spec URL configuration modal"
	>
		<div class="config-modal-content">
			<h2 class="config-modal-title">Spec URL Configuration</h2>
			<div class="config-input-group">
				<label class="config-label" for="spec-url-input">Spec URL</label>
				<input
					id="spec-url-input"
					class="config-input"
					type="text"
					bind:value={inputUrl}
					placeholder="/spec.json"
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							handleSave();
						}
					}}
				/>
			</div>
			<div class="config-actions">
				<button class="config-button-secondary" onclick={closeModal}>Cancel</button>
				<button class="config-button-primary" onclick={handleSave}>Save</button>
			</div>
		</div>
	</div>
{/if}

