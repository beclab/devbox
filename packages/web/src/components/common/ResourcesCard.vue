<template>
	<div class="resources-item row items-center justify-between">
		<div class="resources-left row items-center justify-start">
			<img src="../../assets/icon-cpu.svg" />
			<span class="text-ink-2 text-h6"
				>{{ data.label }} {{ data.require ? ' *' : '' }}</span
			>
		</div>
		<div class="resources-right row items-center justify-end col-6">
			<span
				class="text-body2 text-ink-2 q-mr-lg resource-text"
				v-if="data.required || data.limited"
			>
				Requests {{ data.required }}{{ data.requiredUnit }}, Limits{{
					data.limited
				}}{{ data.limitUnit }}
			</span>
			<span class="text-body2 text-ink-2 q-mr-lg resource-text" v-else>
				{{ data.label }} requirement not specified
			</span>
			<q-icon
				class="cursor-pointer"
				name="sym_r_edit_square"
				size="16px"
				@click="editResource"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, defineEmits } from 'vue';
import { useQuasar } from 'quasar';

import DialogEditResource from '../dialog/DialogEditResource.vue';

const $q = useQuasar();
const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	}
});

const emit = defineEmits(['updateResources']);

const editResource = () => {
	$q.dialog({
		component: DialogEditResource,
		componentProps: {
			data: props.data
		}
	})
		.onOk(() => {
			console.log('OK');
			emit('updateResources');
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};
</script>

<style lang="scss" scoped>
.resources-item {
	border: 1px solid $separator;
	border-radius: 8px;
	margin-top: 12px;
	padding: 14px 20px;
	.resources-left {
		img {
			width: 24px;
			height: 24px;
			margin-right: 12px;
		}
	}

	.resource-text {
		width: calc(100% - 36px);
		display: inline-block;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		text-align: right;
	}
}
</style>
