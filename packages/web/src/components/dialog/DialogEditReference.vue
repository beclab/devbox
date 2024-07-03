<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="title"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<div class="form-item row">
					<div class="form-item-key text-subtitle2 text-ink-1">
						{{ t('name') }} *
					</div>
					<div class="form-item-value">
						<q-input
							dense
							borderless
							no-error-icon
							hint="The app name of the client app."
							v-model="selfSupportData"
							input-class="form-item-input text-ink-2"
							counter
							maxlength="30"
							@update:model-value="updateSupportClient"
							:rules="[
								(val) =>
									(val && val.length > 0) ||
									t('config_option_Reference_name_rules'),
								(val) =>
									/^[a-z]/.test(val) ||
									t('config_option_Reference_name_rules2'),
								(val) =>
									/^[a-z][a-z0-9]*$/.test(val) ||
									t('config_option_Reference_name_rules3')
							]"
						>
						</q-input>
					</div>
				</div>
			</div>

			<terminus-dialog-footer
				okText="Confirm"
				cancelText="Cancel"
				showCancel
				@close="onDialogCancel"
				@submit="submit"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusDialogFooter from '../common/TerminusDialogFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const props = defineProps({
	name: {
		type: String,
		required: false
	},
	mode: {
		type: String,
		required: false
	}
});

const { t } = useI18n();
const $q = useQuasar();
const show = ref(true);
const selfSupportData = ref(props.name);
const store = useDevelopingApps();

const updateSupportClient = (value) => {
	selfSupportData.value = value;
};

const title = computed(() => {
	if (props.mode === 'create') {
		return t('dialog.addClientReference');
	} else {
		return t('dialog.editClientReference');
	}
});

const submit = () => {
	if (!selfSupportData.value) {
		$q.notify({
			type: 'warning',
			message: 'The name is empty!'
		});

		return;
	}

	if (props.mode === 'create') {
		if (!store.cfg.options.appScope.appRef) {
			store.cfg.options.appScope.appRef = [];
		}
		store.cfg.options.appScope.appRef.push(selfSupportData.value);
	} else {
		const index = store.cfg.options.appScope.appRef.findIndex(
			(item) => item === props.name
		);
		store.cfg.options.appScope.appRef[index] = selfSupportData.value;
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
