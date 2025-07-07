<template>
    <aside
        style="width: 250px; min-width: 250px; background: #222; padding: 1rem"
    >
        <h2>Models</h2>
        <ul style="list-style: none; padding: 0">
            <li
                v-for="model in models"
                :key="model"
                @click="$emit('selectModel', model)"
                :style="{
                    cursor: 'pointer',
                    padding: '0.5rem',
                    backgroundColor:
                        model === selectedModel ? '#696969' : 'transparent',
                    borderRadius: '8px',
                    marginBottom: '0.25rem',
                }"
            >
                {{ model }}
            </li>
        </ul>
    </aside>
</template>

<script setup>
defineProps({
    models: Array,
    selectedModel: String,
});
</script>
