<template>
	<router-view />
</template>

<script>
import { defineComponent } from 'vue';
import { useRouter } from 'vue-router';
import { useDevelopingApps } from './stores/app';

export default defineComponent({
	name: 'App',
	preFetch() {
		const appStore = useDevelopingApps();
		const host = window.location.origin;
		appStore.setUrl(host);
		return new Promise((resolve) => {
			appStore.getApps().then(() => {
				appStore.getMyContainers().then(() => {
					resolve({});
				});
			});
		});
	},
	setup() {
		const router = useRouter();
		router.push({ path: '/home' });
		const appStore = useDevelopingApps();
		const host = window.location.origin;
		appStore.setUrl(host);

		return {};
	}
});
</script>
