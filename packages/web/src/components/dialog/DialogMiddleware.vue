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
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('username') }} *
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfMiddleware.username"
								lazy-rules
								:rules="[
									(val) =>
										(val && val.length > 0) ||
										`Please input the required username`
								]"
								class="form-item-input"
								input-class="text-ink-2"
								counter
								maxlength="128"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-ink-1">
							{{ t('password') }}
						</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfMiddleware.password"
								:placeholder="t('config_space_postgres_password_place')"
								lazy-rules
								class="form-item-input"
								input-class="text-ink-2"
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
import { ref, defineProps, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const { t } = useI18n();
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

const selfMiddleware = ref(JSON.parse(JSON.stringify(props.data)));

const title = computed(() => {
	if (props.mode === 'create') {
		return t('addMiddle', { type: props.data.name });
	} else {
		return t('editMiddle', { type: props.data.name });
	}
});

const submit = () => {
	if (!store.cfg.middleware) {
		store.cfg.middleware = {};
	}
	store.cfg.middleware[props.data.name] = {
		username: selfMiddleware.value.username,
		password: selfMiddleware.value.password,
		databases: selfMiddleware.value.databases || null
	};

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
