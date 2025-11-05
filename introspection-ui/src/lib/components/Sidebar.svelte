<script>
	import { browser } from '$app/environment';
	import { specStore, selectMethod } from '$lib/stores/spec.svelte.js';
	import favicon from '$lib/assets/favicon.svg';
	import '$lib/styles/components.css';

	let { isOpen, onClose } = $props();

	let searchQuery = $state('');
	let expandedGroups = $state({});

	function handleSelect(method) {
		selectMethod(method);
		if (browser) {
			window.location.hash = method.name;
		}
		if (onClose) {
			onClose();
		}
	}

	function toggleGroup(groupName) {
		expandedGroups = {
			...expandedGroups,
			[groupName]: !isGroupExpanded(groupName)
		};
	}

	function isGroupExpanded(groupName) {
		if (!expandedGroups || !groupName) return true;
		return expandedGroups[groupName] !== false;
	}

	function toggleAllGroups() {
		const groups = groupedMethods;
		if (!groups || groups.length === 0) return;

		const allExpanded = groups.every(group => isGroupExpanded(group.groupName));

		const newExpanded = {};
		for (const group of groups) {
			newExpanded[group.groupName] = !allExpanded;
		}
		expandedGroups = newExpanded;
	}

	function getGroupName(methodName) {
		const dotIndex = methodName.indexOf('.');
		return dotIndex > 0 ? methodName.substring(0, dotIndex) : 'default';
	}

	let filteredMethods = $derived(searchQuery.trim() === ''
		? specStore.methods
		: specStore.methods.filter((method) =>
				method.name.toLowerCase().includes(searchQuery.toLowerCase())
		  ));

	let groupedMethods = $derived.by(() => {
		const groups = {};
		
		for (const method of filteredMethods) {
			const groupName = getGroupName(method.name);
			if (!groups[groupName]) {
				groups[groupName] = [];
			}
			groups[groupName].push(method);
		}
		
		return Object.entries(groups)
			.map(([groupName, methods]) => ({
				groupName,
				methods: methods.sort((a, b) => a.name.localeCompare(b.name))
			}))
			.sort((a, b) => a.groupName.localeCompare(b.groupName));
	});

	let allGroupsExpanded = $derived.by(() => {
		const groups = groupedMethods;
		if (!groups || groups.length === 0) return true;
		return groups.every(group => isGroupExpanded(group.groupName));
	});

	$effect(() => {
		const groups = groupedMethods;
		if (groups && groups.length > 0) {
			const currentKeys = Object.keys(expandedGroups);
			if (currentKeys.length === 0) {
				const newExpanded = {};
				for (const group of groups) {
					newExpanded[group.groupName] = true;
				}
				expandedGroups = newExpanded;
			}
		}
	});
</script>

{#if isOpen}
	<div 
		class="sidebar-overlay" 
		class:open={isOpen} 
		onclick={onClose}
		onkeydown={(e) => e.key === 'Enter' && onClose()}
		role="button" 
		tabindex="0"
		aria-label="Close sidebar"
	></div>
{/if}
<aside class="sidebar" class:open={isOpen}>
	<div class="sidebar-header">
		<h2 class="sidebar-title" style="display: flex; align-items: center; gap: 0.5rem;">
			<img src={favicon} alt="Brand" style="width: 2rem; height: 2rem;" />
			AutoRPC
		</h2>
		<button
			class="collapse-all-button"
			onclick={toggleAllGroups}
			aria-label={allGroupsExpanded ? 'Collapse all groups' : 'Expand all groups'}
			title={allGroupsExpanded ? 'Collapse all' : 'Expand all'}
		>
			{allGroupsExpanded ? '-' : '+'}
		</button>
	</div>
	<div class="sidebar-search">
		<input
			type="text"
			class="search-input"
			placeholder="Search methods..."
			bind:value={searchQuery}
			aria-label="Search methods"
		/>
	</div>
	<div class="sidebar-list">
		{#each groupedMethods as { groupName, methods } (groupName)}
			<div class="method-group">
				<button
					class="method-group-header"
					onclick={() => toggleGroup(groupName)}
					aria-expanded={isGroupExpanded(groupName)}
				>
					<span class="group-name">{groupName}</span>
					<span class="group-icon" class:expanded={isGroupExpanded(groupName)}>
						▼
					</span>
				</button>
				{#if isGroupExpanded(groupName)}
					<div class="method-group-items">
						{#each methods as method (method.name)}
							<button
								class="sidebar-item"
								class:selected={specStore.selectedMethod?.name === method.name}
								onclick={() => handleSelect(method)}
							>
								{method.name}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/each}
		{#if filteredMethods.length === 0 && specStore.methods.length > 0}
			<div class="empty-state">No methods match your search</div>
		{:else if specStore.methods.length === 0}
			<div class="empty-state">No methods available</div>
		{/if}
	</div>
</aside>

