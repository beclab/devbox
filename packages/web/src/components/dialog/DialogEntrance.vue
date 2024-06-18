<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer" flat>
			<q-scroll-area style="height: 600px">
				<terminus-dialog-bar
					:label="mode === 'create' ? 'Add Entrances' : 'Edit Entrances'"
					icon=""
					titAlign="text-left"
					@close="onDialogCancel"
				/>

				<div class="dialog-desc">
					<q-form @submit="submit" @reset="onCancel">
						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Entrance Name *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model="selfEntrance.name"
									hint="Assign a unique name for this entrance."
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) ||
											'Please input the entrance name',
										(val) =>
											/^[a-z]([-a-z0-9]*[a-z0-9])?$/.test(val) ||
											'must contain only lowercase alphanumeric characters and hyphens.'
									]"
									input-class="form-item-input text-ink-2"
									counter
									maxlength="63"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Entrance Title *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model="selfEntrance.title"
									hint="Title that appears on the Terminus desktop after installation."
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) ||
											'Please input the entrance title'
									]"
									input-class="form-item-input text-ink-2"
									counter
									maxlength="30"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Entrance Icon
							</div>
							<div class="form-item-value">
								<upload-icon
									:default-img="selfEntrance.icon"
									:max-size="512 * 1024"
									accept=".png, .webp"
									:acceptW="256"
									:acceptH="256"
									@uploaded="uploaded"
								/>

								<div class="text-ink-2 q-mt-sm" style="font-size: 11px">
									Icon that appears in the Terminus desktop after installed.
								</div>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Host Name *
							</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model="selfEntrance.host"
									hint="Ingress name for this entrance."
									lazy-rules
									:rules="[
										(val) =>
											(val && val.length > 0) || 'Please input the host name',
										(val) =>
											/^[a-z]([-a-z0-9]*[a-z0-9])?$/.test(val) ||
											'must contain only lowercase alphanumeric characters and hyphens.'
									]"
									counter
									input-class="form-item-input text-ink-2"
									maxlength="63"
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">Port *</div>
							<div class="form-item-value">
								<q-input
									dense
									borderless
									no-error-icon
									v-model.number="selfEntrance.port"
									lazy-rules
									input-class="form-item-input text-ink-2"
									hint="Port for this entrance."
								>
								</q-input>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">Visible</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									:options="visiblityOptions"
									v-model="invisibleSelf"
									dropdown-icon="sym_r_keyboard_arrow_down"
									hint="Show entrance icon and title on the Terminus desktop."
									class="form-item-input text-ink-2"
								>
								</q-select>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Auth Level
							</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									v-model="selfEntrance.authLevel"
									:options="authLevelOptions"
									dropdown-icon="sym_r_keyboard_arrow_down"
									hint="A private entrance requires activating Tailscale for access."
									class="form-item-input q-mt-md text-ink-2"
								>
								</q-select>
							</div>
						</div>

						<div class="form-item row">
							<div class="form-item-key text-subtitle2 text-ink-1">
								Open Method:
							</div>
							<div class="form-item-value">
								<q-select
									dense
									borderless
									:options="openWindowOptions"
									v-model="selfEntrance.openMethod"
									emit-value
									map-options
									dropdown-icon="sym_r_keyboard_arrow_down"
									hint="Show entrance icon and title on the Terminus desktop."
									class="form-item-input text-ink-2"
								>
								</q-select>
							</div>
						</div>
						<div class="q-mt-xl">
							<TerminusFormFooter />
						</div>
					</q-form>
				</div>
			</q-scroll-area>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import { authLevelOptions, visiblityOptions } from '../../types/constants';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';
import UploadIcon from '../common/UploadIcon.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const store = useDevelopingApps();
const show = ref(true);

const openWindowOptions = [
	{
		label: 'Default',
		value: 'default'
	},
	{
		label: 'Window',
		value: 'window'
	},
	{
		label: 'iFrame',
		value: 'iframe'
	}
];

const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	},
	mode: {
		type: String,
		required: false,
		default: 'create'
	}
});

const invisibleSelf = ref('True');
const selfEntrance = ref(JSON.parse(JSON.stringify(props.data)));

if (selfEntrance.value.invisible) {
	invisibleSelf.value = 'True';
} else {
	invisibleSelf.value = 'False';
}

const uploaded = (url) => {
	selfEntrance.value.icon = url;
};

const submit = () => {
	if (invisibleSelf.value === 'True') {
		selfEntrance.value.invisible = true;
	} else {
		selfEntrance.value.invisible = false;
	}

	if (
		props.data.name &&
		selfEntrance.value.name !== props.data.name &&
		store.cfg.options.policies &&
		store.cfg.options.policies.length > 0
	) {
		for (let i = 0; i < store.cfg.options.policies.length; i++) {
			const policie = store.cfg.options.policies[i];
			policie.entranceName = selfEntrance.value.name;
		}
	}

	if (props.mode === 'edit') {
		const index = store.cfg.entrances.findIndex(
			(item) => item.name === props.data.name
		);
		store.cfg.entrances[index] = selfEntrance.value;
	} else {
		store.cfg.entrances.push(selfEntrance.value);
	}

	onDialogOK();
};

const onCancel = () => {
	onDialogCancel();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 720px;
		border-radius: 12px;

		.dialog-desc {
			padding-left: 32px;
			padding-right: 32px;
		}
	}
}

.form-item {
	margin-top: 20px;
	.form-item-key {
		width: 140px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
