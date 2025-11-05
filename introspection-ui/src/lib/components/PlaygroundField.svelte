<script>
	import '$lib/styles/components.css';
	import PlaygroundField from './PlaygroundField.svelte';

	let { field, valuePath, getValue, setValue, getDefaultValue, getNonNullDefaultValue } = $props();

	function getCurrentValue() {
		return getValue(valuePath);
	}

	function setCurrentValue(newValue) {
		setValue(valuePath, newValue);
	}

	function getNestedValue(path) {
		return getValue([...valuePath, ...path]);
	}
</script>

<div class="playground-field">
	<label class="playground-label" for="playground-{valuePath.join('-')}">
		{field.name}
		{#if field.required}
			<span class="required-asterisk">*</span>
		{/if}
	</label>

	{#if field.isArray === true}
		<div class="playground-pointer-container">
			{#if field.isPointer === true}
				<div class="playground-pointer-null">
					<label class="playground-pointer-label">
						<input
							type="checkbox"
							checked={getCurrentValue() === null}
							onchange={(e) => {
								if (e.target.checked) {
									setCurrentValue(null);
								} else {
									setCurrentValue(getNonNullDefaultValue(field));
								}
							}}
						/>
						<span>Null</span>
					</label>
				</div>
			{/if}
			<div class="playground-array-container" class:disabled={getCurrentValue() === null}>
				{#each (Array.isArray(getCurrentValue()) ? getCurrentValue() : []) as item, index}
					<div class="playground-array-item">
						{#if field.fields}
							<div class="playground-nested-struct">
								{#each field.fields as nestedField}
									{@const nestedKey = nestedField.jsonName || nestedField.name}
									<PlaygroundField
										field={nestedField}
										valuePath={[...valuePath, index, nestedKey]}
										getValue={getValue}
										setValue={setValue}
										getDefaultValue={getDefaultValue}
										getNonNullDefaultValue={getNonNullDefaultValue}
									/>
								{/each}
							</div>
						{:else}
							{@const elementType = field.elementType || field.type || field.kind}
							{#if elementType === 'bool'}
								<input
									id="playground-{valuePath.join('-')}-{index}"
									class="playground-input"
									type="checkbox"
									checked={getNestedValue([index]) ?? false}
									disabled={getCurrentValue() === null}
									onchange={(e) => {
										const arr = Array.isArray(getCurrentValue()) ? [...getCurrentValue()] : [];
										arr[index] = e.target.checked;
										setCurrentValue(arr);
									}}
								/>
							{:else}
								<input
									id="playground-{valuePath.join('-')}-{index}"
									class="playground-input"
									type={elementType === 'int' || elementType === 'float32' ? 'number' : 'text'}
									value={getNestedValue([index]) ?? getDefaultValue({ kind: elementType })}
									disabled={getCurrentValue() === null}
									oninput={(e) => {
										const arr = Array.isArray(getCurrentValue()) ? [...getCurrentValue()] : [];
										arr[index] = e.target.value;
										setCurrentValue(arr);
									}}
									placeholder={elementType}
									step={elementType === 'float32' ? 'any' : undefined}
								/>
							{/if}
						{/if}
						<button
							type="button"
							class="playground-remove-button"
							disabled={getCurrentValue() === null}
							onclick={() => {
								const arr = Array.isArray(getCurrentValue()) ? [...getCurrentValue()] : [];
								arr.splice(index, 1);
								setCurrentValue(arr);
							}}
						>
							Remove
						</button>
					</div>
				{/each}
				<button
					type="button"
					class="playground-add-button"
					disabled={getCurrentValue() === null}
					onclick={() => {
						const defaultValue = getDefaultValue(field);
						const arr = Array.isArray(getCurrentValue()) ? [...getCurrentValue()] : [];
						if (field.fields && typeof defaultValue === 'object') {
							setCurrentValue([...arr, { ...defaultValue }]);
						} else {
							setCurrentValue([...arr, defaultValue]);
						}
					}}
				>
					Add
				</button>
			</div>
		</div>
	{:else if field.kind === 'struct' && field.fields}
		<div class="playground-nested-struct">
			{#each field.fields as nestedField}
				{@const nestedKey = nestedField.jsonName || nestedField.name}
				<PlaygroundField
					field={nestedField}
					valuePath={[...valuePath, nestedKey]}
					getValue={getValue}
					setValue={setValue}
					getDefaultValue={getDefaultValue}
					getNonNullDefaultValue={getNonNullDefaultValue}
				/>
			{/each}
		</div>
	{:else if field.isPointer === true}
		<div class="playground-pointer-container">
			<div class="playground-pointer-null">
				<label class="playground-pointer-label">
					<input
						type="checkbox"
						checked={getCurrentValue() === null}
						onchange={(e) => {
							if (e.target.checked) {
								setCurrentValue(null);
							} else {
								setCurrentValue(getNonNullDefaultValue(field));
							}
						}}
					/>
					<span>Null</span>
				</label>
			</div>
			{#if field.kind === 'bool'}
				<input
					id="playground-{valuePath.join('-')}"
					class="playground-input"
					type="checkbox"
					checked={getCurrentValue() ?? false}
					disabled={getCurrentValue() === null}
					onchange={(e) => {
						setCurrentValue(e.target.checked);
					}}
				/>
			{:else}
				<input
					id="playground-{valuePath.join('-')}"
					class="playground-input"
					type={field.kind === 'int' || field.kind === 'float32' ? 'number' : 'text'}
					value={getCurrentValue() ?? getDefaultValue(field)}
					disabled={getCurrentValue() === null}
					oninput={(e) => {
						setCurrentValue(e.target.value);
					}}
					placeholder={field.kind}
					step={field.kind === 'float32' ? 'any' : undefined}
				/>
			{/if}
		</div>
	{:else if field.kind === 'bool'}
		<input
			id="playground-{valuePath.join('-')}"
			class="playground-input"
			type="checkbox"
			checked={getCurrentValue() ?? false}
			onchange={(e) => {
				setCurrentValue(e.target.checked);
			}}
		/>
	{:else}
		<input
			id="playground-{valuePath.join('-')}"
			class="playground-input"
			type={field.kind === 'int' || field.kind === 'float32' ? 'number' : 'text'}
			value={getCurrentValue() ?? getDefaultValue(field)}
			oninput={(e) => {
				setCurrentValue(e.target.value);
			}}
			placeholder={field.kind}
			step={field.kind === 'float32' ? 'any' : undefined}
		/>
	{/if}
</div>

