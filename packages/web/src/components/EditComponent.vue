<template>
	<div class="files row">
		<div class="files-left">
			<app-tree
				:app="app"
				:selected="selected"
				:chartNodes="chartNodes"
				@on-selected="onSelected"
				@update-path-node="updatePathNode"
			/>
		</div>

		<div class="files-right col-9">
			<app-edit-file
				:fileInfo="fileInfo"
				:isEditing="isEditing"
				@on-save-file="onSaveFile"
				@change-code="changeCode"
				@editorMount="editorMount"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, PropType, reactive } from 'vue';
import axios from 'axios';
import { useRouter, useRoute } from 'vue-router';
import { useDevelopingApps } from '../stores/app';
import { ApplicationInfo } from '@devbox/core';
import { FilesSelectType, FilesCodeType } from '../types/types';
import { BtDialog, BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { Encoder } from '@bytetrade/core';

import AppTree from './common/AppTree.vue';
import AppEditFile from './common/AppEditFile.vue';

const store = useDevelopingApps();
const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	}
});

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const chartNodes = ref<any>([]);
const tempFile = ref();
const isEditing = ref(false);
const selected = ref(props.app.chart);

const fileInfo = reactive<FilesCodeType>({
	code: '',
	lang: 'json',
	name: ''
});

onMounted(async () => {
	if (route.params.path) {
		selected.value = Encoder.bytesToString(
			Encoder.base64UrlToBytes(route.params.path)
		);
	}

	await loadChart();
});

window.onbeforeunload = function (e) {
	if (isEditing.value) {
		var ev = window.event || e;
		ev.returnValue = `${fileInfo.name} has been modified. Do you want to save the changes and update the chart repository?'`;
	}
};

watch(
	() => fileInfo.code,
	(newVal) => {
		if (newVal !== tempFile.value) {
			isEditing.value = true;
		} else {
			isEditing.value = false;
		}
	}
);

const changeCode = (value) => {
	fileInfo.code = value;
};

async function onSaveFile() {
	if (!selected.value) return false;
	try {
		await axios.put(store.url + '/api/files/' + selected.value, fileInfo.code, {
			headers: { 'content-type': 'text/plain' }
		});

		isEditing.value = false;
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.save_file_success')
		});
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.save_file_failed') + e.message
		});
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
			expandable: data.isDir,
			selectable: !data.isDir,
			children: data.isDir ? [{}] : null,
			isDir: data.isDir,
			lazy: data.isDir ? true : false
		};
		children.push(selectData);
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
				isDir: true
			}
		];
	} catch (e: any) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.save_loadChart_failed') + e.message
		});
	}
}

const updatePathNode = async (path: string) => {
	try {
		const res: any = await axios.get(store.url + '/api/files/' + path);

		const children = getChildren(res.items);

		const replaceNodes = replaceObjectByPath(chartNodes.value, path, children);

		chartNodes.value = replaceNodes;
	} catch (e: any) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.save_loadChart_failed') + e.message
		});
	}
};

function replaceObjectByPath(array, targetPath, newObject) {
	return array.map((item) => {
		if (item.path === targetPath) {
			item.children = newObject;
			return item;
		}

		if (item.children && Array.isArray(item.children)) {
			return {
				...item,
				children: replaceObjectByPath(item.children, targetPath, newObject)
			};
		}

		return item;
	});
}

const onSelected = async (value) => {
	if (isEditing.value) {
		checkFileSave(value);
	} else {
		selected.value = value;
		fetchData(value);
	}
};

const fetchData = async (value) => {
	try {
		const res: any = await axios.get(store.url + '/api/files/' + value, {});

		isEditing.value = false;
		tempFile.value = res.content ? res.content : '';
		fileInfo.code = res.content ? res.content : '';
		// fileInfo.lang = res.extension;
		fileInfo.name = res.name;

		router.push({
			path: `/app/${props.app.id}/${Encoder.stringToBase64Url(value)}`
		});
	} catch (e: any) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.save_loadChart_failed') + e.message
		});
	}
};

const checkFileSave = (value) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: t('message.confirmation'),
		message: t('message.save_file')
	})
		.then(async (val) => {
			if (val) {
				await onSaveFile();
				selected.value = value;
				await fetchData(value);
			}
		})
		.catch((err) => {
			console.log(err);
		});
};

const editorMount = () => {
	if (selected.value === props.app.chart) {
		if (chartNodes.value && chartNodes.value.length > 0) {
			const defaultFile = chartNodes.value[0].children.find(
				(item) => !item.isDir
			);

			console.log('defaultFile', defaultFile);

			selected.value = defaultFile.path;
			fetchData(defaultFile.path);
		}
	} else {
		fetchData(selected.value);
	}
};
</script>

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
	}
}
</style>
