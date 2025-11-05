class SpecStore {
	specUrl = $state('/spec.json');
	methods = $state([]);
	selectedMethod = $state(null);
	loading = $state(false);
	error = $state(null);

	async fetchSpec(url = null) {
		const currentUrl = url || this.specUrl;
		this.loading = true;
		this.error = null;

		try {
			const response = await fetch(currentUrl);
			if (!response.ok) {
				throw new Error(`Failed to fetch spec: ${response.statusText}`);
			}
			const data = await response.json();
			this.methods = data;
		} catch (err) {
			this.error = err.message;
			console.error('Error fetching spec:', err);
		} finally {
			this.loading = false;
		}
	}

	setSpecUrl(url) {
		this.specUrl = url;
		this.fetchSpec(url);
	}

	selectMethod(method) {
		this.selectedMethod = method;
	}
}

export const specStore = new SpecStore();

export function fetchSpec(url) {
	return specStore.fetchSpec(url);
}

export function setSpecUrl(url) {
	specStore.setSpecUrl(url);
}

export function selectMethod(method) {
	specStore.selectMethod(method);
}
