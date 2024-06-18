<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="`Edit ${data.label}`"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row" v-if="selfEntrance.name === 'gpu'">
						<div class="form-item-key text-subtitle2 text-ink-1">
							GPU Required *
						</div>
						<div class="form-item-value">
							<q-toggle v-model="selfEntrance.require" />
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							Required {{ selfEntrance.label }}
							{{ selfEntrance.require ? '*' : '' }}
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfEntrance.required"
								:hint="`Minimum ${selfEntrance.label} required for the app.`"
								lazy-rules
								:disable="selfEntrance.require ? false : true"
								class="form-item-input"
								input-class="text-ink-2"
							>
								<template v-slot:append>
									<q-select
										dense
										borderless
										v-model="selfEntrance.requiredUnit"
										:options="
											selfEntrance.name === 'cpu'
												? requiredCpuOptions
												: requiredOptions
										"
										dropdown-icon="sym_r_keyboard_arrow_down"
										style="width: 50px"
									/>
								</template>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							Limited {{ selfEntrance.label }}
							{{ selfEntrance.require ? '*' : '' }}
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfEntrance.limited"
								:hint="`${selfEntrance.label} limit for the app. The app will be suspended if the resource limit is exceeded.`"
								lazy-rules
								:disable="selfEntrance.require ? false : true"
								color="teal-4"
								class="form-item-input"
								input-class="text-ink-2"
							>
								<template v-slot:append>
									<q-select
										dense
										borderless
										v-model="selfEntrance.limitUnit"
										:options="
											selfEntrance.name === 'cpu'
												? requiredCpuOptions
												: requiredOptions
										"
										dropdown-icon="sym_r_keyboard_arrow_down"
										style="width: 50px"
									/>
								</template>
							</q-input>
						</div>
					</div>
					<TerminusFormFooter />
				</q-form>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';
import { requiredCpuOptions, requiredOptions } from '../../types/constants';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const store = useDevelopingApps();
const show = ref(true);

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

const selfEntrance = ref(JSON.parse(JSON.stringify(props.data)));

const submit = () => {
	switch (props.data.name) {
		case 'cpu':
			store.cfg.spec.requiredCpu =
				selfEntrance.value.required + selfEntrance.value.requiredUnit;
			store.cfg.spec.limitedCpu =
				selfEntrance.value.limited + selfEntrance.value.limitUnit;
			break;

		case 'memory':
			store.cfg.spec.requiredMemory =
				selfEntrance.value.required + selfEntrance.value.requiredUnit;
			store.cfg.spec.limitedMemory =
				selfEntrance.value.limited + selfEntrance.value.limitUnit;
			break;

		case 'disk':
			store.cfg.spec.requiredDisk =
				selfEntrance.value.required + selfEntrance.value.requiredUnit;
			store.cfg.spec.limitedDisk =
				selfEntrance.value.limited + selfEntrance.value.limitUnit;
			break;

		case 'gpu':
			store.cfg.spec.requiredGpu =
				selfEntrance.value.required + selfEntrance.value.requiredUnit;
			store.cfg.spec.limitedGpu =
				selfEntrance.value.limited + selfEntrance.value.limitUnit;
			break;

		default:
			break;
	}

	onDialogOK();
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
	margin-bottom: 40px;
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
