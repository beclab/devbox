<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-continer">
			<terminus-dialog-bar
				:label="title"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<div class="dialog-desc">
				<q-form @submit="submit" @reset="onDialogCancel">
					<div class="form-item row">
						<div class="form-item-key text-subtitle2 text-grey-10">Name *</div>
						<div class="form-item-value">
							<q-input
								dense
								borderless
								no-error-icon
								v-model="selfDatabases.name"
								hint="Set the scope of the policy."
								lazy-rules
								:rules="[
									(val) => (val && val.length > 0) || 'Please input the name'
								]"
								color="teal-4"
								input-class="form-item-input"
							>
							</q-input>
						</div>
					</div>

					<div class="form-item row" v-if="name === 'postgres'">
						<div class="form-item-key text-subtitle2 text-grey-10">
							Distributed *
						</div>
						<div class="form-item-value">
							<q-select
								dense
								borderless
								:options="databasesOptions"
								v-model="selfDatabases.distributed"
								class="form-item-input"
							>
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
import { useDialogPluginComponent } from 'quasar';
import { useDevelopingApps } from '../../stores/app';
import { databasesOptions } from '../../types/constants';

import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const show = ref(true);
const store = useDevelopingApps();

const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	},
	name: {
		type: String,
		required: false,
		default: 'create'
	},
	mode: {
		type: String,
		required: false,
		default: 'create'
	}
});

const selfDatabases = ref(JSON.parse(JSON.stringify(props.data)));

if (selfDatabases.value.distributed === true) {
	selfDatabases.value.distributed = 'Enable';
} else {
	selfDatabases.value.distributed = 'Disable';
}

const title = computed(() => {
	if (props.mode === 'create') {
		return 'Add Databases';
	} else {
		return 'Edit Databases';
	}
});

const submit = () => {
	selfDatabases.value.distributed =
		selfDatabases.value.distributed === 'Enable' ? true : false;
	if (props.mode === 'create') {
		if (!store.cfg.middleware[props.name].databases) {
			store.cfg.middleware[props.name].databases = [];
		}
		store.cfg.middleware[props.name].databases.push(selfDatabases.value);
	} else {
		const policiesArr = store.cfg.middleware[props.name].databases;
		for (let i = 0; i < policiesArr.length; i++) {
			if (policiesArr[i].name === props.data.name) {
				policiesArr[i] = selfDatabases.value;
			}
		}
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
