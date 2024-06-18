<template>
	<div class="resources-item row items-center justify-between">
		<div class="resources-left row items-center justify-start">
			<q-icon name="sym_r_work_alert" color="text-ink-1" size="16px" />
			<span class="text-ink-1 text-h6 q-ml-sm">{{ data.name }}</span>
			<span class="text-blue-6 text-subtitle2 q-mx-sm">{{ data.version }}</span>
			<span class="text-subtitle3 status">{{ data.type }}</span>
		</div>
		<div class="resources-right row items-center justify-end">
			<q-icon
				class="cursor-pointer"
				name="sym_r_edit_square"
				size="16px"
				color="ink-2"
				@click="editDependencie"
			/>
			<q-icon
				class="q-ml-md cursor-pointer"
				name="sym_r_delete"
				size="16px"
				color="ink-2"
				@click="deleteDependencie"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';
import DialogDependency from '../dialog/DialogDependency.vue';
import DialogConfirm from '../dialog/DialogConfirm.vue';

const $q = useQuasar();
const store = useDevelopingApps();
const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	},
	mode: {
		type: String,
		required: false,
		default: 'create'
	}
});

const editDependencie = () => {
	$q.dialog({
		component: DialogDependency,
		componentProps: {
			data: props.data,
			mode: 'edit'
		}
	})
		.onOk(() => {
			console.log('OK');
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const deleteDependencie = () => {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete Dependencie',
			message: 'Are you sure to delete this dependencie?'
		}
	})
		.onOk(() => {
			const dependenciesArr = store.cfg.options.analytics.dependencies;
			const index = dependenciesArr.findIndex(
				(item) => item.name === props.data.name
			);
			dependenciesArr.splice(index, 1);
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
		.status {
			color: rgba(51, 119, 255, 1);
			padding: 4px 12px;
			border-radius: 20px;
			background: $background-hover;
		}
	}
}
</style>
