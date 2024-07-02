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
			<span class="q-mt-xl">{{ t('no_data') }}</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, PropType } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../stores/app';
import { ApplicationInfo, Container } from '@devbox/core';
import { useI18n } from 'vue-i18n';
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
const { t } = useI18n();

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
