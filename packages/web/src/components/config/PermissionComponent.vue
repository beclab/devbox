<template>
	<div class="column">
		<div class="text-h6 text-ink-1">
			{{ t('enums.CONFIG_TAB.PERMISSIONS') }}
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_space_appdata') }} *
			</div>
			<div class="form-item-value">
				<q-toggle color="teal-6" v-model="store.cfg.permission.appData" />
				<div class="text-body3 text-ink-2">
					{{ t('config_space_appdata.desc') }}
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_space_systemdata') }}
			</div>
			<div class="form-item-value">
				<div class="row items-center justify-between">
					<div class="text-subtitle2 text-ink-2">
						{{ t('config_space_systemdata_desc') }}
					</div>
					<q-btn
						class="add-btn"
						borderless
						flat
						no-caps
						color="teal-8"
						:label="t('config.btn.add')"
						@click="addProvider"
					/>
				</div>

				<provider-card
					v-for="(sys, index) in store.cfg.permission.sysData"
					:key="index"
					:data="sys"
				/>
			</div>
		</div>
	</div>

	<!-- <div class="column q-gutter-md">
    <div class="row">
      <q-checkbox v-model="store.cfg.permission.appData">AppData</q-checkbox>
    </div>

    <div class="row">
      <q-btn @click="newPermission()">New</q-btn>
    </div>
    <q-card
      v-for="(entrance, index) of store.cfg.permission.sysData"
      :key="`ei` + index"
      class="q-md-lg q-pa-md"
    >
      <q-item class="row justify-end">
        <q-btn @click="deletePermission(index)">Delete</q-btn>
      </q-item>
      <q-input
        filled
        v-model="store.cfg.permission.sysData[index].group"
        label="Group Name"
        lazy-rules
        :rules="[
          (val) => (val && val.length > 0) || 'Please input the app name',
        ]"
      />

      <q-input
        filled
        v-model="store.cfg.permission.sysData[index].dataType"
        label="Data Type"
        lazy-rules
        :rules="[
          (val) => (val && val.length > 0) || 'Please input the app name',
        ]"
      />

      <q-input
        filled
        v-model="store.cfg.permission.sysData[index].version"
        label="Version"
        lazy-rules
        :rules="[
          (val) => (val && val.length > 0) || 'Please input the app name',
        ]"
      />

      <q-select
        label="Ports"
        filled
        v-model="store.cfg.permission.sysData[index].ops"
        use-input
        use-chips
        multiple
        hide-dropdown-icon
        input-debounce="0"
        @new-value="createOp"
        style="width: 250px"
      />
    </q-card>
  </div> -->
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import ProviderCard from '../common/ProviderCard.vue';
import DialogEditProvider from '../dialog/DialogEditProvider.vue';

const $q = useQuasar();
const { t } = useI18n();
const store = useDevelopingApps();

const defaultProvider = ref({
	group: '',
	dataType: '',
	version: '',
	ops: []
});

const addProvider = () => {
	$q.dialog({
		component: DialogEditProvider,
		componentProps: {
			data: defaultProvider.value,
			mode: 'create'
		}
	})
		.onOk(() => {
			console.log('OK');
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

function createOp(val: string, done: any) {
	// specific logic to eventually call done(...) -- or not

	done(val, 'add-unique');
}

function newPermission() {
	store.cfg.permission.sysData.push({
		group: 'group name',
		dataType: 'data type',
		version: 'version',
		ops: []
	});
}

function deletePermission(index: number) {
	store.cfg.permission.sysData.splice(index, 1);
}
</script>

<style lang="scss" scoped>
.form-item {
	margin-top: 20px;
	.form-item-key {
		width: 160px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
}

.add-btn {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
