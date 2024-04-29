<template>
	<div class="q-pa-md">
		<div class="row" style="margin-bottom: 15px"></div>
		<div class="row">
			<div class="col-12 col-md-9">
				<vue-monaco-editor
					height="80vh"
					theme="vs-light"
					:language="lang"
					v-model:value="code"
				/>
			</div>
			<div class="col-12 col-md-3">
				<q-tree
					:nodes="chartNodes"
					node-key="path"
					accordion
					v-model:expanded="expanded"
					v-model:selected="selectedKey"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, onErrorCaptured, watch, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
import { server } from '../env/vars';
import { useRoute } from 'vue-router';
import { useDevelopingApps } from '../stores/app';

export default defineComponent({
	setup() {
		const code = ref('');
		const lang = ref('json');
		const $q = useQuasar();
		const Route = useRoute();
		const apps = useDevelopingApps();
		const chartNodes = ref([]);
		const expanded = ref([]);
		const selectedKey = ref(null);

		onErrorCaptured((err) => {
			errNotify(err.message);
			return false;
		});

		onMounted(async () => {
			if (apps.apps?.length <= 0) {
				await apps.getApps();
			}
			const appName = await loadChart(Route.query.app);
			expanded.value = [appName];
		});

		watch(
			() => selectedKey.value,
			() => {
				onSelected(selectedKey.value);
			}
		);

		const errNotify = (msg: string) => {
			$q.notify({
				color: 'red-4',
				textColor: 'white',
				icon: 'error',
				message: msg
			});
		};

		const successNotify = (msg: string) => {
			$q.notify({
				color: 'green-4',
				textColor: 'white',
				icon: 'done',
				message: msg
			});
		};

		const alertErr = (e) => {
			let msg = '';
			if (axios.isAxiosError(e)) {
				if (e.response && e.response?.data != '') {
					msg = e.response?.data;
					console.log(msg);
				} else {
					msg = e.message;
				}
			}

			if (msg == undefined || msg == '') {
				msg = 'unknown error';
			}

			errNotify(msg);
		}; // end alertErr

		const getChildren = (items) => {
			let children = [];

			for (let n in items) {
				const data = items[n];
				children.push({
					label: data.name,
					icon: data.isDir ? 'folder' : 'article',
					path: data.path,
					expandable: data.isDir,
					selectable: !data.isDir,
					children: data.isDir ? [{}] : null,
					handler: data.isDir ? loadChildren : null,
					isDir: data.isDir
				});
			}

			return children;
		}; // end getChildren

		const loadChart = async (name: string | null) => {
			for (let i in apps.apps) {
				const app = apps.apps[i];
				if (app.appName == name) {
					try {
						let res = await axios.get(server + '/api/files' + app.chart, {});
						console.log(res);
						if (res.data.code != 200) {
							errNotify(res.data.msg);
							return;
						}

						const children = getChildren(res.data.data.items);
						chartNodes.value = [
							{
								label: app.appName,
								icon: 'folder',
								children: children,
								selectable: false,
								path: app.appName,
								isDir: true
							}
						];
					} catch (e) {
						alertErr(e);
					}

					// app found
					return app.appName;
				} // end if app name
			} // end for

			errNotify('app not found');
		}; // end loadChart

		const onSelected = async (path: string) => {
			try {
				const res = await axios.get(server + '/api/files/' + path, {});
				if (res.data.code != 200) {
					errNotify(res.data.msg);
					return;
				}

				code.value = res.data.data.content ? res.data.data.content : '';
				lang.value = res.data.data.extension;
			} catch (e) {
				alertErr(e);
			}
		}; // end onSelected

		const loadChildren = async (node) => {
			try {
				const res = await axios.get(server + '/api/files/' + node.path);
				// const res = {
				//   data: {
				//     code: 200,
				//     data: {
				//       items: [
				//         {
				//           path: 'desktop-dev/Chart.yaml',
				//           name: 'Chart.yaml',
				//           size: 1149,
				//           extension: '.yaml',
				//           modified: '2023-12-26T06:37:24.044671435Z',
				//           mode: 420,
				//           isDir: false,
				//           isSymlink: false,
				//           type: 'text',
				//         },
				//         {
				//           path: 'desktop-dev/app.cfg',
				//           name: 'app.cfg',
				//           size: 617,
				//           extension: '.cfg',
				//           modified: '2023-12-26T06:37:24.040671267Z',
				//           mode: 420,
				//           isDir: false,
				//           isSymlink: false,
				//           type: 'text',
				//         },
				//         {
				//           path: 'desktop-dev/templates',
				//           name: 'templates',
				//           size: 4096,
				//           extension: '',
				//           modified: '2023-12-26T06:37:24.036337513Z',
				//           mode: 2147484141,
				//           isDir: true,
				//           isSymlink: false,
				//           type: '',
				//         },
				//         {
				//           path: 'desktop-dev/values.yaml',
				//           name: 'values.yaml',
				//           size: 0,
				//           extension: '.yaml',
				//           modified: '2023-12-26T06:37:24.038595925Z',
				//           mode: 420,
				//           isDir: false,
				//           isSymlink: false,
				//           type: 'text',
				//         },
				//       ],
				//       numDirs: 1,
				//       numFiles: 3,
				//       sorting: {
				//         by: '',
				//         asc: false,
				//       },
				//       path: 'desktop-dev',
				//       name: 'desktop-dev',
				//       size: 4096,
				//       extension: '',
				//       modified: '2023-12-26T06:37:24.036337513Z',
				//       mode: 2147484141,
				//       isDir: true,
				//       isSymlink: false,
				//       type: '',
				//     },
				//   },
				// };

				if (res.data.code != 200) {
					errNotify(res.data.msg);
					return;
				}

				const setChildren = (n, path, children) => {
					for (let i in n) {
						if (n[i].path == path && n[i].isDir) {
							n[i].children = children;
							return;
						}

						if (n[i].isDir && n[i].children.length > 0) {
							setChildren(n[i].children, path, children);
						}
					}
				}; // end setChildren

				const children = getChildren(res.data.data.items);
				let nodes = chartNodes.value;
				setChildren(nodes, node.path, children);
				chartNodes.value = nodes;
			} catch (e) {
				alertErr(e);
			}
		}; // end loadChildren

		return {
			code,
			lang,
			chartNodes,
			expanded,
			selectedKey
		};
	}
});
</script>
