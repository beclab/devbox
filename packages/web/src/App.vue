<template>
	<router-view />
</template>

<script>
import { defineComponent } from 'vue';
import { useDevelopingApps } from './stores/app';
import { supportLanguages } from './i18n';
import { i18n } from './boot/i18n';

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
		const appStore = useDevelopingApps();
		const host = window.location.origin;
		appStore.setUrl(host);

		let terminusLanguage = '';
		let terminusLanguageInfo = document.querySelector(
			'meta[name="terminus-language"]'
		);
		if (terminusLanguageInfo && terminusLanguageInfo.content) {
			terminusLanguage = terminusLanguageInfo.content;
		} else {
			terminusLanguage = navigator.language;
		}

		console.log(navigator.language);

		if (terminusLanguage) {
			if (supportLanguages.find((e) => e.value == terminusLanguage)) {
				i18n.global.locale.value = terminusLanguage;
			}
		}

		return {};
	}
});
</script>
