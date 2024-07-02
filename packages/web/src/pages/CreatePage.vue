<template>
	<div class="wrap">
		<div class="container row">
			<div class="col-sm-12 col-md-6 q-pr-lg">
				<div class="text-h3 text-ink-1">{{ t('home_create') }}</div>

				<div>
					<q-form @submit="onSubmit">
						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_appname') }} *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									:hint="t('home_appname_hint')"
									v-model="config.name"
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) || t('home_appname_rules_1'),
										(val) => /^[a-z]/.test(val) || t('home_appname_rules_2'),
										(val) =>
											/^[a-z][a-z0-9]*$/.test(val) || t('home_appname_rules_3')
									]"
									class="form-item-input"
									input-class="text-ink-2"
									counter
									maxlength="30"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_apptype') }} *
							</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									v-model="config.type"
									:options="ApplicationTypeOptions"
									dropdown-icon="sym_r_keyboard_arrow_down"
									:hint="t('home_apptype_hint')"
									color="text-ink-2"
									class="form-item-input q-mt-md"
								>
									<template v-slot:selected-item="scope">
										<span class="text-ink-2">
											{{ scope.opt.label || config.type }}
										</span>
									</template>
								</q-select>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_entrance_port') }} *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									:hint="t('home_entrance_port_hint')"
									v-model="config.websitePort"
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) ||
											t('home_entrance_port_rules_1'),
										(val) =>
											(val > 0 && val <= 65535) ||
											t('home_entrance_port_rules_2')
									]"
									class="form-item-input"
									input-class="text-ink-2"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_image') }} *
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
									class="form-item-input"
									input-class="text-ink-2"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_port') }} *
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
									:hint="t('home_port_hint')"
									:rules="[
										(vals) =>
											vals.find((val) => val < 0 || val > 65535) &&
											t('home_port_rules')
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
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_memory') }} *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									:hint="t('home_memory_hint')"
									v-model.number="config.requiredMemory"
									lazy-rules
									:rules="[(val) => val > 0 || t('home_memory_rules')]"
									class="form-item-input"
									input-class="text-ink-2"
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
							<div class="form-item-key text-subtitle2 text-ink-1">
								{{ t('home_gpu') }}
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model.number="config.requiredGpu"
									lazy-rules
									:hint="t('home_gpu_hint')"
									color="ink-2"
									class="form-item-input"
									input-class="text-ink-2"
									:placeholder="t('home_gpu_place')"
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
								:label="t('cancel')"
								type="button"
								color="teal-6"
							/>
							<q-btn
								class="form-btn-create col-5"
								dense
								no-caps
								:label="t('create')"
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
import { useI18n } from 'vue-i18n';

import { requiredOptions } from './../types/constants';

const { t } = useI18n();
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
			background: $background-hover;
			color: $ink-2;
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
