<template>
	<div class="entrances-item">
		<div class="item-header">
			<div class="entrances-name row items-center justify-center text-grey-10">
				<q-icon name="sym_r_data_table" size="20px" />
				<span class="text-h6 q-ml-md q-mr-lg">
					{{ data.group }}/{{ data.dataType }}
				</span>
			</div>

			<div class="row items-center justify-center text-grey-8">
				<q-icon
					class="cursor-pointer"
					name="sym_r_edit_square"
					size="16px"
					@click="editProvider"
				/>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_delete"
					size="16px"
					@click="deleteProvider"
				/>
			</div>
		</div>

		<div class="item-content">
			<template v-for="op in data.ops" :key="op">
				<span class="q-ml-md text-grey-8 text-caption ops">{{ op }}</span>
			</template>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
import { useQuasar } from 'quasar';

import { useDevelopingApps } from '../../stores/app';

import DialogConfirm from '../dialog/DialogConfirm.vue';
import DialogEditProvider from '../dialog/DialogEditProvider.vue';

const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	}
});

const $q = useQuasar();
const store = useDevelopingApps();

const editProvider = () => {
	$q.dialog({
		component: DialogEditProvider,
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

const deleteProvider = () => {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete Provider Data',
			message: 'Are you sure to delete this provider data?'
		}
	})
		.onOk(() => {
			const providerArr = store.cfg.permission.sysData;
			const index = providerArr.findIndex(
				(item) => item.grop === props.data.grop
			);
			providerArr.splice(index, 1);
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
.entrances-item {
	border: 1px solid rgba(235, 235, 235, 1);
	border-radius: 12px;
	margin-top: 20px;
	.item-header {
		width: 100%;
		height: 56px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid rgba(235, 235, 235, 1);
		padding: 0 24px 0 20px;
		.entrances-name {
			img {
				width: 24px;
				height: 24px;
			}
			.status {
				padding: 4px 12px;
				border-radius: 20px;
				background: var(--Grey-01-, #f6f6f6);
			}
		}
	}

	.item-content {
		padding: 14px 12px;

		.ops {
			padding: 4px 12px;
			border: 1px solid rgba(235, 235, 235, 1);
			border-radius: 20px;
		}
	}
}
</style>
