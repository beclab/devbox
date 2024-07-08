<template>
	<div class="entrances-item">
		<div class="item-header">
			<div class="entrances-name row items-center justify-center">
				<span class="text-ink-1 text-h6 q-ml-md q-mr-lg capitalize">{{
					name
				}}</span>
				<span class="text-ink-2 text-subtitle3 q-mr-md">{{
					data.username
				}}</span>
				<span
					class="text-subtitle3 password"
					v-if="!passwordVisibility && data.password"
				>
					<i v-for="(c, i) in data.password.length" :key="i" />
				</span>
				<span class="text-ink-2 text-subtitle3 password" v-else>
					{{ data.password }}
				</span>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_visibility"
					size="16px"
					v-if="!passwordVisibility"
					@click="passwordVisibility = !passwordVisibility"
				/>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_visibility_off"
					size="16px"
					v-else
					@click="passwordVisibility = !passwordVisibility"
				/>
			</div>

			<div class="row items-center justify-center">
				<q-icon
					class="cursor-pointer"
					name="sym_r_edit_square"
					color="ink-2"
					size="16px"
					@click="editMiddle"
				/>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_delete"
					size="16px"
					@click="deleteMiddle"
				/>
			</div>
		</div>

		<div class="item-content">
			<div class="content-hearder row items-center justify-between">
				<span class="q-ml-md">{{
					t('config_space_middleware_Databases')
				}}</span>
				<q-icon
					class="cursor-pointer q-mr-md"
					name="sym_r_add_box"
					color="ink-2"
					size="16px"
					@click="addPolicie"
				/>
			</div>
			<template v-if="data.databases && data.databases.length > 0">
				<div
					class="content row items-center justify-between"
					v-for="(database, index) in data.databases"
					:key="index"
				>
					<div class="row items-center justify-start">
						<div class="file-icon q-mr-md">
							<q-icon class="cursor-pointer" name="sym_r_article" size="16px" />
						</div>
						<div>
							<div>
								<span class="text-ink-1 text-subtitle3 q-mr-md">{{
									database.name
								}}</span>
								<span
									class="text-ink-2 text-caption q-px-md q-py-xs label"
									v-if="database.distributed"
									>distributed</span
								>
							</div>
						</div>
					</div>

					<div class="row items-center justify-center">
						<q-icon
							class="cursor-pointer"
							name="sym_r_edit_square"
							color="ink-2"
							size="16px"
							@click="editPolicie(database)"
						/>
						<q-icon
							class="q-ml-md cursor-pointer"
							name="sym_r_delete"
							color="ink-2"
							size="16px"
							@click="deletePolicie(database)"
						/>
					</div>
				</div>
			</template>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useDevelopingApps } from '../../stores/app';

import DialogMiddleware from '../dialog/DialogMiddleware.vue';
import DialogConfirm from '../dialog/DialogConfirm.vue';
import DialogMiddlePolicies from '../dialog/DialogMiddlePolicies.vue';

const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	},
	name: {
		type: String,
		required: false,
		default: 'postgres'
	}
});

const $q = useQuasar();
const { t } = useI18n();
const store = useDevelopingApps();
const passwordVisibility = ref(false);

const editMiddle = () => {
	$q.dialog({
		component: DialogMiddleware,
		componentProps: {
			data: { ...props.data, name: props.name },
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

const deleteMiddle = () => {
	const title = `Delete ${props.name}`;
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: title,
			message: `Are you sure to delete ${props.name}?`
		}
	})
		.onOk(() => {
			console.log('OK');
			delete store.cfg.middleware[props.name];
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const initPolicie = () => {
	return {
		name: '',
		distributed: true
	};
};

const addPolicie = () => {
	$q.dialog({
		component: DialogMiddlePolicies,
		componentProps: {
			data: initPolicie(),
			name: props.name,
			mode: 'create'
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

const editPolicie = (database) => {
	$q.dialog({
		component: DialogMiddlePolicies,
		componentProps: {
			data: database,
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

const deletePolicie = (database) => {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete Policie',
			message: 'Are you sure to delete this policie?'
		}
	})
		.onOk(() => {
			const index = store.cfg.middleware[props.name].databases.findIndex(
				(item) => item.name === database.name
			);
			store.cfg.middleware[props.name].databases.splice(index, 1);
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
			.password {
				display: flex;
				align-items: center;
				justify-content: flex-start;
				i {
					display: inline-block;
					width: 8px;
					height: 8px;
					border-radius: 4px;
					background: rgba(92, 85, 81, 1);
					margin-right: 4px;
				}
			}
		}
	}

	.item-content {
		padding: 14px 12px;

		.content {
			background: rgba(246, 246, 246, 1);
			border-radius: 12px;
			margin: 12px 0;
			padding: 12px;
			.file-icon {
				width: 32px;
				height: 32px;
				display: flex;
				align-items: center;
				justify-content: center;
				background: #ffffff;
				border-radius: 16px;
			}
			.label {
				border: 1px solid rgba(219, 219, 219, 1);
				border-radius: 20px;
			}
		}
	}
}
</style>
