<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="
					mode === 'create' ? t('dialog.addEntrance') : t('dialog.editEntrance')
				"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('type') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDependency.type"
								disable
								input-class="form-item-input text-ink-2"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('name') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDependency.name"
								:hint="t('config_space_addentrance_name_hint')"
								lazy-rules
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_space_addentrance_name_rules1'),
									(val) =>
										/^[a-z]/.test(val) ||
										t('config_space_addentrance_name_rules2'),
									(val) =>
										/^[a-z][a-z0-9]*$/.test(val) ||
										t('config_space_addentrance_name_rules3')
								]"
								input-class="form-item-input text-ink-2"
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
								v-model="selfDependency.version"
								:hint="t('config_space_addentrance_version_hint')"
								lazy-rules
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										t('config_space_addentrance_version_rules')
								]"
								input-class="form-item-input text-ink-2"
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
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const store = useDevelopingApps();
const show = ref(true);
const { t } = useI18n();

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
