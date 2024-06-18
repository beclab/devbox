<template>
	<div class="resources-item row items-center justify-between">
		<div class="resources-left row items-center justify-start">
			<q-icon name="sym_r_assignment" color="textink-1" size="16px" />
			<span class="textink-1 text-h6 q-ml-sm">{{ name }}</span>
		</div>
		<div class="resources-right row items-center justify-end">
			<q-icon
				class="cursor-pointer"
				name="sym_r_edit_square"
				color="ink-2"
				size="16px"
				@click="editDependencie"
			/>
			<q-icon
				class="q-ml-md cursor-pointer"
				name="sym_r_delete"
				color="ink-2"
				size="16px"
				@click="deleteDependencie"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';
import DialogConfirm from '../dialog/DialogConfirm.vue';
import DialogEditReference from '../dialog/DialogEditReference.vue';

const $q = useQuasar();
const store = useDevelopingApps();
const props = defineProps({
	name: {
		type: String,
		required: false,
		default: '-'
	}
});

const editDependencie = () => {
	$q.dialog({
		component: DialogEditReference,
		componentProps: {
			name: props.name,
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
			title: 'Delete Client Reference',
			message: 'Are you sure to delete this client reference?'
		}
	})
		.onOk(() => {
			const dependenciesArr = store.cfg.options.appScope.appRef;
			const index = dependenciesArr.findIndex((item) => item === props.name);
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
	border: 1px solid rgba(235, 235, 235, 1);
	border-radius: 8px;
	margin-top: 12px;
	padding: 14px 20px;
	.resources-left {
		.status {
			color: rgba(51, 119, 255, 1);
			padding: 4px 12px;
			border-radius: 20px;
			background: var(--Grey-01-, #f6f6f6);
		}
	}
}
</style>
