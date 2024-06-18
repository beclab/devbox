<template>
	<div class="dev-container">
		<!-- <div class="text-h3 text-ink-1">Dev Container List</div> -->
		<div class="container" v-if="containers && containers.length > 0">
			<container-card
				v-for="(container, index) in containers"
				:key="index"
				mode="application"
				:container="container"
				@bindContainer="bindContainer"
				@unbindContainer="unbindContainer"
			/>
		</div>
		<div class="nodata" v-else>
			<img src="../assets/nodata.svg" />
			<span class="q-mt-xl">No data.</span>
		</div>
	</div>

	<!-- <div class="q-pa-md column">
    <q-table
      flat
      bordered
      title="Sender"
      :rows="containers"
      :columns="columns"
      row-key="id"
      wrap-cells
      :pagination="initialPagination"
    >
      <template v-slot:body-cell-action="props">
        <q-td :props="props">
          <q-btn
            v-if="!props.row.id"
            label="Bind"
            class="col-1"
            flat
            color="primary"
            @click="bindContainer(props.row)"
          />

          <q-btn
            v-if="props.row.id"
            label="UnBind"
            class="col-1"
            flat
            color="primary"
            @click="unbindContainer(props.row)"
          />
        </q-td>
      </template>
    </q-table>
  </div> -->
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, PropType } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
import { useRoute } from 'vue-router';
import { useDevelopingApps } from '../stores/app';
import { ApplicationInfo, Container } from '@devbox/core';
import ChooseContainer from './dialog/ChooseContainer.vue';

import ContainerCard from './common/ContainerCard.vue';

const store = useDevelopingApps();
const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	}
});
const $q = useQuasar();

const initialPagination = {
	sortBy: 'desc',
	descending: false,
	page: 1,
	rowsPerPage: 1000
};

const containers = ref<Container[]>([]);

onMounted(async () => {
	containers.value = await store.getAppContainer(props.app.appName);
});

const columns = [
	{
		name: 'image',
		align: 'left',
		label: 'Name',
		field: 'image',
		sortable: false
	},
	{
		name: 'podSelector',
		align: 'left',
		label: 'podSelector',
		field: 'podSelector',
		sortable: false
	},
	{
		name: 'containerName',
		align: 'left',
		label: 'containerName',
		field: 'containerName',
		sortable: false
	},
	{
		name: 'id',
		align: 'left',
		label: 'Container id',
		field: 'id',
		sortable: false
	},
	{
		name: 'state',
		align: 'left',
		label: 'state',
		field: 'state',
		sortable: false
	},
	{
		name: 'devEnv',
		align: 'left',
		label: 'devEnv',
		field: 'devEnv',
		sortable: false
	},
	{
		name: 'createTime',
		align: 'left',
		label: 'createTime',
		field: 'createTime',
		sortable: false
	},
	// {
	//   name: 'updateTime',
	//   align: 'left',
	//   label: 'updateTime',
	//   field: 'updateTime',
	//   sortable: false,
	// },

	{ name: 'action', align: 'right', label: 'Action', sortable: false }
];

function bindContainer(container: Container) {
	$q.dialog({
		component: ChooseContainer,
		persistent: true,
		componentProps: {}
	})
		.onOk(async (data) => {
			$q.loading.show();
			try {
				const res: any = {
					appId: props.app.id,
					podSelector: container.podSelector,
					containerName: container.containerName,
					devContainerName: data.devContainerName,
					devEnv: data.devEnv
				};

				if (data.container) {
					res.containerId = data.container;
				}

				await store.bindContainer(res);
				await store.getMyContainers();
				containers.value = await store.getAppContainer(props.app.appName);
				store.getMyContainers();
			} catch (e: any) {
				$q.notify({
					type: 'negative',
					message: 'update app cfg failed: ' + e.message
				});
			} finally {
				$q.loading.hide();
			}
		})
		.onCancel(() => {
			// console.log('>>>> Cancel')
		})
		.onDismiss(() => {
			// console.log('I am triggered on both OK and Cancel')
		});
}

async function unbindContainer(container: Container) {
	$q.loading.show();
	try {
		const res: any = {
			appId: props.app.id,
			podSelector: container.podSelector,
			containerName: container.containerName
		};

		res.containerId = container.id;

		await store.unbindContainer(res);
		await store.getMyContainers();
		containers.value = await store.getAppContainer(props.app.appName);
		store.getMyContainers();
	} catch (e: any) {
		$q.notify({
			type: 'negative',
			message: 'update app cfg failed: ' + e.message
		});
	} finally {
		$q.loading.hide();
	}
}
</script>

<!-- <style>
.my-table-details {
  font-size: 0.85em;
  font-style: italic;
  max-width: 200px;
  white-space: normal;
  color: #555;
  margin-top: 4px;
}
</style> -->

<style lang="scss" scoped>
.dev-container {
	width: 100%;
	padding: 44px 0;
}

.container {
	width: 100%;
	padding-top: 24px;
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
	grid-gap: 32px;
}

.nodata {
	width: 100%;
	height: calc(100vh - 200px);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
}
</style>
