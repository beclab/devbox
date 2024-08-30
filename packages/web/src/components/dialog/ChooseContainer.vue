<template>
	<q-dialog class="card-dialog" ref="dialogRef">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="t('dialog_title_bind')"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<div class="text-ink-2 text-body3">{{ t('containers_dev_env') }}</div>
				<q-select
					dense
					options-dense
					emit-value
					map-options
					borderless
					v-model="devEnv"
					:options="devOptions"
					dropdown-icon="sym_r_keyboard_arrow_down"
					class="form-item-input q-mt-xs"
					color="teal-6"
					popup-content-style="padding: 10px 0;"
				>
					<template v-slot:option="{ itemProps, opt }">
						<q-item dense v-bind="itemProps" class="select-item">
							<q-item-section class="select-section text-ink-2">
								<q-item-label>{{ opt.label }}</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</q-select>

				<div class="text-ink-2 text-body3 q-mt-md">
					{{ t('containers_select_env') }}
				</div>
				<q-select
					dense
					options-dense
					emit-value
					map-options
					borderless
					v-model="container"
					:options="containerList"
					dropdown-icon="sym_r_keyboard_arrow_down"
					class="form-item-input q-mt-xs"
					color="teal-6"
					popup-content-style="padding: 10px 0;"
				>
					<template v-slot:option="{ itemProps, opt }">
						<q-item dense v-bind="itemProps" class="select-item">
							<q-item-section class="select-section text-ink-2">
								<q-item-label>{{ opt.label }}</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</q-select>

				<div class="text-ink-2 text-body3 q-mt-md">
					{{ t('containers_input_name') }}
				</div>
				<q-input
					dense
					borderless
					no-error-icon
					v-model="devContainerName"
					lazy-rules
					color="teal-4"
					class="form-item-input q-mt-xs"
				>
				</q-input>
			</div>

			<terminus-dialog-footer
				:okText="t('btn_confirm')"
				:cancelText="t('btn_cancel')"
				showCancel
				@close="onDialogCancel"
				@submit="onOKClick"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useDialogPluginComponent, Loading } from 'quasar';
import { ref, watch, onMounted } from 'vue';
import axios from 'axios';
import { useI18n } from 'vue-i18n';
import { useDevelopingApps } from '../../stores/app';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusDialogFooter from '../common/TerminusDialogFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const store = useDevelopingApps();
const { t } = useI18n();

const devOptions = [
	{
		label: 'NodeJS',
		value: 'NodeJS'
	},
	{
		label: 'Golang',
		value: 'Golang'
	},
	{
		label: 'Python',
		value: 'Python'
	}
];

const devEnv = ref<string>('NodeJS');
const container = ref<number | null>(null);
const containerList = ref<any>([]);
const devContainerName = ref();
const unbindContainers = ref<any[]>([]);

defineEmits([...useDialogPluginComponent.emits]);

function update() {
	let res = [
		{
			label: 'Create New Container',
			value: null
		}
	];
	for (const container of unbindContainers.value) {
		if (container.appId || container.appName) {
			continue;
		}
		if (container.devEnv != devEnv.value) {
			continue;
		}

		res.push({
			label: container.id + ': ' + container.devEnv,
			value: container.id
		});
	}
	containerList.value = res;
	container.value = null;
}

const getUnbindContainers = async () => {
	try {
		unbindContainers.value = await axios.get(
			store.url + `/api/list-my-containers?unbind=true`
		);
	} catch (error) {
		unbindContainers.value = [];
	}
};

watch(
	() => devEnv.value,
	async () => {
		update();
	}
);

onMounted(async () => {
	await getUnbindContainers();
	await update();
});

async function onOKClick() {
	Loading.show();
	try {
		Loading.hide();
		onDialogOK({
			devEnv: devEnv.value,
			container: container.value,
			devContainerName: devContainerName.value
		});
	} catch (e) {
		console.log(e);
		Loading.hide();
	}
}
</script>
<style lang="scss">
.select-item {
	color: rgba(31, 24, 20, 1);
	margin: 0 12px;
	border-radius: 8px;
}
</style>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-desc {
			padding: 12px 20px;
		}
		.dialog-checkbox {
			padding: 0px 20px;
		}
	}
}
</style>
