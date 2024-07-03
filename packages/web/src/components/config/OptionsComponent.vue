<template>
	<div class="column">
		<div class="text-h6 text-ink-1">{{ t('enums.OPTIONS') }}</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_option_cluster') }}
			</div>
			<div class="form-item-value">
				<q-toggle
					color="teal-6"
					v-model="store.cfg.options.appScope.clusterScoped"
				/>
				<div class="text-body3 text-ink-2">
					{{ t('config_option_cluster_desc') }}
				</div>
			</div>
		</div>

		<div class="form-item row" v-if="store.cfg.options.appScope.clusterScoped">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_space_client') }}
			</div>
			<div class="form-item-value">
				<div class="row items-center justify-between">
					<div class="text-subtitle2 text-ink-2">
						{{ t('config_space_client_desc') }}
					</div>
					<q-btn
						class="add-btn"
						borderless
						flat
						no-caps
						color="teal-8"
						label="Add"
						@click="addReference"
					/>
				</div>

				<reference-card
					v-for="(app, index) in store.cfg.options.appScope.appRef"
					:key="index"
					:name="app"
				/>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_space_analytics') }}
			</div>
			<div class="form-item-value">
				<q-toggle
					color="teal-6"
					v-model="store.cfg.options.analytics.enabled"
				/>
				<div class="text-body3 text-ink-2">
					{{ t('config_space_analytics_desc') }}
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				Enable Websocket
			</div>
			<div class="form-item-value">
				<q-toggle
					color="teal-6"
					v-model="websocketToggle"
					@update:model-value="updateWebsocket"
				/>
				<div class="text-body3 text-ink-2">
					{{ t('config_space_websocket_desc') }}
				</div>

				<template v-if="websocketToggle">
					<div class="text-body3 text-ink-1 q-mt-md q-mb-sm">
						{{ t('config_space_port') }} *
					</div>
					<q-input
						dense
						borderless
						no-error-icon
						lazy-rules
						input-class="form-item-input text-ink-2"
						v-model.number="websocket.port"
						:rules="[
							(val) =>
								(val > 0 && val <= 65535) || 'must be an int from 0 to 65535'
						]"
						@update:model-value="updatePort"
					>
					</q-input>
					<div class="text-body3 text-ink-1 q-mt-md q-mb-sm">
						{{ t('config_space_url') }} *
					</div>
					<q-input
						dense
						borderless
						no-error-icon
						lazy-rules
						v-model="websocket.url"
						input-class="form-item-input text-ink-2"
						placeholder="/ws/"
						@update:model-value="updateUrl"
					>
					</q-input>
				</template>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useDevelopingApps } from '../../stores/app';
import ReferenceCard from '../common/ReferenceCard.vue';
import DialogEditReference from '../dialog/DialogEditReference.vue';

const { t } = useI18n();
const store = useDevelopingApps();

const levelOptions = [
	{
		label: 'Public',
		value: 'public'
	},
	{
		label: 'One Factor',
		value: 'one_factor'
	},
	{
		label: 'Two Factor',
		value: 'two_factor'
	},
	{
		label: 'Deny',
		value: 'deny'
	}
];

const typeOptions = [
	{
		label: 'Application',
		value: 'application'
	},
	{
		label: 'System',
		value: 'system'
	}
];

const $q = useQuasar();
const websocketToggle = ref(false);
const websocket = ref({
	port: '',
	url: ''
});

const updateWebsocketToggle = () => {
	if (store.cfg.options.websocket) {
		websocketToggle.value = true;
		websocket.value = {
			port: store.cfg.options.websocket.port,
			url: store.cfg.options.websocket.url
		};
	} else {
		websocketToggle.value = false;
	}
};

const updatePort = (value) => {
	websocket.value.port = value;
	store.cfg.options.websocket.port = value;
};

const updateUrl = (value) => {
	websocket.value.url = value;
	store.cfg.options.websocket.url = value;
};

const updateWebsocket = (value) => {
	websocketToggle.value = value;
	if (!value) {
		delete store.cfg.options.websocket;
	} else {
		store.cfg.options.websocket = websocket.value;
	}
};

const addReference = () => {
	$q.dialog({
		component: DialogEditReference,
		componentProps: {
			name: '',
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

function addPolicy() {
	store.cfg.options.policies.push({
		entranceName: 'entrance name',
		description: 'policy description',
		uriRegex: 'uri regex',
		level: 'two_factor',
		oneTime: false,
		validDuration: '5s'
	});
}

function deletePolicy(index: number) {
	store.cfg.options.policies.splice(index, 1);
}

function addDependencies() {
	store.cfg.options.dependencies.push({
		name: 'app name',
		version: '>=0.1.0',
		type: 'application'
	});
}

function deleteDependencies(index: number) {
	store.cfg.options.dependencies.splice(index, 1);
}

function createAppRef(val: string, done: any) {
	// specific logic to eventually call done(...) -- or not

	done(val, 'add-unique');
}

onMounted(() => {
	updateWebsocketToggle();
});
</script>

<style lang="scss" scoped>
.form-item {
	margin-top: 20px;
	.form-item-key {
		width: 160px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
}

.add-btn {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
