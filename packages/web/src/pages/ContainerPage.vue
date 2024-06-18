<template>
	<div class="dev-container">
		<div class="text-h3 text-ink-1">Dev Container List</div>
		<div
			class="container"
			v-if="store.containers && store.containers.length > 0"
		>
			<container-card
				v-for="(container, index) in store.containers"
				:key="index"
				:container="container"
			/>
		</div>
		<div class="nodata" v-else>
			<img src="../assets/nodata.svg" />
			<span class="q-mt-xl">No data.</span>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { onMounted } from 'vue';
import { useDevelopingApps } from '../stores/app';
import ContainerCard from '../components/common/ContainerCard.vue';

const store = useDevelopingApps();

onMounted(() => {
	store.getMyContainers();
});
</script>

<style lang="scss" scoped>
.dev-container {
	width: 100%;
	padding: 44px;
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
