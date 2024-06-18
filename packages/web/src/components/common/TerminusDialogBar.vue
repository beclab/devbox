<template>
	<q-bar class="bar q-mb-md">
		<q-icon v-if="icon" :name="icon" size="18px" />
		<div class="title col" :class="titAlign ? titAlign : 'text-left'">
			{{ label }}
		</div>
		<q-space />
		<q-btn dense flat icon="close" @click="onCancel" color="ink-3">
			<q-tooltip>Close</q-tooltip>
		</q-btn>
	</q-bar>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	titAlign: {
		type: String,
		default: 'left',
		required: false
	}
});

const emit = defineEmits(['close']);

const onCancel = () => {
	emit('close');
};
</script>

<style scoped lang="scss">
.mobile-title {
	color: #322f2d;
	font-size: 16px;
	text-align: center;
	font-weight: 500;
	margin: 20px 0;
}
.bar {
	background-color: $background-3;
	.title {
		height: 18px;
		line-height: 18px;
		color: $ink-1;
		font-size: 14px;
	}
}
</style>
