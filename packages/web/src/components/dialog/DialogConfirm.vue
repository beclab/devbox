<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer">
			<terminus-dialog-bar
				:label="title"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">{{ message }}</div>

			<div class="dialog-checkbox" v-if="showAgain">
				<q-checkbox
					v-model="checkbox"
					color="green"
					label="Do not show this confirmation again."
				/>
			</div>

			<terminus-dialog-footer
				:okText="okText"
				cancelText="Cancel"
				showCancel
				@close="onDialogCancel"
				@submit="submit"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps } from 'vue';
import { useDialogPluginComponent } from 'quasar';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusDialogFooter from '../common/TerminusDialogFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const show = ref(true);
const checkbox = ref(false);

defineProps({
	title: {
		type: String,
		required: false,
		default: 'Alert'
	},
	message: {
		type: String,
		required: false,
		default: 'Some message'
	},
	showAgain: {
		type: Boolean,
		required: false,
		default: false
	},
	okText: {
		type: String,
		required: false,
		default: 'Confirm'
	}
});

const submit = () => {
	onDialogOK(checkbox.value);
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-desc {
			padding: 12px 20px;
		}
		.dialog-checkbox {
			padding: 0px 20px;
		}
	}
}
</style>
