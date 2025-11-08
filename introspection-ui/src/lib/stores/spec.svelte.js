class SpecStore {
	specUrl = $state('/spec.json');
	spec = $state({ methods: [], types: {} });
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
			this.spec = {
				methods: data.methods || [],
				types: data.types || {}
			};
		} catch (err) {
			this.error = err.message;
			console.error('Error fetching spec:', err);
		} finally {
			this.loading = false;
		}
	}

	get methods() {
		return this.spec.methods;
	}

	get types() {
		return this.spec.types;
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

const PRIMITIVE_TYPES = new Set([
	'bool', 'int', 'int8', 'int16', 'int32', 'int64',
	'uint', 'uint8', 'uint16', 'uint32', 'uint64',
	'float32', 'float64', 'string'
]);

function createPrimitiveTypeInfo(typeName) {
	const kind = typeName;
	return {
		name: typeName,
		package: '',
		kind: kind,
		isArray: false,
		arrayDepth: 0,
		isPointer: false,
		pointerDepth: 0,
		elementType: '',
		fields: []
	};
}

export function getTypeInfo(typeName, types) {
	if (!typeName) return null;

	let arrayDepth = 0;
	let pointerDepth = 0;
	let baseTypeName = typeName;

	while (baseTypeName.startsWith('[]')) {
		arrayDepth++;
		baseTypeName = baseTypeName.substring(2);
	}

	while (baseTypeName.startsWith('*')) {
		pointerDepth++;
		baseTypeName = baseTypeName.substring(1);
	}

	if (PRIMITIVE_TYPES.has(baseTypeName)) {
		const typeInfo = createPrimitiveTypeInfo(baseTypeName);
		typeInfo.isArray = arrayDepth > 0;
		typeInfo.arrayDepth = arrayDepth;
		typeInfo.isPointer = pointerDepth > 0;
		typeInfo.pointerDepth = pointerDepth;
		return typeInfo;
	}

	const typeInfo = types[baseTypeName];
	if (!typeInfo) {
		return {
			name: baseTypeName,
			package: '',
			kind: 'unknown',
			isArray: arrayDepth > 0,
			arrayDepth: arrayDepth,
			isPointer: pointerDepth > 0,
			pointerDepth: pointerDepth,
			elementType: '',
			fields: []
		};
	}

	return {
		...typeInfo,
		isArray: arrayDepth > 0 || typeInfo.isArray,
		arrayDepth: arrayDepth,
		isPointer: pointerDepth > 0 || typeInfo.isPointer,
		pointerDepth: pointerDepth
	};
}
