import { boot } from 'quasar/wrappers';
import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor';

export default boot(async ({ app }) => {
	app.use(VueMonacoEditorPlugin, {
		paths: {
			// The recommended CDN config
			vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.38.0/min/vs'
		}
	});
});
