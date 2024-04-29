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
	color: rgba(0, 150, 125, 1);
	margin-top: 12px;
	a {
		color: rgba(0, 150, 125, 1);
		text-decoration: none;
	}
}
</style>
