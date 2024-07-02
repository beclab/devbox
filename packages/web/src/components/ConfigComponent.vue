<template>
	<div class="config-right" v-if="store.cfg">
		<div class="config-right-header row items-center justify-between">
			<div class="row items-center justify-start">
				<img class="q-mr-sm" src="../assets/icon-txt.svg" style="width: 12px" />
				<span class="text-ink-1">{{ t('config_name') }}</span>
				<span
					class="statusIcon q-ml-sm"
					:style="{
						background: configStatus ? '#FFC46D' : 'rgba(41, 204, 95, 1)'
					}"
				></span>
			</div>
			<div>
				<q-icon
					class="q-ml-md cursor-pointer"
					name="sym_r_save"
					size="20px"
					@click="onSave"
				/>
			</div>
		</div>
		<div class="config-right-content row">
			<div class="config-menu">
				<q-tabs
					v-model="tab"
					dense
					vertical
					class="text-ink-2"
					active-class="my-menu-class"
					active-color="teal-8"
					indicator-color="transparent"
					active-bg-color="background-selected"
					no-caps
				>
					<q-tab
						:name="menu.label"
						class="menuTab"
						v-for="menu in store.configMenu"
						:key="menu.label"
					>
						<div class="row items-center justify-center">
							<img
								src="../assets/icon-default-active.svg"
								v-if="menu.status === 0 && menu.label === tab"
							/>
							<img
								src="../assets/icon-default.svg"
								v-if="menu.status === 0 && menu.label !== tab"
							/>
							<q-icon
								name="sym_r_check_circle"
								v-if="menu.status === 1"
								color="green"
								size="16px"
							/>
							<q-icon
								name="sym_r_error"
								v-if="menu.status === 2"
								color="negative"
								size="16px"
							/>
							<span class="q-ml-sm">{{
								t(`enums.CONFIG_TAB.${menu.name}`)
							}}</span>
						</div>
					</q-tab>
				</q-tabs>
			</div>
			<div class="config-content">
				<q-tab-panels
					v-model="tab"
					animated
					swipeable
					vertical
					transition-prev="jump-up"
					transition-next="jump-up"
				>
					<q-tab-panel name="Metadata" class="q-pt-none">
						<MetadataComponent :app="app" />
					</q-tab-panel>
					<q-tab-panel name="Details">
						<EntrancesComponent :app="app" />
					</q-tab-panel>
					<q-tab-panel name="Specs">
						<SpecComponent :app="app" />
					</q-tab-panel>
					<q-tab-panel name="Permissions">
						<PermissionComponent :app="app" />
					</q-tab-panel>
					<q-tab-panel name="Options">
						<OptionsComponent :app="app" />
					</q-tab-panel>
				</q-tab-panels>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, PropType, watch } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { onBeforeRouteLeave } from 'vue-router';
import { useDevelopingApps } from '../stores/app';
import { ApplicationInfo } from '@devbox/core';
import MetadataComponent from './config/MetadataComponent.vue';
import EntrancesComponent from './config/EntrancesComponent.vue';
import SpecComponent from './config/SpecComponent.vue';
import PermissionComponent from './config/PermissionComponent.vue';
import OptionsComponent from './config/OptionsComponent.vue';
import DialogConfirm from './dialog/DialogConfirm.vue';

const { t } = useI18n();
const $q = useQuasar();
const store = useDevelopingApps();
const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	},
	downloading: {
		type: Boolean,
		default: false,
		required: false
	}
});
const tab = ref('Metadata');
const configStatus = ref(false);
const initCfg = ref();

onMounted(async () => {
	store.cfg = await store.getAppCfg(props.app.appName);
	initCfg.value = JSON.stringify(store.cfg);
});

onBeforeRouteLeave((to, from, next) => {
	if (configStatus.value) {
		const answer = window.confirm(
			'App configuration has been changed. Do you want to save the changes?'
		);
		if (answer) {
			onSave();
			next();
		}
	} else {
		next();
	}
});

watch(
	() => props.app.appName,
	async (newVal) => {
		store.cfg = await store.getAppCfg(newVal);
		initCfg.value = JSON.stringify(store.cfg);
	}
);

watch(
	() => store.cfg,
	(newVal) => {
		if (JSON.stringify(newVal) !== initCfg.value) {
			configStatus.value = true;
		} else {
			configStatus.value = false;
		}
	},
	{
		deep: true
	}
);

window.onbeforeunload = function (e) {
	if (configStatus.value && props.downloading) {
		var ev = window.event || e;
		ev.returnValue =
			'App configuration has been changed. Do you want to save the changes?';
	}
};

async function onSave() {
	if (localStorage.getItem('hideConfirmationAgain') === 'true') {
		__onSave();
		return false;
	}
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: 'Confirmation',
			message:
				'There are errors in TerminusManifest.yaml files. Your app may not be able to install correctly. Do you want to save any way?',
			showAgain: true,
			okText: 'Save'
		}
	})
		.onOk((checkbox) => {
			if (checkbox) {
				localStorage.setItem('hideConfirmationAgain', 'true');
			} else {
				localStorage.setItem('hideConfirmationAgain', 'false');
			}

			__onSave();
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
}

const __onSave = async () => {
	try {
		$q.loading.show();

		await store.putAppCfg(props.app.appName, store.cfg);
		store.cfg = await store.getAppCfg(props.app.appName);
		initCfg.value = JSON.stringify(store.cfg);
	} catch (e: any) {
		console.log(e);
		// $q.notify({
		//   type: 'negative',
		//   message: 'update app cfg failed: ' + e.message,
		// });
	} finally {
		$q.loading.hide();
	}
};
</script>
<style lang="scss">
.my-menu-class {
	border-radius: 8px;
}
.q-tab-panels {
	background-color: $background-1;
}
</style>

<style lang="scss" scoped>
.config-right {
	height: calc(100vh - 110px);
	flex: 1;
	border-radius: 12px;
	border: 1px solid $separator;
	overflow: hidden;
	background: $background-3;
	margin-top: 16px;

	.config-right-header {
		width: 100%;
		height: 32px;
		line-height: 32px;
		padding: 0 12px;
		border-bottom: 1px solid $separator;
		background: $background-3;
		.statusIcon {
			width: 6px;
			height: 6px;
			border-radius: 3px;
			display: inline-block;
		}
	}
	.config-right-content {
		height: calc(100% - 56px);
		background: $background-2;
		margin: 12px;
		padding: 32px;
		border-radius: 8px;
		display: flex;
		align-items: flex-start;
		justify-content: flex-start;
		overflow: scroll;

		.config-menu {
			width: 136px;
			position: -webkit-sticky;
			position: sticky;
			top: 0;
			.menuTab {
				width: 100%;
				padding-left: 8px;
				display: flex;
				align-items: flex-start;
				justify-content: flex-start;
			}
		}
		.config-content {
			flex: 1;
			max-width: 800px;
		}
	}
}
</style>
