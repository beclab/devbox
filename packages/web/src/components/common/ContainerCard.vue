<template>
	<div class="container-item">
		<div class="item-header">
			<div
				class="container-name row items-center justify-start"
				:style="{
					width:
						mode === 'application' ? 'calc(100% - 220px)' : 'calc(100% - 170px)'
				}"
			>
				<div class="q-mr-sm row items-center justify-center container-icon">
					<img src="../../assets/icon-docker.svg" />
					<span
						v-if="container.state"
						:style="{
							background: stateColor[container.state] || stateColor.unknown
						}"
					></span>
				</div>
				<span class="text-ink-1 text-h4 nameText">
					{{
						mode === 'application'
							? container.containerName
							: container.devContainerName
					}}
				</span>
				<!-- {{ container.state }} -->
			</div>
			<div
				class="row items-center justify-end"
				:style="{ width: mode === 'application' ? '220px' : '170px' }"
			>
				<div
					class="bind row items-center justify-center"
					v-if="container.id && mode === 'application'"
					@click="unbindContainer"
				>
					<q-icon name="sym_r_link_off" size="16px" />
					<span class="text-body2 text-ink-2 q-ml-sm">{{
						t('btn_unbind')
					}}</span>
				</div>
				<div
					class="bind row items-center justify-center"
					v-if="!container.id && mode === 'application'"
					@click="bindContainer"
				>
					<q-icon name="sym_r_link" size="16px" />
					<span class="text-body2 text-ink-2 q-ml-sm">{{ t('btn_bind') }}</span>
				</div>
				<div
					class="openIde row items-center justify-center"
					v-if="container.state === 'Running'"
					@click="onCodeBox"
				>
					<img src="../../assets/icon-terminal.svg" />
					<span class="text-body2 text-white q-ml-sm">{{
						t('btn_open_ide')
					}}</span>
				</div>
				<div
					class="openIde disabled-openIde row items-center justify-center"
					v-else
				>
					<img src="../../assets/icon-terminal.svg" />
					<span class="text-body2 text-white q-ml-sm">{{
						t('btn_open_ide')
					}}</span>
				</div>

				<div class="oprate-more q-ml-md" v-if="mode !== 'application'">
					<q-icon name="sym_r_more_vert" />
					<q-menu class="menu-wrap">
						<q-list dense padding>
							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onRenameModel"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_edit_square" size="20px" />
								{{ t('btn_rename') }}
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onUnbind"
								v-close-popup
								:disable="container.appId ? false : true"
							>
								<q-icon class="q-mr-xs" name="sym_r_link_off" size="20px" />

								{{ t('btn_unbind') }}
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onDelete"
								v-close-popup
								:disable="container.appId ? true : false"
							>
								<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
								{{ t('btn_delete') }}
							</q-item>
						</q-list>
					</q-menu>
				</div>
			</div>
		</div>

		<div class="item-content">
			<div class="content-1 row">
				<div class="col-6">
					<div class="row items-center justify-start text-ink-2">
						<q-icon class="q-mr-xs" name="sym_r_code_blocks" size="16px" />
						<span>{{ t('containers_env') }}</span>
					</div>
					<div class="text-ink-1 text-subtitle2" style="padding-left: 20px">
						{{ container.devEnv || '-' }}
					</div>
				</div>

				<div class="col-6">
					<div class="items-center justify-start text-ink-2">
						<q-icon class="q-mr-xs" name="sym_r_deployed_code" size="16px" />
						<span>{{
							mode === 'application'
								? t('containers_bind_dev')
								: t('containers_bind_app')
						}}</span>
					</div>
					<div class="text-ink-1 text-subtitle2" style="padding-left: 20px">
						{{
							mode === 'application'
								? container.devContainerName || '-'
								: container.appName || '-'
						}}
					</div>
				</div>
			</div>

			<div class="content-2 row" style="margin: 20px 0">
				<div class="col-12 items-center justify-start text-ink-2">
					<q-icon class="q-mr-xs" name="sym_r_package_2" size="16px" />
					<span>{{ t('containers_pod_selector') }}</span>
				</div>
				<div
					class="col-12 text-ink-1 text-subtitle2 podSelector"
					style="padding-left: 20px"
				>
					{{ container.podSelector || '-' }}
				</div>
			</div>

			<div class="content-3 row">
				<div class="col-12 row items-center justify-start text-ink-2">
					<q-icon class="q-mr-xs" name="sym_r_schedule" size="16px" />
					<span>{{ t('containers_update_time') }}</span>
				</div>
				<div
					class="col-12 text-ink-1 text-subtitle2"
					style="padding-left: 20px"
				>
					{{ container.updateTime || '-' }}
				</div>
			</div>
		</div>

		<div
			class="item-footer row items-center justify-between"
			:class="{ 'footer-bind': container.appId }"
			v-if="mode !== 'application'"
		>
			<div class="row items-center justify-center">
				<img v-if="container.appId" :src="container.icon" />
				<span v-else class="img-empty"></span>
				<span v-if="container.appId">{{ container.appName }}</span>
				<span v-else class="text-ink-2">{{ t('application') }}</span>
			</div>
			<div class="text-teal-6 text-subtitle3" v-if="container.appId">
				{{ t('btn_binding') }}
			</div>
			<div class="text-red-6 text-subtitle3" v-else>{{ t('btn_unbind') }}</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, ref, onMounted } from 'vue';
import axios from 'axios';
import { useQuasar } from 'quasar';
import { ApplicationInfo } from '@devbox/core';
import { BtDialog } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';

import { useDevelopingApps } from '../../stores/app';

const props = defineProps({
	container: {
		type: Object,
		required: false,
		default: () => ({})
	},
	mode: {
		type: String,
		required: false,
		default: 'devbox'
	}
});

const emits = defineEmits(['bindContainer', 'unbindContainer']);

const { t } = useI18n();
const $q = useQuasar();
const store = useDevelopingApps();
const app = ref<ApplicationInfo | undefined>(undefined);
const stateColor = ref({
	Running: 'rgba(41, 204, 95, 1)',
	Waiting: 'rgba(254, 190, 1, 1)',
	Terminated: 'rgba(255, 77, 77, 1)',
	unknown: 'rgba(173, 173, 173, 1)'
});

const bindContainer = () => {
	emits('bindContainer', props.container);
};

const unbindContainer = () => {
	emits('unbindContainer', props.container);
};

const onCodeBox = async () => {
	if (window.top == window) {
		window.open('//' + app.value.entrance + props.container.devPath, '_blank');
	} else {
		await store.openApplication({
			appid: app.value.appID,
			path: props.container.devPath
		});
	}
};

const onRenameModel = () => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: 'Rename',
		prompt: {
			model: props.container.devContainerName,
			isValid: (val) => val.length > 2,
			type: 'text',
			name: 'New Name',
			placeholder: ''
		}
	})
		.then((val) => {
			if (!val) return false;
			onRename(val);
		})
		.catch((err) => {
			console.log(err);
		});
};

const onRename = async (val) => {
	const res: any = await axios.patch(
		store.url + '/api/dev-container/' + props.container.devContainerName,
		{
			devContainerName: val
		},
		{
			headers: { 'content-type': 'application/json' }
		}
	);
	$q.notify('success to rename');
	store.getMyContainers();
};

const onUnbind = () => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: 'Unbind',
		message: 'Are you sure you want to unbind this Container?'
	})
		.then(async (val) => {
			if (!val) return false;
			const params = {
				appId: props.container.appId,
				podSelector: props.container.podSelector,
				containerName: props.container.containerName,
				containerId: props.container.id
			};

			await store.unbindContainer(params);
			$q.notify('unbind to rename');
			await store.getMyContainers();
		})
		.catch((err) => {
			console.log(err);
		});
};

const onDelete = () => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: 'Delete',
		message: 'Are you sure you want to delete this Container?'
	})
		.then(async (val) => {
			if (!val) return false;
			await axios.delete(
				store.url + '/api/dev-container/' + props.container.devContainerName,
				{
					headers: { 'content-type': 'application/json' }
				}
			);
			$q.notify('delete to rename');
			store.getMyContainers();
		})
		.catch((err) => {
			console.log(err);
		});
};

async function refreshApplication() {
	app.value = store.apps.find((app) => app.id == props.container.appId);
}

onMounted(() => {
	refreshApplication();
});
</script>

<style lang="scss" scoped>
.container-item {
	border: 1px solid $separator;
	border-radius: 12px;
	padding: 0 32px 32px 32px;
	.item-header {
		width: 100%;
		height: 96px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid $separator;
		.container-name {
			.nameText {
				display: inline-block;
				width: calc(100% - 62px);
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
			}
			.container-icon {
				width: 32px;
				height: 32px;
				border-radius: 16px;
				background: rgba(232, 244, 255, 1);
				position: relative;
				span {
					width: 12px;
					height: 12px;
					border-radius: 6px;
					box-sizing: border-box;
					border: 3px solid #ffffff;
					position: absolute;
					right: 0;
					bottom: 0;
					background: rgba(41, 204, 95, 1);
				}
			}
		}
		.openIde {
			padding: 6px 12px;
			background: rgba(0, 190, 158, 1);
			border-radius: 8px;
			margin-left: 12px;
			cursor: pointer;
			&.disabled-openIde {
				background: rgba(0, 190, 158, 0.5);
			}
		}
		.bind {
			padding: 6px 12px;
			border-radius: 8px;
			border: 1px solid $separator;
			cursor: pointer;
		}
	}

	.item-content {
		padding: 32px 0 0 0;

		.podSelector {
			width: 100%;
			text-overflow: ellipsis;
			white-space: nowrap;
			overflow: hidden;
		}
	}

	.item-footer {
		width: 100%;
		height: 72px;
		padding: 0 20px;
		margin-top: 32px;
		border-radius: 12px;
		border: 1px dashed $separator;
		img {
			width: 32px;
			height: 32px;
			border-radius: 8px;
			margin-right: 12px;
		}
		.img-empty {
			display: inline-block;
			width: 32px;
			height: 32px;
			border-radius: 8px;
			border: 1px dashed #dbdbdb;
			box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
			margin-right: 12px;
		}

		&.footer-bind {
			background: linear-gradient(
				274deg,
				#dafbf6 29.63%,
				rgba(214, 249, 243, 0.8) 49.97%,
				rgba(210, 247, 241, 0.6) 73.52%
			);
			border: 1px solid $separator;
		}
	}
}

.oprate-more {
	width: 32px;
	height: 32px;
	border-radius: 8px;
	display: flex;
	align-items: center;
	justify-content: center;
	box-sizing: border-box;
	border: 1px solid $separator;
	cursor: pointer;
	&:hover {
		background: rgba(0, 0, 0, 0.05);
	}
}
</style>
