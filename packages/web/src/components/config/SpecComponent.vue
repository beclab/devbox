<template>
	<div class="column">
		<div class="text-h6 text-grey-10">Specs</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Entrances *</div>
			<div class="form-item-value">
				<div class="item-explain row items-center justify-between">
					<span class="text-subtitle2 text-grey-8"
						>Specify how to access this app, at least 1 required.
					</span>
					<q-btn
						borderless
						flat
						no-caps
						@click="addEntrance"
						color="teal-8"
						style="border: 1px solid rgba(235, 235, 235, 1); border-radius: 8px"
						label="Add Entrance"
					/>
				</div>

				<entrances-card
					v-for="(entrance, index) in store.cfg.entrances"
					:key="index"
					:data="entrance"
					:disabledRemove="store.cfg.entrances.length <= 1"
				/>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Resources *</div>
			<div class="form-item-value">
				<div class="item-explain row items-center justify-between">
					<span class="text-subtitle2 text-grey-8"
						>Specify requested and limited resources for your app.
					</span>
				</div>

				<resources-card
					v-for="(option, index) in resourcesOptions"
					:key="index"
					:data="option"
					@updateResources="updateResources"
				/>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Middleware</div>
			<div class="form-item-value">
				<div class="item-explain row items-center justify-between">
					<span class="text-subtitle2 text-grey-8"
						>Add the necessary middleware for your app.
					</span>

					<q-btn-dropdown
						borderless
						flat
						no-caps
						color="teal-8"
						style="border: 1px solid rgba(235, 235, 235, 1); border-radius: 8px"
						label="Add"
						dropdown-icon="sym_r_keyboard_arrow_down"
					>
						<q-list>
							<q-item
								clickable
								v-close-popup
								:disable="
									store.cfg.middleware && store.cfg.middleware[option]
										? true
										: false
								"
								@click="addMiddleware(option)"
								v-for="option in middlewareOptions"
								:key="option"
							>
								<q-item-section>
									<q-item-label>{{ option }}</q-item-label>
								</q-item-section>
							</q-item>
						</q-list>
					</q-btn-dropdown>
				</div>

				<template v-if="store.cfg.middleware">
					<middleware-card
						v-for="(value, key) in store.cfg.middleware"
						:key="key"
						:name="key"
						:data="value"
					/>
				</template>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Dependencies *
			</div>
			<div class="form-item-value">
				<div class="item-explain row items-center justify-between">
					<span class="text-subtitle2 text-grey-8" style="flex: 1"
						>Indicate if your app depends on other apps or requires a specific
						OS version.
					</span>

					<q-btn
						borderless
						flat
						no-caps
						color="teal-8"
						style="border: 1px solid rgba(235, 235, 235, 1); border-radius: 8px"
						label="Add Dependencies"
						@click="addDependency"
					/>
				</div>

				<template
					v-if="
						store.cfg.options.dependencies &&
						store.cfg.options.dependencies.length > 0
					"
				>
					<dependencies-card
						v-for="(dependencie, index) in store.cfg.options.dependencies"
						:key="index"
						:data="dependencie"
					/>
				</template>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import { middlewareOptions } from '../../types/constants';

import EntrancesCard from '../common/EntrancesCard.vue';
import ResourcesCard from '../common/ResourcesCard.vue';
import MiddlewareCard from '../common/MiddlewareCard.vue';
import DependenciesCard from '../common/DependenciesCard.vue';
import DialogEntrance from '../dialog/DialogEntrance.vue';
import DialogDependency from '../dialog/DialogDependency.vue';

import DialogMiddleware from '../dialog/DialogMiddleware.vue';

const $q = useQuasar();
const store = useDevelopingApps();
const defaultEntrances = ref({
	name: '',
	host: '',
	port: '8080',
	icon: '',
	title: '',
	authLevel: 'private',
	invisible: false,
	openMethod: 'default'
});

const defaultDependency = ref({
	name: '',
	version: '',
	type: 'application'
});

const resourcesOptions = ref([
	{
		label: 'CPU',
		name: 'cpu',
		required: '',
		limited: '',
		requiredUnit: 'm',
		limitUnit: 'm',
		require: true
	},
	{
		label: 'Memory',
		name: 'memory',
		required: '',
		limited: '',
		requiredUnit: 'Mi',
		limitUnit: 'Mi',
		require: true
	},
	{
		label: 'Disk',
		name: 'disk',
		required: '',
		limited: '',
		requiredUnit: 'Mi',
		limitUnit: 'Mi',
		require: true
	},
	{
		label: 'GPU',
		name: 'gpu',
		required: '',
		limited: '',
		requiredUnit: 'Mi',
		limitUnit: 'Mi',
		require: false
	}
]);

const initMiddleware = (option) => {
	return {
		name: option,
		username: '',
		passsword: '',
		databases: null
	};
};

const addMiddleware = (option) => {
	$q.dialog({
		component: DialogMiddleware,
		componentProps: {
			data: initMiddleware(option),
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

const addEntrance = () => {
	$q.dialog({
		component: DialogEntrance,
		componentProps: {
			data: defaultEntrances.value
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

const addDependency = () => {
	$q.dialog({
		component: DialogDependency,
		componentProps: {
			data: defaultDependency.value,
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

const updateResources = () => {
	for (let i = 0; i < resourcesOptions.value.length; i++) {
		const el = resourcesOptions.value[i];
		switch (el.name) {
			case 'cpu':
				el.required = parseFloat(store.cfg.spec.requiredCpu);
				el.requiredUnit = store.cfg.spec.requiredCpu
					? store.cfg.spec.requiredCpu.replace(/[^a-zA-Z]/g, '')
					: 'm';
				el.limited = parseFloat(store.cfg.spec.limitedCpu);
				el.limitUnit = store.cfg.spec.limitedCpu
					? store.cfg.spec.limitedCpu.replace(/[^a-zA-Z]/g, '')
					: 'm';
				// alert(
				//   store.cfg.spec.requiredCpu &&
				//     store.cfg.spec.requiredCpu.replace(/[^a-zA-Z]/g, '')
				// );
				break;

			case 'memory':
				el.required = parseFloat(store.cfg.spec.requiredMemory);
				el.requiredUnit = store.cfg.spec.requiredMemory
					? store.cfg.spec.requiredMemory.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				el.limited = parseFloat(store.cfg.spec.limitedMemory);
				el.limitUnit = store.cfg.spec.limitedMemory
					? store.cfg.spec.limitedMemory.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				break;

			case 'disk':
				el.required = parseFloat(store.cfg.spec.requiredDisk);
				el.requiredUnit = store.cfg.spec.requiredDisk
					? store.cfg.spec.requiredDisk.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				el.limited = parseFloat(store.cfg.spec.limitedDisk);
				el.limitUnit = store.cfg.spec.limitedDisk
					? store.cfg.spec.limitedDisk.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				break;

			case 'gpu':
				el.required = parseFloat(store.cfg.spec.requiredGp);
				el.requiredUnit = store.cfg.spec.requiredGpu
					? store.cfg.spec.requiredGpu.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				el.limited = parseFloat(store.cfg.spec.limitedGpu);
				el.limitUnit = store.cfg.spec.limitedGpu
					? store.cfg.spec.limitedGpu.replace(/[^a-zA-Z]/g, '')
					: 'Mi';
				break;
		}
	}
};

onMounted(() => {
	updateResources();
});
</script>

<style lang="scss" scoped>
.form-item {
	margin-top: 32px;
	.form-item-key {
		width: 160px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
