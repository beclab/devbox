<template>
	<q-page class="row items-center justify-evenly">
		<div style="width: 100%; line-height: 80px">
			<div>
				<q-btn-toggle
					v-model="lang"
					toggle-color="primary"
					flat
					:options="[
						{ label: 'EN', value: '' },
						{ label: '中文', value: '_cn' }
					]"
				/>
			</div>
			<div class="markdown-body" style="padding-left: 20px">
				<vue-markdown :source="src" />
			</div>
		</div>
	</q-page>
</template>

<script lang="ts" setup>
import { ref, onErrorCaptured, onMounted, watch } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';

const src = ref('');
const $q = useQuasar();
const lang = ref('_cn');

// onErrorCaptured((err) => {
//   errNotify(err.message);
//   return false;
// });

// const errNotify = (msg: string) => {
//   $q.notify({
//     color: 'red-4',
//     textColor: 'white',
//     icon: 'error',
//     message: msg,
//   });
// };

const load = async (lang: string) => {
	const res = await axios.get('/docs/help' + lang + '.md');

	src.value = res.data;
};

onMounted(async () => {
	await load(lang.value);
});

watch(
	() => lang.value,
	async () => {
		await load(lang.value);
	}
);
</script>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.3.0/github-markdown-light.min.css');
</style>
<style scoped>
main {
	align-items: normal;
}
</style>
