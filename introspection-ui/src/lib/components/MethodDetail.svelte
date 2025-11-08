<script>
	import { specStore, getTypeInfo } from '$lib/stores/spec.svelte.js';
	import { onMount, onDestroy } from 'svelte';
	import loader from '@monaco-editor/loader';
	import '$lib/styles/components.css';

	let editorContainer = $state(null);
	let editor = $state(null);
	let editorValue = $state('');
	// initializing to undefined is important since rpc result can be null
	let playgroundResult = $state(undefined);
	let playgroundError = $state(null);
	let playgroundLoading = $state(false);
	let jsonError = $state(null);

	function generateExampleJSON(typeInfo) {
		if (!typeInfo) {
			return null;
		}

		const schema = formatSchema(typeInfo);
		console.log(schema);
		
		if (schema.type === 'struct') {
			if (schema.fields && schema.fields.length > 0) {
				const json = {};
				schema.fields.forEach((field) => {
					const resolvedField = getFieldTypeInfo(field);
					const key = resolvedField.jsonName || resolvedField.name;
					json[key] = generateFieldValue(resolvedField);
				});
				return json;
			}
			return {};
		} else if (schema.type === 'custom') {
			return getPrimitiveDefaultValue(schema.kind);
		} else if (schema.type === 'array') {
			return [];
		} else if (schema.type === 'map') {
			return {};
		} else if (schema.type === 'primitive') {
			return getPrimitiveDefaultValue(schema.kind);
		}
		
		
		return null;
	}

	function generateFieldValue(field) {
		const fieldInfo = getFieldTypeInfo(field);
		
		if (fieldInfo.fields && fieldInfo.fields.length > 0) {
			const structValue = {};
			fieldInfo.fields.forEach((f) => {
				const resolvedField = getFieldTypeInfo(f);
				structValue[resolvedField.jsonName || resolvedField.name] = generateFieldValue(resolvedField);
			});
			return structValue;
		}
		
		if (fieldInfo.package && fieldInfo.name && fieldInfo.kind !== 'map' && fieldInfo.kind !== 'struct') {
			return getPrimitiveDefaultValue(fieldInfo.kind);
		}
		
		if (fieldInfo.isPointer === true || fieldInfo.pointerDepth > 0) {
			return null;
		}
		
		if (fieldInfo.isArray === true || fieldInfo.arrayDepth > 0) {
			return [];
		}
		
		if (fieldInfo.kind === 'map') {
			return {};
		}
		
		return getPrimitiveDefaultValue(fieldInfo.kind);
	}

	function getPrimitiveDefaultValue(kind) {
		if (kind === 'int' || kind === 'int8' || kind === 'int16' || 
		    kind === 'int32' || kind === 'int64' ||
		    kind === 'uint' || kind === 'uint8' || kind === 'uint16' ||
		    kind === 'uint32' || kind === 'uint64') {
			return 0;
		} else if (kind === 'float32' || kind === 'float64') {
			return 0.0;
		} else if (kind === 'string') {
			return '';
		} else if (kind === 'bool') {
			return false;
		}
		return null;
	}

	function formatType(type) {
		if (!type) return 'void';
		
		if (type.isArray && type.arrayDepth > 0) {
			const arrayPrefix = '[]'.repeat(type.arrayDepth);
			return `${arrayPrefix}${type.name || type.kind}`;
		}
		
		if (type.kind === 'struct' && type.fields) {
			return type.name || 'struct';
		}
		
		return type.name || type.kind || 'unknown';
	}

	function formatSchema(schema) {
		if (!schema) return null;
		
		if (schema.isArray === true && schema.arrayDepth > 0) {
			return {
				type: 'array',
				arrayDepth: schema.arrayDepth,
				kind: schema.kind,
				name: schema.name
			};
		}
		
		if (schema.kind === 'map') {
			return {
				type: 'map',
				keyType: schema.keyType,
				valueType: schema.valueType,
				name: schema.name
			};
		}
		
		if (schema.fields && schema.fields.length > 0) {
			return {
				type: 'struct',
				name: schema.name,
				package: schema.package,
				fields: schema.fields
			};
		}
		
		
		// primitve types do not have a package
		if (schema.package && schema.name) {
			return {
				type: 'custom',
				name: schema.name,
				package: schema.package,
				kind: schema.kind
			};
		}
		
		return {
			type: 'primitive',
			name: schema.name,
			kind: schema.kind
		};
	}

	function getMethodTypeInfo(typeName) {
		if (!typeName) return null;
		return getTypeInfo(typeName, specStore.types);
	}

	function getFieldTypeInfo(field) {
		if (!field.type) return field;
		
		const arrayDepth = field.arrayDepth || 0;
		const pointerDepth = field.pointerDepth || 0;
		
		const baseTypeInfo = specStore.types[field.type];
		if (!baseTypeInfo) {
			return field;
		}
		
		return {
			...baseTypeInfo,
			...field,

			name: field.name,
			jsonName: field.jsonName,
			required: field.required,
			validationRules: field.validationRules,
			kind: baseTypeInfo.kind || field.kind,
			fields: baseTypeInfo.fields != null ? baseTypeInfo.fields : field.fields,
			keyType: baseTypeInfo.keyType || field.keyType,
			valueType: baseTypeInfo.valueType || field.valueType,
			isArray: arrayDepth > 0 || baseTypeInfo.isArray,
			arrayDepth: arrayDepth,
			isPointer: pointerDepth > 0 || baseTypeInfo.isPointer,
			pointerDepth: pointerDepth
		};
	}


	function initializeEditor() {
		if (!specStore.selectedMethod?.params) {
			editorValue = '';
			if (editor) {
				editor.setValue('');
			}
			return;
		}

		const typeInfo = getMethodTypeInfo(specStore.selectedMethod.params);
		if (!typeInfo) {
			editorValue = '';
			if (editor) {
				editor.setValue('');
			}
			return;
		}

		const exampleJSON = generateExampleJSON(typeInfo);
		const jsonString = JSON.stringify(exampleJSON, null, 2);
		editorValue = jsonString;
		
		if (editor) {
			editor.setValue(jsonString);
		}
		
		playgroundResult = undefined;
		playgroundError = null;
		jsonError = null;
	}

	function validateJSON(text) {
		try {
			JSON.parse(text);
			jsonError = null;
			return true;
		} catch (e) {
			jsonError = e.message;
			return false;
		}
	}

	$effect(() => {
		if (specStore.selectedMethod) {
			initializeEditor();
		}
	});

	let monacoInstance = $state(null);

	onMount(async () => {
		monacoInstance = await loader.init();
		tryInitEditor();
	});

	async function tryInitEditor() {
		if (!monacoInstance || !editorContainer || editor) return;

		editor = monacoInstance.editor.create(editorContainer, {
			value: editorValue,
			language: 'json',
			theme: 'vs-dark',
			automaticLayout: false,
			minimap: { enabled: false },
			formatOnPaste: true,
			formatOnType: true
		});

		editor.onDidChangeModelContent(() => {
			const value = editor.getValue();
			editorValue = value;
			validateJSON(value);
		});

		// CTRL+Enter executes the procedure
		editor.addCommand(monacoInstance.KeyMod.CtrlCmd | monacoInstance.KeyCode.Enter, () => {
			if (!playgroundLoading && !jsonError) {
				executeRpc();
			}
		});

		if (specStore.selectedMethod) {
			initializeEditor();
		}
	}

	$effect(() => {
		if (monacoInstance && editorContainer && !editor) {
			setTimeout(() => {
				tryInitEditor();
			}, 0);
		}
	});

	$effect(() => {
		if (editor && specStore.selectedMethod) {
			initializeEditor();
		}
	});

	onDestroy(() => {
		if (editor) {
			editor.dispose();
		}
	});


	function getRpcParams() {
		if (!editorValue) {
			return null;
		}

		try {
			const json = JSON.parse(editorValue);
			return json;
		} catch (e) {
			return null;
		}
	}

	async function executeRpc() {
		if (!specStore.selectedMethod) return;

		if (!validateJSON(editorValue)) {
			playgroundError = {
				code: -32700,
				message: 'Parse error',
				data: jsonError
			};
			return;
		}

		playgroundLoading = true;
		playgroundResult = undefined;
		playgroundError = null;

		try {
			const params = getRpcParams();
			const request = {
				jsonrpc: '2.0',
				method: specStore.selectedMethod.name,
				id: Date.now(),
				params: params
			};

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

		<div class="section">
			<h2 class="section-title">Playground</h2>
			{#if specStore.selectedMethod.params}
				<div class="playground">
					<div class="playground-editor-container">
						<div bind:this={editorContainer} class="playground-editor"></div>
						{#if jsonError}
							<div class="playground-error">
								JSON Error: {jsonError}
							</div>
						{/if}
					</div>
					<div class="playground-actions">
						<button 
							type="button" 
							class="playground-button" 
							disabled={playgroundLoading || !!jsonError}
							onclick={executeRpc}
						>
							{playgroundLoading ? 'Executing...' : 'Execute (Ctrl+Enter)'}
						</button>
					</div>
					
					{#if playgroundError}
						<div class="playground-error-result">
							<h3>Error</h3>
							<pre>{JSON.stringify(playgroundError, null, 2)}</pre>
						</div>
					{/if}
					
					{#if playgroundResult !== undefined}
						<div class="playground-result">
							<h3>Result</h3>
							<pre>{JSON.stringify(playgroundResult, null, 2)}</pre>
						</div>
					{/if}
				</div>
			{:else}
				<div class="playground">
					<p>This method has no parameters.</p>
					<div class="playground-actions">
						<button 
							type="button" 
							class="playground-button" 
							disabled={playgroundLoading}
							onclick={executeRpc}
						>
							{playgroundLoading ? 'Executing...' : 'Execute (Ctrl+Enter)'}
						</button>
					</div>
					
					{#if playgroundError}
						<div class="playground-error-result">
							<h3>Error</h3>
							<pre>{JSON.stringify(playgroundError, null, 2)}</pre>
						</div>
					{/if}
					
					{#if playgroundResult !== null}
						<div class="playground-result">
							<h3>Result</h3>
							<pre>{JSON.stringify(playgroundResult, null, 2)}</pre>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		{#if specStore.selectedMethod.params}
			{@const typeInfo = getMethodTypeInfo(specStore.selectedMethod.params)}
			{@const schema = formatSchema(typeInfo)}
			<div class="section">
				<h2 class="section-title">Parameters</h2>
				
				{#if schema.type === 'struct'}
					<div class="type-info">
						<div class="type-name">
							{schema.name}
							{#if schema.package}
								<span class="type-package">({schema.package})</span>
							{/if}
							<span class="type-kind">struct</span>
						</div>
						{#if schema.fields && schema.fields.length > 0}
							<ul class="field-list">
								{#each schema.fields as field}
									{@const resolvedField = getFieldTypeInfo(field)}
									<li class="field-item">
										<div class="field-name">
											{resolvedField.name}
											{#if resolvedField.required}
												<span class="required-asterisk">*</span>
											{/if}
											{#if resolvedField.jsonName && resolvedField.jsonName !== resolvedField.name}
												<span class="field-json-name">({resolvedField.jsonName})</span>
											{/if}
										</div>
										<div class="field-type">
											{#if resolvedField.arrayDepth > 0}
												{'[]'.repeat(resolvedField.arrayDepth)}
											{/if}
											{#if resolvedField.pointerDepth > 0}
												{'*'.repeat(resolvedField.pointerDepth)}
											{/if}
											{#if resolvedField.kind === 'map' && resolvedField.keyType && resolvedField.valueType}
												map[{resolvedField.keyType}]{resolvedField.valueType}
											{:else}
												{resolvedField.type}
											{/if}
										</div>
										<div class="field-meta">
											{#if resolvedField.validationRules && resolvedField.validationRules.length > 0}
												{#each resolvedField.validationRules.filter(rule => rule.trim() !== 'required') as rule}
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
						<div class="field-type">
							Array depth: {schema.arrayDepth}, Element type: {schema.name || schema.kind}
						</div>
					</div>
				{:else if schema.type === 'map'}
					<div class="type-info">
						<div class="type-name">
							Map
							<span class="type-kind">map</span>
						</div>
						<div class="field-type">
							map[{schema.keyType}]{schema.valueType}
						</div>
					</div>
				{:else if schema.type === 'custom'}
					<div class="type-info">
						<div class="type-name">
							{schema.name}
							{#if schema.package}
								<span class="type-package">({schema.package})</span>
							{/if}
							<span class="type-kind">{schema.kind}</span>
						</div>
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
			{@const resultTypeInfo = getMethodTypeInfo(specStore.selectedMethod.result)}
			{@const resultSchema = formatSchema(resultTypeInfo)}
			<div class="section">
				<h2 class="section-title">Result</h2>
				
				{#if resultSchema.type === 'struct'}
					<div class="type-info">
						<div class="type-name">
							{resultSchema.name}
							{#if resultSchema.package}
								<span class="type-package">({resultSchema.package})</span>
							{/if}
							<span class="type-kind">struct</span>
						</div>
						{#if resultSchema.fields && resultSchema.fields.length > 0}
							<ul class="field-list">
								{#each resultSchema.fields as field}
									{@const resolvedField = getFieldTypeInfo(field)}
									<li class="field-item">
										<div class="field-name">
											{resolvedField.name}
											{#if resolvedField.jsonName && resolvedField.jsonName !== resolvedField.name}
												<span class="field-json-name">({resolvedField.jsonName})</span>
											{/if}
										</div>
										<div class="field-type">
											{#if resolvedField.arrayDepth > 0}
												{'[]'.repeat(resolvedField.arrayDepth)}
											{/if}
											{#if resolvedField.pointerDepth > 0}
												{'*'.repeat(resolvedField.pointerDepth)}
											{/if}
											{#if resolvedField.kind === 'map' && resolvedField.keyType && resolvedField.valueType}
												map[{resolvedField.keyType}]{resolvedField.valueType}
											{:else}
												{resolvedField.type}
											{/if}
										</div>
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
						<div class="field-type">
							Array depth: {resultSchema.arrayDepth}, Element type: {resultSchema.name || resultSchema.kind}
						</div>
					</div>
				{:else if resultSchema.type === 'map'}
					<div class="type-info">
						<div class="type-name">
							Map
							<span class="type-kind">map</span>
						</div>
						<div class="field-type">
							map[{resultSchema.keyType}]{resultSchema.valueType}
						</div>
					</div>
				{:else if resultSchema.type === 'custom'}
					<div class="type-info">
						<div class="type-name">
							{resultSchema.name}
							{#if resultSchema.package}
								<span class="type-package">({resultSchema.package})</span>
							{/if}
							<span class="type-kind">{resultSchema.kind}</span>
						</div>
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

	</div>
{:else}
	<div class="method-detail">
		<div class="empty-state">Select a method to view details</div>
	</div>
{/if}

