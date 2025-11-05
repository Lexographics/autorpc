<script>
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { specStore, fetchSpec, selectMethod } from '$lib/stores/spec.svelte.js';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import MethodDetail from '$lib/components/MethodDetail.svelte';
	import SpecConfig from '$lib/components/SpecConfig.svelte';
	import '$lib/styles/components.css';

	let sidebarOpen = $state(false);

	function toggleSidebar() {
		sidebarOpen = !sidebarOpen;
	}

	function closeSidebar() {
		sidebarOpen = false;
	}

	function selectMethodFromHash() {
		if (!browser || !specStore.methods || specStore.methods.length === 0) {
			return;
		}

		const hash = window.location.hash.substring(1); // Remove #
		if (hash) {
			const method = specStore.methods.find((m) => m.name === hash);
			if (method) {
				selectMethod(method);
			} else {
				selectMethod(null);
			}
		}
	}

	onMount(() => {
		fetchSpec();
	});

	$effect(() => {
		if (specStore.methods && specStore.methods.length > 0 && browser) {
			selectMethodFromHash();
		}
	});
</script>

<svelte:head>
	<title>AutoRPC Introspection</title>
</svelte:head>

<div class="container">
	<Sidebar isOpen={sidebarOpen} onClose={closeSidebar} />
	
	<div class="main-content">
		<header class="header">
			<div class="header-left">
				<button class="menu-button" onclick={toggleSidebar} aria-label="Toggle menu">
					☰
				</button>
				<h1 class="header-title">Introspection</h1>
			</div>
			<SpecConfig />
		</header>

		{#if specStore.loading}
			<div class="loading">Loading spec...</div>
		{:else if specStore.error}
			<div class="error">Error: {specStore.error}</div>
		{:else}
			<MethodDetail />
		{/if}
	</div>
</div>
