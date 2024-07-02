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
				<q-form @submit="submit" @reset="onCancle">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_permissions_datagroup') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								:hint="t('config_permissions_datagroup_hint')"
								v-model="selfSupportData.group"
								input-class="form-item-input text-ink-2"
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_permissions_datagroup_rules')
								]"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('config_permissions_datatype') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								hint="Type of required data."
								v-model="selfSupportData.dataType"
								input-class="form-item-input text-ink-2"
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_permissions_datatype_rules')
								]"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('version') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								:hint="t('config_permissions_version_hint')"
								v-model="selfSupportData.version"
								input-class="form-item-input text-ink-2"
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_permissions_version_rules')
								]"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('operations') }} *
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								v-model="selfSupportData.ops"
								use-input
								use-chips
								multiple
								no-error-icon
								hide-dropdown-icon
								input-debounce="0"
								@new-value="createPort"
								class="form-item-input"
								input-class="text-ink-2"
								:hint="t('config_permissions_operations_hint')"
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_permissions_operations_rules')
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

					<TerminusFormFooter />
				</q-form>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const { t } = useI18n();
const store = useDevelopingApps();

const props = defineProps({
	data: {
		type: Object as () => any,
		required: false
	},
	mode: {
		type: String,
		required: false
	}
});

const $q = useQuasar();
const show = ref(true);
const selfSupportData = ref(JSON.parse(JSON.stringify(props.data)));

const title = computed(() => {
	if (props.mode === 'create') {
		return t('dialog.title.addRequiredData');
	} else {
		return t('dialog.title.editRequiredData');
	}
});

const verifyDuplicate = () => {
	let hasSameData = store.cfg.permission.sysData.findIndex(
		(item) => item.group === selfSupportData.value.group
	);
	if (hasSameData >= 0) {
		$q.notify({
			type: 'warning',
			message: 'This group already exists!'
		});
		return false;
	}
	return true;
};

const onCancle = () => {
	onDialogCancel();
};

const submit = () => {
	if (!store.cfg.permission.sysData) {
		store.cfg.permission.sysData = [];
	}
	if (!verifyDuplicate()) {
		return false;
	}

	if (props.mode === 'create') {
		store.cfg.permission.sysData.push(selfSupportData.value);
	} else {
		const index = store.cfg.permission.sysData.findIndex(
			(item) => item.group === props.data.group
		);
		store.cfg.permission.sysData[index] = selfSupportData.value;
	}

	onDialogOK();
};

function createPort(val: string, done: any) {
	done(val, 'add-unique');
}
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
	.tagChip {
		margin-right: 4px;
		border-radius: 10px;
		background: rgba(246, 246, 246, 1);
		color: rgba(31, 24, 20, 1);
		font-size: 12px;
	}
}
</style>
