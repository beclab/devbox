<template>
	<div class="files row">
		<div class="files-left">
			<bt-menu
				v-if="chartNodes && chartNodes.length > 0 && chartNodes[0].children"
				active-class="my-code-link"
				:defaultClose="true"
				:hideExpandIcon="true"
				:items="chartNodes"
				v-model="selectedKey"
				@select="onSelected"
				@toggleVaule="toggleVaule"
			>
				<template
					v-for="node in chartNodes"
					:key="node.key"
					v-slot:[`extra-${node.key}`]
				>
					<q-icon size="xs" name="sym_r_add_circle" color="ink-2">
						<PopupMenu
							:items="fileMenu"
							:path="node.path"
							:label="node.label"
							@handleEvent="handleEvent"
						/>
					</q-icon>
				</template>

				<template
					v-for="nod in menuSlotkeys"
					:key="nod"
					v-slot:[`extra-${nod.key}`]
				>
					<div>
						<q-icon
							class="q-mr-xs"
							size="18px"
							name="sym_r_add_circle"
							v-if="nod.isDir"
						>
							<PopupMenu
								:items="fileMenu"
								:path="nod.path"
								:label="nod.label"
								@handleEvent="handleEvent"
							/>
						</q-icon>

						<q-icon rounded clickable name="sym_r_more_horiz" size="18px">
							<PopupMenu
								:items="oprateMenu"
								:path="nod.path"
								:label="nod.label"
								@handleEvent="handleEvent"
							/>
						</q-icon>
					</div>
				</template>
			</bt-menu>
		</div>

		<div class="files-right col-9">
			<div class="files-right-header row items-center justify-between">
				<div class="row items-center justify-start">
					<img
						class="q-mr-sm"
						src="../assets/icon-txt.svg"
						style="width: 12px"
					/>
					<span>{{ fileInfo.name }}</span>
					<span
						class="statusIcon q-ml-sm"
						:style="{
							background: fileStatus ? '#FFC46D' : 'rgba(41, 204, 95, 1)'
						}"
					></span>
				</div>
				<div>
					<q-icon
						class="q-ml-md cursor-pointer"
						name="sym_r_save"
						size="20px"
						@click="onSaveFile"
					/>
				</div>
			</div>
			<div class="files-right-content">
				<vue-monaco-editor
					class="files-monaco"
					theme="vs-light"
					:language="fileInfo.lang"
					v-model:value="fileInfo.code"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, PropType, reactive } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
import { useDevelopingApps } from '../stores/app';
import { ApplicationInfo } from '@devbox/core';
import { OPERATE_ACTION } from '../types/constants';
import { FilesSelectType } from '../types/types';
import { BtDialog } from '@bytetrade/ui';

import PopupMenu from './common/PopupMenu.vue';

const menuSlotkeys = ref<FilesSelectType[]>([]);
const store = useDevelopingApps();
const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	}
});

const $q = useQuasar();
const appName = ref(props.app.appName);
const chartNodes = ref<any>([]);
const selectedKey = ref(null);
const tempFile = ref();
const fileStatus = ref(false);

const fileInfo = reactive({
	code: '',
	lang: 'json',
	name: ''
});

const fileMenu = ref([
	{
		name: OPERATE_ACTION.ADD_FOLDER,
		icon: 'sym_r_create_new_folder'
	},
	{
		name: OPERATE_ACTION.ADD_FILE,
		icon: 'sym_r_note_add'
	}
]);

const oprateMenu = ref([
	{
		name: OPERATE_ACTION.RENAME,
		icon: 'sym_r_edit_square'
	},
	{
		name: OPERATE_ACTION.DELETE,
		icon: 'sym_r_delete'
	}
]);

onMounted(async () => {
	await loadChart();
});

window.onbeforeunload = function (e) {
	if (fileStatus.value) {
		var ev = window.event || e;
		ev.returnValue = `${fileInfo.name} has been modified. Do you want to save the changes and update the chart repository?'`;
	}
};

watch(
	() => fileInfo.code,
	(newVal) => {
		if (newVal !== tempFile.value) {
			fileStatus.value = true;
		} else {
			fileStatus.value = false;
		}
	}
);

async function onSaveFile() {
	if (selectedKey.value != null) {
		const res: any = await axios.put(
			store.url + '/api/files/' + selectedKey.value,
			fileInfo.code,
			{ headers: { 'content-type': 'text/plain' } }
		);
		if (res.code != 200) {
			return;
		}
		fileStatus.value = false;
		$q.notify('success to save file');
	}
}

const getChildren = (items: any) => {
	let children: FilesSelectType[] = [];

	for (let n in items) {
		const data = items[n];
		const selectData: FilesSelectType = {
			label: data.name,
			icon: data.isDir ? 'folder' : 'article',
			path: data.path,
			key: data.path,
			expandable: data.isDir,
			selectable: !data.isDir,
			children: data.isDir ? [{}] : null,
			handler: data.isDir ? loadChildren : null,
			isDir: data.isDir,
			defaultHide: true
		};
		children.push(selectData);

		if (
			!menuSlotkeys.value.find(
				(key: { key: string }) => key.key === selectData.key
			)
		) {
			menuSlotkeys.value.push(selectData);
		}
	}

	return children;
};

async function loadChart() {
	try {
		const res: any = await axios.get(
			store.url + '/api/files' + props.app.chart
		);

		const children = getChildren(res.items);
		chartNodes.value = [
			{
				label: props.app.appName,
				icon: 'folder',
				children: children,
				selectable: false,
				path: props.app.appName,
				key: props.app.appName,
				isDir: true,
				defaultHide: true
			}
		];
	} catch (e: any) {
		$q.notify({
			type: 'negative',
			message: 'failed to loadChart; ' + e.message
		});
	}
}

const toggleVaule = (data) => {
	loadChildren(data.item);
};

const onSelected = async (value) => {
	selectedKey.value = value.item.path;
	try {
		const res: any = await axios.get(
			store.url + '/api/files/' + value.item.path,
			{}
		);

		fileStatus.value = false;
		tempFile.value = res.content ? res.content : '';
		fileInfo.code = res.content ? res.content : '';
		fileInfo.lang = res.extension;
		fileInfo.name = res.name;
	} catch (e: any) {
		$q.notify({
			type: 'negative',
			message: 'onSelect failed; ' + e.message
		});
	}
};

const loadChildren = async (node: any) => {
	try {
		const res: any = await axios.get(store.url + '/api/files/' + node.path);

		const setChildren = (n: any, path: any, children: any) => {
			for (let i in n) {
				if (n[i].path == path && n[i].isDir) {
					n[i].children = children;
					return;
				}

				if (n[i].isDir && n[i].children.length > 0) {
					setChildren(n[i].children, path, children);
				}
			}
		};

		const children = getChildren(res.items);
		let nodes = chartNodes.value;
		setChildren(nodes, node.path, children);

		chartNodes.value = nodes;
	} catch (e: any) {
		$q.notify({
			type: 'negative',
			message: 'loadChildren failed; ' + e.message
		});
	}
};

const handleEvent = (action: OPERATE_ACTION, path: string, label: string) => {
	switch (action) {
		case OPERATE_ACTION.ADD_FOLDER:
			createDialg(path, action);
			break;

		case OPERATE_ACTION.ADD_FILE:
			createDialg(path, action);
			break;

		case OPERATE_ACTION.RENAME:
			renameDialg(path, label, action);
			break;

		case OPERATE_ACTION.DELETE:
			deletefile(path);
			break;
	}
};

const createDialg = (path: string, action: OPERATE_ACTION) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: action === OPERATE_ACTION.ADD_FILE ? 'Create File' : 'Create Folder',
		prompt: {
			isValid: (val) => val.length > 2,
			type: 'text',
			name: 'Name',
			placeholder: ''
		}
	})
		.then((val) => {
			if (!val) return false;
			const filepath = `${path}/${val}`;
			if (action === OPERATE_ACTION.ADD_FOLDER) {
				createFolder(filepath);
			} else if (action === OPERATE_ACTION.ADD_FILE) {
				createFile(filepath);
			}
		})
		.catch((err) => {
			console.log(err);
		});
};

const createFile = async (path: string) => {
	const res = await axios.put(
		store.url + '/api/files/' + path,
		{},
		{
			headers: { 'content-type': 'text/plain' }
		}
	);
	$q.notify('success to create file');
	await loadChart();
};

const createFolder = async (path: string) => {
	const res = await axios.post(
		store.url + '/api/files/' + path + '?file_type=dir',
		{},
		{ headers: { 'content-type': 'text/plain' } }
	);
	$q.notify('success to create folder');
	await loadChart();
};

const renameDialg = (path: string, label: string, action: OPERATE_ACTION) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: 'Reanme',
		prompt: {
			model: label,
			isValid: (val) => val.length > 2,
			type: 'text',
			name: 'New Name',
			placeholder: ''
		}
	})
		.then((val) => {
			if (val) {
				renamefile(path, label, val);
			}
		})
		.catch((err) => {
			console.log(err);
		});
};

const renamefile = async (path: string, label: string, newname: any) => {
	const newpath = path.replace(label, newname);
	const res = await axios.patch(
		store.url + '/api/files/' + path + '?action=rename&destination=' + newpath,
		{},
		{
			headers: { 'content-type': 'text/plain' }
		}
	);
	$q.notify('success to rename');
	await loadChart();
};

const deletefile = async (path: string) => {
	const res = await axios.delete(store.url + '/api/files/' + path, {
		headers: { 'content-type': 'text/plain' }
	});
	$q.notify('success to create file');
	await loadChart();
};
</script>
<style lang="scss">
.my-code-link {
	background: $background-hover;
	color: $ink-1;
}
::-webkit-scrollbar {
	width: 0px !important;
	height: 0px !important;
}

::-webkit-scrollbar-thumb {
	border-radius: 10px;
	width: 1px;
	background: rgba(255, 255, 255, 0.5);
}

::-webkit-scrollbar-track {
	box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);
	border-radius: 10px;
	background: rgba(57, 177, 255, 0.16);
}
.monaco-editor .margin {
	background-color: $background-2 !important;
}

.lines-content.monaco-editor-background {
	background-color: $background-2 !important;
}

.minimap.slider-mouseover {
	background-color: $background-2 !important;
}
.minimap-decorations-layer {
	background-color: $background-2 !important;
}
.decorationsOverviewRuler {
	width: 0px !important;
}

.inputarea.monaco-mouse-cursor-text {
	background-color: $ink-1 !important;
	caret-color: red !important;
}
.monaco-editor .inputarea {
	background-color: $ink-1 !important;
	z-index: 1 !important;
	caret-color: red !important;
}

.view-lines .view-line {
	span {
		color: $ink-1 !important;
	}
	.mtk1 {
		color: $ink-1 !important;
	}
}
</style>
<style lang="scss" scoped>
.files {
	height: calc(100vh - 112px);
	margin-top: 32px;
	.files-left {
		width: 240px;
		background-color: $background-1;
	}
	.files-right {
		flex: 1;
		border-radius: 12px;
		border: 1px solid $separator;
		overflow: hidden;
		background: $background-3;
		.files-right-header {
			width: 100%;
			height: 32px;
			line-height: 32px;
			padding: 0 12px;
			border-bottom: 1px solid $separator;
			background: $background-3;
			.statusIcon {
				width: 6px;
				height: 6px;
				border-radius: 3px;
				display: inline-block;
			}
		}
		.files-right-content {
			height: calc(100% - 32px);
			padding: 12px;
			background: $background-3;

			.files-monaco {
				height: 100%;
				border-radius: 12px;
				overflow: hidden;
			}
		}
	}
}
</style>
