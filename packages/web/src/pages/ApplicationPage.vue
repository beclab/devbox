<template>
	<div class="container" v-if="app">
		<div class="container-header row items-center justify-between">
			<div class="row items-center justify-between">
				<div class="row items-center justify-start">
					<div class="app row items-center justify-center">
						<img
							v-if="store.cfg && store.cfg.metadata"
							:src="store.cfg.metadata.icon"
						/>
						<span class="q-ml-sm text-h6 text-ink-1">{{ app.appName }}</span>
					</div>
					<q-tabs
						v-model="menuStore.appCurrentItem"
						dense
						no-caps
						class="text-ink-2"
						indicator-color="teal-pressed"
						active-color="teal-pressed"
						align="justify"
						narrow-indicator
					>
						<!-- <q-tab
							name="config"
							:label="t('app_config')"
							content-class="my-tab-class"
							class="my-active-class"
						/> -->
						<q-tab
							name="files"
							:label="t('app_files')"
							content-class="my-tab-class"
							class="my-active-class"
						/>
						<q-tab
							name="containers"
							:label="t('app_containers')"
							content-class="my-tab-class"
							class="my-active-class"
						/>
					</q-tabs>
				</div>
			</div>
			<div class="row items-center justify-end">
				<div class="status" v-if="appState">
					<span
						class="color"
						:style="{
							background:
								appState.toLowerCase() === 'running'
									? statusStyle.running.color
									: statusStyle.pending.color
						}"
					></span>
					<span class="text-capitalize q-mr-xs">
						{{
							appState === 'uninstalling' && appOperateType === 'install'
								? 'installing'
								: appState
						}}</span
					>
				</div>
				<q-separator class="q-mx-md" v-if="appState" vertical inset />

				<div
					class="event-btn q-mr-sm"
					v-if="
						appOperateState === '' ||
						(appOperateType === 'uninstall' && appOperateState === 'completed')
					"
					@click="onInstall"
				>
					{{ t('btn_install') }}
				</div>

				<div
					v-if="
						['downloading', 'processing'].includes(appOperateState) &&
						appState !== 'running' &&
						appOperateType === 'install'
					"
					class="event-btn bg-teal-6 text-white q-mr-sm"
					@mouseenter="handleMouseEnter"
					@mouseleave="handleMouseLeave"
					@click="onCancel"
				>
					{{ showCancel ? t('btn_cancel') : t('btn_installing') }}
				</div>

				<div
					v-if="appState === 'running'"
					class="event-btn q-mr-sm"
					@click="onUpgrade"
				>
					{{ t('btn_upgrade') }}
				</div>

				<div
					v-if="appOperateType === 'uninstall' && appState === 'uninstalling'"
					class="event-btn q-mr-sm"
				>
					Uninstalling
				</div>

				<div
					v-if="appOperateState === 'pending'"
					class="event-btn bg-teal-6 q-mr-sm"
				>
					<img class="ani-loading" src="../assets/icon-loading.svg" />
				</div>

				<div
					class="oprate-btn q-mr-sm"
					@click="onPreview"
					v-if="appState === 'running'"
				>
					<div class="oprate-btn-install">{{ t('btn_preview') }}</div>
				</div>

				<div class="oprate-btn oprate-disabled q-mr-sm" v-else>
					<div class="oprate-btn-install">{{ t('btn_preview') }}</div>
				</div>

				<input
					ref="uploadInput"
					type="file"
					style="display: none"
					accept=".zip,.rar,.7z,.tar,.tgz"
					@change="uploadFile"
				/>

				<div class="oprate-more">
					<q-icon name="sym_r_more_vert" color="ink-1" />
					<q-menu class="rounded-borders" flat>
						<q-list dense padding>
							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onUploadChart"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_upload" size="20px" />
								{{ t('btn_upload') }}
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onDownload"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_download" size="20px" />
								{{ t('btn_download') }}
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								v-close-popup
								:disable="appState === 'running' ? false : true"
								@click="onUninstall"
							>
								<q-icon class="q-mr-xs" name="sym_r_reset_tv" size="20px" />
								{{ t('btn_uninstall') }}
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onDeleteApplication"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
								{{ t('btn_delete') }}
							</q-item>
						</q-list>
					</q-menu>
				</div>
			</div>
		</div>

		<div class="container-left">
			<q-tab-panels v-model="menuStore.appCurrentItem" animated>
				<!-- <q-tab-panel name="config">
					<ConfigComponent :app="app" :downloading="downloading" />
				</q-tab-panel> -->

				<q-tab-panel name="files" class="q-pa-none">
					<EditComponent :app="app" />
				</q-tab-panel>

				<q-tab-panel name="containers" class="q-px-none">
					<ContainerComponent :app="app" />
				</q-tab-panel>
			</q-tab-panels>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onUnmounted } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { useDevelopingApps } from '../stores/app';
import { useMenuStore } from '../stores/menu';
import axios from 'axios';
import { ApplicationInfo } from '@devbox/core';
import { statusStyle } from '../types/constants';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

// import ConfigComponent from '../components/ConfigComponent.vue';
import ContainerComponent from '../components/ContainerComponent.vue';
import EditComponent from '../components/EditComponent.vue';
import DialogConfirm from '../components/dialog/DialogConfirm.vue';
import { log } from 'console';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const store = useDevelopingApps();
const menuStore = useMenuStore();
const $q = useQuasar();

const uploadInput = ref();
const appState = ref();
const appOperateState = ref();
const appOperateType = ref();
const timer = ref();
const downloading = ref(false);
const showCancel = ref(false);

const appid = ref<string | undefined>(undefined);
const app = ref<ApplicationInfo | undefined>(undefined);

async function refreshApplication() {
	appid.value = route.params.id as string;
	app.value = store.apps.find((app) => app.id == appid.value);
}

onUnmounted(() => {
	clearInterval(timer.value);
});

watch(
	() => route.params.id,
	async (newVal) => {
		if (newVal) {
			await refreshApplication();
			await getAppState();
		}
	},
	{
		immediate: true
	}
);

async function onPreview() {
	if (window.top == window) {
		window.open('//' + app.value.entrance, '_blank');
	} else {
		await store.openApplication({
			appid: app.value.appID,
			path: ''
		});
	}
}

async function onInstall() {
	$q.loading.show();
	try {
		await axios.post(store.url + '/api/command/install-app', {
			name: app.value.appName
		});
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.start_installing')
		});
		getAppState();
	} catch (e: any) {
		console.log(e);
	} finally {
		$q.loading.hide();
	}
}

function onCancel() {
	if (showCancel.value) {
		return false;
	}
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('cancel'),
			message: t('cancel_installation')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(store.url + `/api/apps/${app.value.appName}/cancel`, {});
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.successfully')
			});
			getAppState();
		} catch (e: any) {
			console.log(e);
		} finally {
			$q.loading.hide();
		}
	});
}

function onUpgrade() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_upgrade'),
			message: t('upgrade_app')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(store.url + '/api/command/install-app', {
				name: app.value.appName
			});
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.start_upgrade')
			});

			getAppState();
		} catch (e: any) {
			console.log(e);
		} finally {
			$q.loading.hide();
		}
	});
}

function onUninstall() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_uninstall'),
			message: t('message.uninstall_app')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(
				store.url + `/api/command/uninstall/${app.value.appName}`,
				{}
			);
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.start_uninstalling')
			});
			getAppState();
		} catch (e: any) {
			console.log(e);
		} finally {
			$q.loading.hide();
		}
	});
}

async function getAppState() {
	if (timer.value) clearInterval(timer.value);
	await __getAppState();
	timer.value = setInterval(async () => {
		await __getAppState();
	}, 2000);
}

async function __getAppState() {
	try {
		let appOperateState_temp: string;
		let appState_temp: string;
		let appOperateType_temp: string;

		// operate Status
		const data: any = await store.getAppState(app.value.appName);
		console.log('appOperateState_temp', data);
		if (!data) {
			return false;
		}
		appOperateState_temp = data.state;
		appOperateType_temp = data.opType;

		if (
			appOperateState_temp === '' ||
			appOperateState_temp === 'canceled' ||
			appOperateState_temp === 'failed' ||
			appOperateState_temp === 'completed'
		) {
			clearInterval(timer.value);
		}

		// App Status
		const res: any = await store.getAppStatus(app.value.appName);
		console.log('appState_temp', res);
		if (!res) {
			return false;
		}
		appState_temp = res.status.state;

		if (
			appOperateState_temp === 'failed' &&
			appOperateType_temp === 'install'
		) {
			appState_temp = '';
		}

		if (
			appOperateState_temp === 'failed' &&
			appOperateType_temp === 'uninstall'
		) {
			appState_temp = 'running';
		}

		appOperateType.value = appOperateType_temp;
		appOperateState.value = appOperateState_temp || '';
		appState.value = appState_temp || '';

		if (appState_temp === 'running') {
			clearInterval(timer.value);
			refreshApplication();
		}

		if (appState_temp === 'suspend') {
			clearInterval(timer.value);
		}
	} catch (e: any) {
		// appState.value = null;
		clearInterval(timer.value);
	}
}

function handleMouseEnter() {
	showCancel.value = true;
}

function handleMouseLeave() {
	showCancel.value = false;
}

async function onDeleteApplication() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_delete'),
			message: t('message.delete_app')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(store.url + '/api/command/delete-app', {
				name: app.value.appName
			});
			await store.getApps();
			await store.getMyContainers();
			await menuStore.updateApplications();

			router.push({ path: '/home' });
		} catch (e: any) {
			console.log(e);
		} finally {
			$q.loading.hide();
		}
	});
}

async function onDownload() {
	downloading.value = true;
	window.location.href =
		store.url + '/api/command/download-app-chart?app=' + app.value.appName;
	downloading.value = false;
}

async function onUploadChart() {
	//  $q.notify('not implement');
	uploadInput.value.value = null;
	uploadInput.value.click();
}

async function uploadFile(event: any) {
	const file = event.target.files[0];
	if (file) {
		const { status, message } = await upload_dev_file(file);
		if (status) {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: message
			});
			store.cfg = await store.getAppCfg(app.value.appName);
		} else {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: message
			});
		}
	} else {
		console.log('file selected failure');
	}
}

async function upload_dev_file(
	file: any
): Promise<{ status: boolean; message: string }> {
	try {
		const formData = new FormData();
		formData.append('chart', file);
		formData.append('app', app.value.appName);
		await axios.post(store.url + '/api/command/upload-app-chart', formData, {
			headers: { 'Content-Type': 'multipart/form-data' }
		});
		return { status: true, message: t('message.successfully') };
	} catch (e: any) {
		console.log(e);
		return { status: false, message: e.message };
	}
}
</script>
<style lang="scss">
.rounded-borders {
	border-radius: 12px !important;
	overflow: hidden;
}

.my-active-class {
	height: 52px !important;
}
</style>
<style scoped lang="scss">
// .content {
//   height: calc(100vh - 100px);
// }

.container {
	padding: 12px 44px;
	height: calc(100vh - 88px);
	.container-header {
		width: 100%;
		.app {
			margin-right: 20px;

			img {
				width: 32px;
				height: 32px;
				border-radius: 8px;
			}
		}

		.event-btn {
			width: 80px;
			height: 32px;
			line-height: 32px;
			text-align: center;
			display: inline-block;
			border-radius: 8px;
			border: 1px solid $btn-stroke;
			overflow: hidden;
			color: $ink-1;
			cursor: pointer;

			&:hover {
				background: $background-hover;
			}
			&.oprate-disabled {
				opacity: 0.5;
			}
		}

		.oprate-btn {
			width: 80px;
			height: 32px;
			line-height: 32px;
			text-align: center;
			display: inline-block;
			border-radius: 8px;
			border: 1px solid $btn-stroke;
			overflow: hidden;
			color: $ink-1;

			span {
				margin-left: 4px;
			}
			&:hover {
				background: $background-hover;
			}
			&.oprate-disabled {
				opacity: 0.5;
			}

			.oprate-btn-install {
				width: 100%;
				height: 100%;
				line-height: 32px;
				padding: 0 8px;
				color: $ink-1;
				font-size: 12px;
				line-height: 100%;
				text-align: center;
				cursor: pointer;
				display: flex;
				align-items: center;
				justify-content: center;

				.ani-loading {
					animation: rotate 1s linear infinite;
				}
			}
		}

		@keyframes rotate {
			from {
				transform: rotate(360deg);
			}
			to {
				transform: rotate(0deg);
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
			border: 1px solid $btn-stroke;
			cursor: pointer;
			&:hover {
				background: $background-hover;
			}
		}
		.status {
			display: flex;
			align-items: center;
			justify-content: center;
			.color {
				width: 12px;
				height: 12px;
				border-radius: 6px;
				display: inline-block;
				margin-right: 8px;
			}
		}
	}
}
</style>
