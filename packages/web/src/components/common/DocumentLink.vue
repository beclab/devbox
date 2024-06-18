<template>
	<div class="row items-center justify-start item">
		<q-icon name="sym_r_feed" size="16px" />
		<a class="q-ml-sm text-subtitle1" :href="data.link">{{ data.message }}</a>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	}
});
</script>

<style lang="scss" scoped>
.item {
	color: $teal-pressed;
	margin-top: 12px;
	a {
		color: $teal-pressed;
		text-decoration: none;
	}
}
</style>
