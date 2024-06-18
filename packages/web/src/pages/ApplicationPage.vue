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
							label="Config"
							content-class="my-tab-class"
							class="my-active-class"
						/> -->
						<q-tab
							name="files"
							label="Files"
							content-class="my-tab-class"
							class="my-active-class"
						/>
						<q-tab
							name="containers"
							label="Containers"
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
							background: statusStyle[appState]
								? statusStyle[appState].color
								: statusStyle.canceled.color
						}"
					></span>
					<span class="text-capitalize">{{
						appState === 'completed' ? 'Running' : appState
					}}</span>
				</div>
				<q-separator class="q-mx-md" v-if="appState" vertical inset />
				<div
					class="oprate-btn q-mr-sm"
					@mouseenter="handleMouseEnter"
					@mouseleave="handleMouseLeave"
				>
					<div class="oprate-btn-install" v-if="!appState" @click="onInstall">
						Install
					</div>

					<div
						class="oprate-btn-install bg-teal-6 text-white"
						v-if="appState === 'processing' && !showCancel"
					>
						Installing
					</div>

					<div
						class="oprate-btn-install"
						v-if="appState === 'completed' && !showCancel"
						@click="onUpgrade"
					>
						Upgrade
					</div>

					<div
						class="oprate-btn-install bg-teal-6"
						v-if="appState === 'pending' && !showCancel"
					>
						<img class="ani-loading" src="../assets/icon-loading.svg" />
					</div>

					<div
						class="oprate-btn-install"
						v-if="appState !== 'completed' && showCancel"
						@click="onCancel"
					>
						Cancel
					</div>
				</div>

				<div
					class="oprate-btn q-mr-sm"
					@click="onPreview"
					v-if="appState === 'completed'"
				>
					<div class="oprate-btn-install">Preview</div>
				</div>

				<div class="oprate-btn oprate-disabled q-mr-sm" v-else>
					<div class="oprate-btn-install">Preview</div>
				</div>

				<input
					ref="uploadInput"
					type="file"
					style="display: none"
					accept=".tgz"
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
								Upload
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onDownload"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_download" size="20px" />
								Download
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								v-close-popup
								:disable="appState === 'completed' ? false : true"
								@click="onUninstall"
							>
								<q-icon class="q-mr-xs" name="sym_r_reset_tv" size="20px" />
								Uninstall
							</q-item>

							<q-item
								class="row items-center justify-start text-ink-2"
								clickable
								v-ripple
								@click="onDeleteApplication"
								v-close-popup
							>
								<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
								Delete
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
import { ref, onMounted, watch, onUnmounted } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { useDevelopingApps } from '../stores/app';
import { useMenuStore } from '../stores/menu';
import axios from 'axios';
import { ApplicationInfo } from '@devbox/core';
import { BtNotify } from '@bytetrade/ui';
import { statusStyle } from '../types/constants';

import ConfigComponent from '../components/ConfigComponent.vue';
import ContainerComponent from '../components/ContainerComponent.vue';
import EditComponent from '../components/EditComponent.vue';
import DialogConfirm from '../components/dialog/DialogConfirm.vue';

const route = useRoute();
const router = useRouter();
const store = useDevelopingApps();
const menuStore = useMenuStore();
const $q = useQuasar();

const uploadInput = ref();
const appState = ref();
const timer = ref();
const downloading = ref(false);
const showCancel = ref(false);

const appid = ref<string | undefined>(undefined);
const app = ref<ApplicationInfo | undefined>(undefined);

async function refreshApplication() {
	appid.value = route.params.id as string;
	app.value = store.apps.find((app) => app.id == appid.value);
}

onMounted(async () => {
	await refreshApplication();
	await getAppState('init');
});

onUnmounted(() => {
	clearInterval(timer.value);
});

watch(
	() => route.params.id,
	async () => {
		await refreshApplication();
		await __getAppState('init');
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
		$q.notify('Start installing / re-installing');
		getAppState();
	} catch (e: any) {
		console.log(e);
	} finally {
		$q.loading.hide();
	}
}

function onCancel() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Cancel',
			message: 'Are you sure you want to cancel the installation?'
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(store.url + `/api/apps/${app.value.appName}/cancel`, {});
			$q.notify('Cancel successful!');
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
			title: 'Upgrade',
			message: 'Are you sure you want to Upgrade the app?'
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(store.url + '/api/command/install-app', {
				name: app.value.appName
			});
			$q.notify('Start upgraging / re-upgraging');
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
			title: 'Uninstall',
			message: 'Are you sure you want to uninstall the app?'
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await axios.post(
				store.url + `/api/command/uninstall/${app.value.appName}`,
				{}
			);
			$q.notify('Start uninstalling / re-uninstalling');
			getAppState();
		} catch (e: any) {
			console.log(e);
		} finally {
			$q.loading.hide();
		}
	});
}

async function getAppState(isInit?: string) {
	if (timer.value) clearInterval(timer.value);
	await __getAppState(isInit);
	timer.value = setInterval(async () => {
		await __getAppState();
	}, 10000);
}

async function __getAppState(isInit?: string) {
	try {
		const data: any = await store.getAppState(app.value.appName);
		if (!data) {
			return false;
		}

		const status: any = await store.getAppStatus(app.value.appName);
		if (!status) {
			return false;
		}

		appState.value = data.state;
		if (data.state === 'completed') {
			clearInterval(timer.value);
			refreshApplication();
		} else if (
			data.state === 'canceled' ||
			data.state === 'failed' ||
			data.state === 'suspend'
		) {
			if (!isInit) {
				BtNotify.show({
					type: 'bt-failed',
					message: `State ${data.state}`
				});
			}

			appState.value = null;
			clearInterval(timer.value);
		}
	} catch (e: any) {
		appState.value = null;
		clearInterval(timer.value);
	}
}

function handleMouseEnter() {
	if (appState.value === 'pending' || appState.value === 'processing') {
		showCancel.value = true;
	}
}

function handleMouseLeave() {
	showCancel.value = false;
}

async function onDeleteApplication() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Delete',
			message: 'Are you sure to delete the current application?'
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
			$q.notify(message);
			store.cfg = await store.getAppCfg(app.value.appName);
		} else {
			$q.notify(message);
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
		return { status: true, message: 'upload chart success' };
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
		.oprate-btn {
			width: 80px;
			height: 32px;
			line-height: 32px;
			text-algin: center;
			display: inline-block;
			border-radius: 8px;
			border: 1px solid $btn-stroke;
			overflow: hidden;

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
				text-algin: center;
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
				margin-top: 4px;
			}
		}
	}
	.container-left {
	}

	.container-right {
	}
}
</style>
