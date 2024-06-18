<template>
	<div class="entrances-item">
		<div class="item-header">
			<div class="entrances-name row items-center justify-center">
				<img src="../../assets/icon-docker.svg" />
				<span class="text-ink-1 text-h6 q-ml-md q-mr-lg">{{ data.name }}</span>
				<span class="text-green text-subtitle3 status">{{
					data.authLevel
				}}</span>
			</div>

			<div class="row items-center justify-center">
				<q-icon
					class="cursor-pointer"
					name="sym_r_edit_square"
					size="16px"
					color="ink-2"
					@click="editEntrance"
				/>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_delete"
					size="16px"
					color="ink-2"
					@click="deleteEntrance"
					v-if="!disabledRemove"
				/>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_delete"
					size="16px"
					color="ink-2"
					disabled
					v-else
				/>
			</div>
		</div>

		<div class="item-content">
			<div class="content-hearder row items-center justify-between">
				<span class="q-ml-md text-ink-2">Policies</span>
				<q-icon
					class="cursor-pointer q-mr-md"
					name="sym_r_add_box"
					size="16px"
					color="ink-2"
					@click="addPolicies"
				/>
			</div>
			<template
				v-for="(policy, index) in store.cfg.options.policies"
				:key="index"
			>
				<div
					class="content row items-center justify-between"
					v-if="
						store.cfg.options.policies &&
						store.cfg.options.policies.length > 0 &&
						policy.entranceName === props.data.name
					"
				>
					<div class="row items-center justify-start">
						<div class="file-icon q-mr-md">
							<q-icon class="cursor-pointer" name="sym_r_article" size="16px" />
						</div>
						<div>
							<div>
								<span class="text-ink-1 text-subtitle3 q-mr-md">{{
									policy.uriRegex
								}}</span>
								<span class="text-ink-2 text-caption q-px-md q-py-xs label">{{
									policy.level
								}}</span>
							</div>
							<div class="text-overline text-ink-2 q-mt-xs">
								{{ policy.description }}
							</div>
						</div>
					</div>

					<div class="row items-center justify-center">
						<q-icon
							class="cursor-pointer"
							name="sym_r_edit_square"
							color="ink-2"
							size="16px"
							@click="editPolicie(policy)"
						/>
						<q-icon
							class="q-ml-md cursor-pointer"
							name="sym_r_delete"
							color="ink-2"
							size="16px"
							@click="deletePolicie(policy)"
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

import { useDevelopingApps } from '../../stores/app';

import DialogEntrance from '../dialog/DialogEntrance.vue';
import DialogPolicies from '../dialog/DialogPolicies.vue';
import DialogConfirm from '../dialog/DialogConfirm.vue';

const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	},
	disabledRemove: {
		type: Boolean,
		required: false,
		default: false
	}
});

const $q = useQuasar();
const store = useDevelopingApps();
const defaultPolicies = ref({
	uriRegex: '',
	level: 'two_factor',
	oneTime: false,
	validDuration: '',
	entranceName: props.data.name,
	description: ''
});

const editEntrance = () => {
	$q.dialog({
		component: DialogEntrance,
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

const deleteEntrance = () => {
	if (props.disabledRemove) {
		return false;
	}
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete Entrance',
			message: 'Are you sure you want to delete this item?'
		}
	})
		.onOk(() => {
			const entrancesArr = store.cfg.entrances;
			const index = entrancesArr.findIndex(
				(item) => item.name === props.data.name
			);
			entrancesArr.splice(index, 1);
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const addPolicies = () => {
	defaultPolicies.value.entranceName = props.data.name;
	$q.dialog({
		component: DialogPolicies,
		componentProps: {
			data: defaultPolicies.value,
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

const editPolicie = (policy) => {
	$q.dialog({
		component: DialogPolicies,
		componentProps: {
			data: policy,
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

const deletePolicie = (policy) => {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete Policy',
			message: 'Are you sure to delete this Policy?'
		}
	})
		.onOk(() => {
			const policiesArr = store.cfg.options.policies;
			const index = policiesArr.findIndex(
				(item) => item.entranceName === policy.entranceName
			);
			policiesArr.splice(index, 1);
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
	border: 1px solid $separator;
	border-radius: 12px;
	margin-top: 20px;
	.item-header {
		width: 100%;
		height: 56px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid $separator;
		padding: 0 24px 0 20px;
		.entrances-name {
			img {
				width: 24px;
				height: 24px;
			}
			.status {
				padding: 4px 12px;
				border-radius: 20px;
				background: $background-hover;
			}
		}
	}

	.item-content {
		padding: 14px 12px;

		.content {
			background: rgba(246, 246, 246, 1);
			border-radius: 12px;
			margin-top: 12px;
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
