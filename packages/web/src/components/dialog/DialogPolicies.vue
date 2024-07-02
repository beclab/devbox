<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="
					mode === 'create'
						? t('dialog.title.addSubPolicies')
						: t('dialog.title.editSubPolicies')
				"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_space_policy') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfEntrance.uriRegex"
								:hint="t('config_space_policy_hint')"
								:placeholder="t('config_space_policy_place')"
								lazy-rules
								:rules="[
									(val) =>
										(val && val.length > 0) || t('config_space_policy_rules')
								]"
								input-class="form-item-input text-ink-2"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_space_mfalevel') }} *
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								:options="mfaLevelOptions"
								v-model="selfEntrance.level"
								class="form-item-input q-mt-md"
								input-class="text-ink-2"
							>
							</q-select>
							<div class="text-body3 text-ink-2">
								{{ t('.config_space_mfalevel_hint') }}
							</div>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_space_onetimevalid') }}
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								:options="visiblityOptions"
								v-model="oneTimeValid"
								class="form-item-input q-mt-md"
								input-class="text-ink-2"
							>
							</q-select>
							<div class="text-body3 text-ink-2">
								Authentication is required every time to access this entrance.
							</div>
						</div>
					</div>

					<div class="form-item row" v-if="selfEntrance.oneTime === false">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_space_validduration') }}
						</div>

						<div class="form-item-value q-mb-lg">
							<q-input
								dense
								borderless
								no-error-icon
								v-model.number="validDurationValue"
								:hint="t('config_space_validduration_hint')"
								lazy-rules
								class="form-item-input"
								input-class="text-ink-2"
								:rules="[
									(val) => val > 0 || 'must be a number greater than 0.'
								]"
							>
								<template v-slot:append>
									<q-select
										dense
										borderless
										v-model="validDurationUnit"
										:options="validDuration"
										dropdown-icon="sym_r_keyboard_arrow_down"
										style="width: 50px"
									/>
								</template>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_space_description') }}
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								type="textarea"
								no-error-icon
								v-model="selfEntrance.description"
								:hint="t('config_space_description_hint')"
								lazy-rules
								input-class="form-item-input text-ink-2"
								counter
								maxlength="512"
							>
							</q-input>
						</div>
					</div>
					<TerminusFormFooter />
				</q-form>
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
import { ref, defineProps, watch } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';
import { useI18n } from 'vue-i18n';
import {
	mfaLevelOptions,
	visiblityOptions,
	validDuration
} from '../../types/constants';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

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

const { t } = useI18n();
const store = useDevelopingApps();
const show = ref(true);
const selfEntrance = ref(JSON.parse(JSON.stringify(props.data)));
const oneTimeValid = ref(selfEntrance.value.oneTime ? 'True' : 'False');
const validDurationUnit = ref(
	selfEntrance.value.validDuration
		? selfEntrance.value.validDuration.replace(/\d/g, '')
		: 'ms'
);
const validDurationValue = ref(
	selfEntrance.value.validDuration
		? parseFloat(selfEntrance.value.validDuration)
		: 0
);

watch(
	() => oneTimeValid.value,
	(newVal) => {
		if (newVal === 'True') {
			selfEntrance.value.oneTime = true;
		} else {
			selfEntrance.value.oneTime = false;
		}
	}
);

const submit = () => {
	selfEntrance.value.validDuration =
		validDurationValue.value + validDurationUnit.value;
	if (props.mode === 'edit') {
		const policiesArr = store.cfg.options.policies;
		for (let i = 0; i < policiesArr.length; i++) {
			if (policiesArr[i].entranceName === selfEntrance.value.entranceName) {
				policiesArr[i] = selfEntrance.value;
			}
		}
	} else {
		if (!store.cfg.options.policies) {
			store.cfg.options.policies = [];
		}
		store.cfg.options.policies.push(selfEntrance.value);
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
