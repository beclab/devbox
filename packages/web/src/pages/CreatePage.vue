<template>
	<div class="wrap">
		<div class="container row">
			<div class="col-sm-12 col-md-6 q-pr-lg">
				<div class="text-h3 text-grey-10">Create a new application</div>

				<div>
					<q-form @submit="onSubmit">
						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								App Name *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									hint="App’s namespace in Terminus system."
									v-model="config.name"
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) || 'Please input the app name.',
										(val) =>
											/^[a-z]/.test(val) ||
											'must start with an alphabetic character.',
										(val) =>
											/^[a-z][a-z0-9]*$/.test(val) ||
											'must contain only lowercase alphanumeric characters.'
									]"
									color="teal-4"
									class="form-item-input"
									counter
									maxlength="30"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								App Type *
							</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									v-model="config.type"
									:options="ApplicationTypeOptions"
									dropdown-icon="sym_r_keyboard_arrow_down"
									hint="Choose application type."
									class="form-item-input q-mt-md"
								>
								</q-select>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								Main Entrance Port *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									hint="Port of main entrance."
									v-model="config.websitePort"
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) ||
											'Please input the main entrance port',
										(val) =>
											(val > 0 && val <= 65535) ||
											'must be an int from 0 to 65535'
									]"
									color="teal-4"
									class="form-item-input"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								Image *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									hint="Image for app containers."
									v-model="config.img"
									lazy-rules
									:rules="[
										(val) => (val && val.length > 0) || 'Please input the image'
									]"
									color="teal-4"
									class="form-item-input"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								Port *
							</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									v-model="config.ports"
									use-input
									use-chips
									multiple
									no-error-icon
									hide-dropdown-icon
									input-debounce="0"
									@new-value="createPort"
									class="form-item-input"
									hint="Specify ports that need to be exposed."
									:rules="[
										(vals) =>
											vals.find((val) => val < 0 || val > 65535) &&
											'must be an int from 0 to 65535'
									]"
								>
									<template v-slot:selected-item="scope">
										<q-chip
											square
											icon-remove="sym_r_close"
											removable
											@remove="scope.removeAtIndex(scope.index)"
											:tabindex="scope.tabindex"
											class="q-ma-none tagChip"
										>
											{{ scope.opt }}
										</q-chip>
									</template>
								</q-select>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								Required Memory *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									hint="Requested memory resources for the app."
									v-model.number="config.requiredMemory"
									lazy-rules
									:rules="[
										(val) => val > 0 || 'must be a number greater than 0.'
									]"
									color="teal-4"
									class="form-item-input"
								>
									<template v-slot:append>
										<q-select
											dense
											borderless
											v-model="requiredMemoryUnit"
											:options="requiredOptions"
											dropdown-icon="sym_r_keyboard_arrow_down"
											style="width: 50px"
										/>
									</template>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-grey-10">
								Required GPU
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model.number="config.requiredGpu"
									lazy-rules
									hint="Requested GPU memory resources for the app."
									color="teal-4"
									class="form-item-input"
									placeholder="Leave empty if no GPU required."
								>
									<template v-slot:append>
										<q-select
											dense
											borderless
											v-model="requiredGpuUnit"
											:options="requiredOptions"
											dropdown-icon="sym_r_keyboard_arrow_down"
											style="width: 50px"
										/>
									</template>
								</q-input>
							</div>
						</div>

						<div class="form-btn row items-center justify-between">
							<q-btn
								class="form-btn-cancel col-5"
								dense
								flat
								no-caps
								@click="cancel"
								label="Cancel"
								type="button"
								color="teal-6"
							/>
							<q-btn
								class="form-btn-create col-5"
								dense
								no-caps
								label="Create"
								type="submit"
								color="teal-6"
							/>
						</div>
					</q-form>
				</div>
			</div>
			<div class="col-sm-0 col-md-6 right">
				<div class="flur"></div>
				<img src="../assets/ill-1.png" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../stores/app';
import { useMenuStore } from '../stores/menu';
import { CreateApplicationConfig, ApplicationType } from '@devbox/core';
import { useRouter } from 'vue-router';

import { requiredOptions } from './../types/constants';

const $q = useQuasar();

const apps = useDevelopingApps();
const menuStore = useMenuStore();

const router = useRouter();

const ApplicationTypeOptions = [
	{
		label: 'App',
		value: 'app'
	},
	{
		label: 'Recommended',
		value: 'recommended'
	},
	{
		label: 'Model',
		value: 'model'
	},
	{
		label: 'Agent',
		value: 'agent'
	}
];

const config = ref<CreateApplicationConfig>({
	name: '',
	type: 'app',
	osVersion: '0.1.0',
	img: 'bytetrade/devbox-app:0.0.1',
	//devEnv: 'beclab/node-ts-dev',
	ports: [8080],
	websitePort: '8080',

	systemDB: false,
	redis: false,
	mongodb: false,
	postgreSQL: false,

	systemCall: false,
	ingressRouter: false,
	traefik: false,

	appData: true,
	appCache: true,
	userData: [],

	needGpu: false,
	requiredGpu: '',
	requiredMemory: ''
});

const requiredMemoryUnit = ref('Mi');
const requiredGpuUnit = ref('Gi');

function createPort(val: string, done: any) {
	// specific logic to eventually call done(...) -- or not

	const p = parseInt(val);
	if (!p) {
		$q.notify('port must be a number');
		return;
	}

	done(p, 'add-unique');
}

async function onSubmit() {
	const params = JSON.parse(JSON.stringify(config.value));

	params.requiredMemory = params.requiredMemory + requiredMemoryUnit.value;
	params.requiredGpu = params.requiredGpu + requiredGpuUnit.value;
	params.osVersion = '>=' + params.osVersion;

	// params.requiredMemory = params.requiredMemory + 'G';
	// params.requiredGpu = params.requiredGpu + 'G';

	$q.loading.show();
	try {
		const appId = await apps.createApplication(params);
		if (appId) {
			await apps.getApps();
			await router.push({ path: '/app/' + appId });
			await menuStore.updateApplications();
			menuStore.currentItem = '/app/' + appId;
		}
	} catch (e) {
		console.log(e);
	} finally {
		$q.loading.hide();
	}
}

const cancel = () => {
	router.back();
};
</script>

<style lang="scss" scoped>
.wrap {
	width: 100%;
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 44px;
	.container {
		width: 100%;
		max-width: 1280px;
		height: 100vh;
		padding-top: 56px;

		.my-special-class {
			border: 1px solid green !important;
		}

		.form-item {
			margin-top: 50px;
			.form-item-key {
				width: 160px;
				height: 40px;
				line-height: 40px;
			}
			.form-item-value {
				flex: 1;
			}
		}

		.tagChip {
			margin-right: 4px;
			border-radius: 10px;
			background: rgba(246, 246, 246, 1);
			color: rgba(31, 24, 20, 1);
			font-size: 12px;
		}
		.form-btn {
			margin-top: 72px;
			margin-bottom: 40px;
		}

		.right {
			position: relative;
			z-index: 0;
			img {
				width: 90%;
			}
			.flur {
				width: 80%;
				height: 200px;
				position: absolute;
				top: 0;
				left: 0;
				right: 0;
				margin: auto;
				border-radius: 0 0 100px 100px;
				background: var(--Teal-01, #bffff4);
				filter: blur(150px);
				z-index: -1;
			}
		}
	}
}
</style>
