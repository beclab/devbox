<template>
	<q-dialog class="card-dialog" ref="dialogRef">
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
						<div class="form-item-key text-subtitle2 text-ink-1">URL *</div>
						<div class="form-item-value q-mb-lg">
							<q-input
								dense
								borderless
								no-error-icon
								hint="URL of the client."
								v-model="selfSupportData.url"
								class="form-item-input"
								input-class="text-ink-2"
								@update:model-value="updateSupportClient"
								placeholder="https://"
								:rules="[
									(val) => (val && val.length > 0) || 'Please input the URL'
								]"
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
import { useDialogPluginComponent } from 'quasar';
import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const props = defineProps({
	data: {
		type: Object,
		required: false
	},
	mode: {
		type: String,
		required: false
	}
});

const selfSupportData = ref(JSON.parse(JSON.stringify(props.data)));
// const prefix = ref('https://');

const updateSupportClient = (value) => {
	selfSupportData.value.url = value;
};

const title = computed(() => {
	if (props.mode === 'create') {
		return `Add ${selfSupportData.value.label} Client`;
	} else {
		return `Edit ${selfSupportData.value.label} Client`;
	}
});

const submit = () => {
	console.log('submitsubmit', selfSupportData.value);
	onDialogOK(selfSupportData.value);
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
