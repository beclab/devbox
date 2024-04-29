<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer">
			<terminus-dialog-bar
				:label="mode === 'create' ? 'Add Entrance' : 'Edit Entrance'"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">Type *</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDependency.type"
								disable
								color="teal-4"
								input-class="form-item-input"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">Name *</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDependency.name"
								hint="The app name of dependent app."
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
								input-class="form-item-input"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">
							Version *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDependency.version"
								hint="Required version."
								lazy-rules
								:rules="[
									(val) => (val && val.length > 0) || 'Please input the version'
								]"
								color="teal-4"
								input-class="form-item-input"
							>
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

const selfDependency = ref(JSON.parse(JSON.stringify(props.data)));

const submit = () => {
	if (props.mode === 'edit') {
		const dependenciesArr = store.cfg.options.dependencies;
		for (let i = 0; i < dependenciesArr.length; i++) {
			if (dependenciesArr[i].name === props.data.name) {
				dependenciesArr[i] = selfDependency.value;
			}
		}
	} else {
		if (!store.cfg.options.dependencies) {
			store.cfg.options.dependencies = [];
		}
		store.cfg.options.dependencies.push(selfDependency.value);
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
