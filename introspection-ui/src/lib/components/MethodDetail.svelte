<script>
	import { specStore } from '$lib/stores/spec.svelte.js';
	import PlaygroundField from './PlaygroundField.svelte';
	import '$lib/styles/components.css';

	let playgroundParams = $state({});
	let playgroundResult = $state(null);
	let playgroundError = $state(null);
	let playgroundLoading = $state(false);

	function getValue(path) {
		let current = playgroundParams;
		for (const key of path) {
			if (current == null || current[key] === undefined) {
				return undefined;
			}
			current = current[key];
		}
		return current;
	}

	function setValue(path, value) {
		const newParams = { ...playgroundParams };
		let current = newParams;
		for (let i = 0; i < path.length - 1; i++) {
			const key = path[i];
			if (current[key] == null || typeof current[key] !== 'object') {
				current[key] = typeof path[i + 1] === 'number' ? [] : {};
			}
			if (Array.isArray(current[key])) {
				current[key] = [...current[key]];
			} else {
				current[key] = { ...current[key] };
			}
			current = current[key];
		}
		current[path[path.length - 1]] = value;
		playgroundParams = newParams;
	}

	function formatType(type) {
		if (!type) return 'void';
		
		if (type.isArray) {
			return `${type.elementType}[]`;
		}
		
		if (type.kind === 'struct' && type.fields) {
			return type.name || 'struct';
		}
		
		return type.name || type.kind || 'unknown';
	}

	function formatSchema(schema) {
		if (!schema) return null;
		
		if (schema.isArray === true) {
			return {
				type: 'array',
				elementType: schema.elementType || schema.kind,
				kind: schema.kind
			};
		}
		
		if (schema.kind === 'struct' && schema.fields) {
			return {
				type: 'struct',
				name: schema.name,
				fields: schema.fields
			};
		}
		
		return {
			type: 'primitive',
			name: schema.name,
			kind: schema.kind
		};
	}

	function getDefaultValue(field) {
		if (field.isPointer === true) {
			return null;
		}
		
		if (field.isArray === true) {
			return [];
		}
		if (field.kind === 'struct' && field.fields) {
			const structValue = {};
			field.fields.forEach((f) => {
				structValue[f.jsonName || f.name] = getDefaultValue(f);
			});
			return structValue;
		}
		if (field.kind === 'int') {
			return 0;
		} else if (field.kind === 'float32') {
			return 0.0;
		} else if (field.kind === 'string') {
			return '';
		} else if (field.kind === 'bool') {
			return false;
		}
		return null;
	}

	function getNonNullDefaultValue(field) {
		if (field.isArray === true) {
			return [];
		}
		if (field.kind === 'struct' && field.fields) {
			const structValue = {};
			field.fields.forEach((f) => {
				structValue[f.jsonName || f.name] = getNonNullDefaultValue(f);
			});
			return structValue;
		}
		if (field.kind === 'int') {
			return 0;
		} else if (field.kind === 'float32') {
			return 0.0;
		} else if (field.kind === 'string') {
			return '';
		} else if (field.kind === 'bool') {
			return false;
		}
		return null;
	}

	function initializePlaygroundParams() {
		if (!specStore.selectedMethod?.params) {
			playgroundParams = {};
			return;
		}

		const schema = formatSchema(specStore.selectedMethod.params);
		const newParams = {};

		if (schema.type === 'struct' && schema.fields) {
			schema.fields.forEach((field) => {
				const key = field.jsonName || field.name;
				if (!newParams[key]) {
					newParams[key] = getDefaultValue(field);
				}
			});
		} else if (schema.type === 'array') {
			newParams.value = [];
		} else if (schema.type === 'primitive') {
			if (schema.kind === 'int') {
				newParams.value = 0;
			} else if (schema.kind === 'float32') {
				newParams.value = 0.0;
			} else if (schema.kind === 'string') {
				newParams.value = '';
			} else if (schema.kind === 'bool') {
				newParams.value = false;
			} else {
				newParams.value = null;
			}
		}

		playgroundParams = newParams;
		playgroundResult = null;
		playgroundError = null;
	}

	$effect(() => {
		if (specStore.selectedMethod) {
			initializePlaygroundParams();
		}
	});

	function parseValue(value, kind, isPointer = false) {
		if (value === null || value === undefined) {
			return null;
		}
		
		if (kind === 'string') {
			return value;
		}

		if (value === '') {
			return null;
		}

		if (kind === 'int') {
			const parsed = parseInt(value, 10);
			return isNaN(parsed) ? null : parsed;
		} else if (kind === 'float32') {
			const parsed = parseFloat(value);
			return isNaN(parsed) ? null : parsed;
		} else if (kind === 'bool') {
			if (typeof value === 'boolean') return value;
			if (typeof value === 'string') {
				return value.toLowerCase() === 'true' || value === '1';
			}
			return Boolean(value);
		}

		return value;
	}

	function buildFieldValue(field, value) {
		if (field.isPointer === true && value === null) {
			return null;
		}

		if (field.isArray === true) {
			if (field.isPointer === true && value === null) {
				return null;
			}
			if (!Array.isArray(value)) {
				if (field.isPointer === true) {
					return null;
				}
				return [];
			}
			return value.map((item) => {
				if (field.kind === 'struct' && field.fields) {
					return buildStructValue(field, item);
				}
				return parseValue(item, field.elementType || field.kind, false);
			}).filter((v) => v !== null && v !== undefined);
		}
		if (field.kind === 'struct' && field.fields) {
			return buildStructValue(field, value || {});
		}
		return parseValue(value, field.kind, field.isPointer === true);
	}

	function buildStructValue(field, value) {
		if (!field.fields) return {};
		const struct = {};
		field.fields.forEach((f) => {
			const key = f.jsonName || f.name;
			const fieldValue = value?.[key];
			if (f.isPointer === true && fieldValue === null) {
				struct[key] = null;
			} else {
				struct[key] = buildFieldValue(f, fieldValue);
			}
		});
		return struct;
	}

	function buildRpcParams() {
		if (!specStore.selectedMethod?.params) {
			return null;
		}

		const schema = formatSchema(specStore.selectedMethod.params);

		if (schema.type === 'struct' && schema.fields) {
			const params = {};
			schema.fields.forEach((field) => {
				const key = field.jsonName || field.name;
				const value = playgroundParams[key];
				params[key] = buildFieldValue(field, value);
			});
			return params;
		} else if (schema.type === 'array') {
			const value = playgroundParams.value;
			if (!Array.isArray(value)) {
				return [];
			}
			return value.map((item) => parseValue(item, schema.elementType)).filter((v) => v !== null && v !== undefined);
		} else if (schema.type === 'primitive') {
			return parseValue(playgroundParams.value, schema.kind);
		}

		return null;
	}

	async function executeRpc() {
		if (!specStore.selectedMethod) return;

		playgroundLoading = true;
		playgroundResult = null;
		playgroundError = null;

		try {
			const params = buildRpcParams();
			const request = {
				jsonrpc: '2.0',
				method: specStore.selectedMethod.name,
				id: Date.now()
			};
			
			if (params !== null && params !== undefined) {
				request.params = params;
			}

			const response = await fetch('/rpc', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(request)
			});

			const data = await response.json();

			if (data.error) {
				playgroundError = data.error;
			} else {
				playgroundResult = data.result;
			}
		} catch (err) {
			playgroundError = {
				code: -32603,
				message: 'Internal error',
				data: err.message
			};
		} finally {
			playgroundLoading = false;
		}
	}
</script>

{#if specStore.selectedMethod}
	<div class="method-detail">
		<h1 class="method-name">{specStore.selectedMethod.name}</h1>

		{#if specStore.selectedMethod.params}
			{@const schema = formatSchema(specStore.selectedMethod.params)}
			<div class="section">
				<h2 class="section-title">Parameters</h2>
				
				{#if schema.type === 'struct'}
					<div class="type-info">
						<div class="type-name">
							{schema.name}
							<span class="type-kind">struct</span>
						</div>
						{#if schema.fields && schema.fields.length > 0}
							<ul class="field-list">
								{#each schema.fields as field}
									<li class="field-item">
										<div class="field-name">
											{field.name}
											{#if field.required}
												<span class="required-asterisk">*</span>
											{/if}
											{#if field.jsonName && field.jsonName !== field.name}
												<span class="field-json-name">({field.jsonName})</span>
											{/if}
										</div>
										<div class="field-type">{field.type}</div>
										<div class="field-meta">
											{#if field.validationRules && field.validationRules.length > 0}
												{#each field.validationRules.filter(rule => rule.trim() !== 'required') as rule}
													<span class="field-tag">{rule}</span>
												{/each}
											{/if}
										</div>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				{:else if schema.type === 'array'}
					<div class="type-info">
						<div class="type-name">
							Array
							<span class="type-kind">array</span>
						</div>
						<div class="field-type">Element type: {schema.elementType}</div>
					</div>
				{:else}
					<div class="type-info">
						<div class="type-name">
							{schema.name || schema.kind}
							<span class="type-kind">{schema.kind}</span>
						</div>
					</div>
				{/if}
			</div>
		{/if}

		{#if specStore.selectedMethod.result}
			{@const resultSchema = formatSchema(specStore.selectedMethod.result)}
			<div class="section">
				<h2 class="section-title">Result</h2>
				
				{#if resultSchema.type === 'struct'}
					<div class="type-info">
						<div class="type-name">
							{resultSchema.name}
							<span class="type-kind">struct</span>
						</div>
						{#if resultSchema.fields && resultSchema.fields.length > 0}
							<ul class="field-list">
								{#each resultSchema.fields as field}
									<li class="field-item">
										<div class="field-name">
											{field.name}
											{#if field.jsonName && field.jsonName !== field.name}
												<span class="field-json-name">({field.jsonName})</span>
											{/if}
										</div>
										<div class="field-type">{field.type}</div>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				{:else if resultSchema.type === 'array'}
					<div class="type-info">
						<div class="type-name">
							Array
							<span class="type-kind">array</span>
						</div>
						<div class="field-type">Element type: {resultSchema.elementType}</div>
					</div>
				{:else}
					<div class="type-info">
						<div class="type-name">
							{resultSchema.name || resultSchema.kind}
							<span class="type-kind">{resultSchema.kind}</span>
						</div>
					</div>
				{/if}
			</div>
		{/if}

		<div class="section">
			<h2 class="section-title">Playground</h2>
			{#if specStore.selectedMethod.params}
				{@const schema = formatSchema(specStore.selectedMethod.params)}
				<div class="playground">
					<form class="playground-form" onsubmit={(e) => { e.preventDefault(); executeRpc(); }}>
						{#if schema.type === 'struct' && schema.fields}
							{#each schema.fields as field}
								{@const fieldKey = field.jsonName || field.name}
								<PlaygroundField
									field={field}
									valuePath={[fieldKey]}
									getValue={getValue}
									setValue={setValue}
									getDefaultValue={getDefaultValue}
									getNonNullDefaultValue={getNonNullDefaultValue}
								/>
							{/each}
						{:else if schema.type === 'array'}
							<PlaygroundField
								field={{
									name: 'Array',
									kind: specStore.selectedMethod.params.elementType || specStore.selectedMethod.params.kind,
									isArray: true,
									elementType: specStore.selectedMethod.params.elementType || specStore.selectedMethod.params.kind,
									fields: specStore.selectedMethod.params.fields,
									isPointer: specStore.selectedMethod.params.isPointer
								}}
								valuePath={['value']}
								getValue={getValue}
								setValue={setValue}
								getDefaultValue={getDefaultValue}
								getNonNullDefaultValue={getNonNullDefaultValue}
							/>
						{:else if schema.type === 'primitive'}
							<PlaygroundField
								field={{
									name: 'Value',
									kind: specStore.selectedMethod.params.kind,
									isPointer: specStore.selectedMethod.params.isPointer
								}}
								valuePath={['value']}
								getValue={getValue}
								setValue={setValue}
								getDefaultValue={getDefaultValue}
								getNonNullDefaultValue={getNonNullDefaultValue}
							/>
						{/if}
						<button type="submit" class="playground-button" disabled={playgroundLoading}>
							{playgroundLoading ? 'Executing...' : 'Execute'}
						</button>
					</form>

					{#if playgroundResult !== null}
						<div class="playground-result">
							<h3 class="playground-result-title">Result</h3>
							<pre class="playground-result-content">{JSON.stringify(playgroundResult, null, 2)}</pre>
						</div>
					{/if}

					{#if playgroundError}
						<div class="playground-error">
							<h3 class="playground-error-title">Error</h3>
							<div class="playground-error-content">
								<div class="playground-error-code">Code: {playgroundError.code}</div>
								<div class="playground-error-message">Message: {playgroundError.message}</div>
								{#if playgroundError.data}
									<div class="playground-error-data">Data: {JSON.stringify(playgroundError.data, null, 2)}</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{:else}
				<div class="empty-state">This method has no parameters</div>
			{/if}
		</div>
	</div>
{:else}
	<div class="method-detail">
		<div class="empty-state">Select a method to view details</div>
	</div>
{/if}

