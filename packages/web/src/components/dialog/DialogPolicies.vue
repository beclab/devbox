<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer">
			<terminus-dialog-bar
				:label="mode === 'create' ? 'Add Sub Policies' : 'Edit Sub Policies'"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">
							Policy Scope *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfEntrance.uriRegex"
								hint="Set the affected domain of this policy.  Regular expressions are supported."
								placeholder="Add effected URLs of the policy, regular expression supported"
								lazy-rules
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										'Please input the effected domain'
								]"
								color="teal-4"
								input-class="form-item-input"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">
							MFA Level *
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								:options="mfaLevelOptions"
								v-model="selfEntrance.level"
								class="form-item-input q-mt-md"
							>
							</q-select>
							<div class="text-body3 text-grey-5">
								Two-Factor requires additional credentials with an OTP (One-Time
								Password) to access the entrance.
							</div>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">
							One Time Valid
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								:options="visiblityOptions"
								v-model="oneTimeValid"
								class="form-item-input q-mt-md"
							>
							</q-select>
							<div class="text-body3 text-grey-5">
								Authentication is required every time to access this entrance.
							</div>
						</div>
					</div>

					<div class="form-item row" v-if="selfEntrance.oneTime === false">
						<div class="form-item-key text-subtitle2 text-grey-10">
							Valid Duration
						</div>

						<div class="form-item-value q-mb-lg">
							<q-input
								dense
								borderless
								no-error-icon
								v-model.number="validDurationValue"
								hint="Set the time period (in seconds) before a user is asked to MFA again. Leave empty for one time valid."
								lazy-rules
								color="teal-4"
								class="form-item-input"
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
						<div class="form-item-key text-subtitle2 text-grey-10">
							Description
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								type="textarea"
								no-error-icon
								v-model="selfEntrance.description"
								hint="A brief description of this policy."
								lazy-rules
								color="teal-4"
								input-class="form-item-input"
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
