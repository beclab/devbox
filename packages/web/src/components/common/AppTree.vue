<template>
	<q-tree
		v-if="chartNodes && chartNodes.length > 0 && chartNodes[0].children"
		ref="treeRef"
		dense
		:nodes="chartNodes"
		node-key="path"
		label-key="label"
		selected-color="blue-default"
		:selected="selected"
		v-model:expanded="expanded"
		@update:selected="onSelected"
		@update:expanded="onExpanded"
		no-selection-unset
		no-results-label=" "
		no-nodes-label=" "
		text-color="ink-2"
		class="my-tree-wrapper"
	>
		<template v-slot:default-header="prop">
			<div
				:id="prop.node.key"
				class="my-tree-header-container"
				@mouseenter="mouseenter(prop.node.path)"
				@mouseleave="mouseleave(prop.node.path)"
			>
				<div
					class="row no-wrap items-center justify-between my-tree-header-wrapper"
					:id="prop.node.key"
				>
					<div class="row items-center justify-start">
						<img
							v-if="!prop.node.isDir"
							class="q-mx-xs"
							src="../../assets/icon-txt.svg"
							style="width: 12px"
						/>

						<span
							style="
								white-space: nowrap;
								overflow: hidden;
								text-overflow: ellipsis;
							"
							>{{ prop.node.label }}</span
						>
					</div>

					<q-icon
						class="horiz"
						:class="{ 'hide-horiz': mouseItemKey === prop.node.path }"
						rounded
						clickable
						name="sym_r_more_horiz"
						size="18px"
						@click.stop
					>
						<PopupMenu
							:items="
								prop.node.isDir ? [...fileMenu, ...operateMenu] : operateMenu
							"
							:path="prop.node.path"
							:label="prop.node.label"
							@handleEvent="handleEvent"
						/>
					</q-icon>
				</div>
			</div>
		</template>
	</q-tree>
</template>

<script lang="ts" setup>
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import axios from 'axios';
import { ApplicationInfo } from '@devbox/core';
import { BtDialog, BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { OPERATE_ACTION } from '../../types/constants';
import { FilesSelectType } from '../../types/types';
import { useDevelopingApps } from '../../stores/app';

import PopupMenu from '../common/PopupMenu.vue';

const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	},
	selected: {
		type: String,
		required: true,
		default: ''
	},
	chartNodes: {
		type: Object as PropType<FilesSelectType[]>,
		required: true
	}
});

console.log('app', props.app);

const emits = defineEmits(['onSelected', 'loadChart']);

const { t } = useI18n();
const store = useDevelopingApps();

const expanded = ref<string[]>([props.app.appName]);
const mouseItemKey = ref();

const fileMenu = ref([
	{
		label: t(`enums.${OPERATE_ACTION.ADD_FOLDER}`),
		name: OPERATE_ACTION.ADD_FOLDER,
		icon: 'sym_r_create_new_folder'
	},
	{
		label: t(`enums.${OPERATE_ACTION.ADD_FILE}`),
		name: OPERATE_ACTION.ADD_FILE,
		icon: 'sym_r_note_add'
	}
]);

const operateMenu = ref([
	{
		label: t(`enums.${OPERATE_ACTION.RENAME}`),
		name: OPERATE_ACTION.RENAME,
		icon: 'sym_r_edit_square'
	},
	{
		label: t(`enums.${OPERATE_ACTION.DELETE}`),
		name: OPERATE_ACTION.DELETE,
		icon: 'sym_r_delete'
	}
]);

const onSelected = async (value) => {
	emits('onSelected', value);
};

const onExpanded = (value) => {
	expanded.value = value;
};

const mouseenter = (key: string) => {
	mouseItemKey.value = key;
};

const mouseleave = (key: string) => {
	mouseItemKey.value = null;
};

const handleEvent = (action: OPERATE_ACTION, path: string, label: string) => {
	switch (action) {
		case OPERATE_ACTION.ADD_FOLDER:
			createDialog(path, action);
			break;

		case OPERATE_ACTION.ADD_FILE:
			createDialog(path, action);
			break;

		case OPERATE_ACTION.RENAME:
			renameDialog(path, label, action);
			break;

		case OPERATE_ACTION.DELETE:
			deleteFile(path);
			break;
	}
};

const createDialog = (path: string, action: OPERATE_ACTION) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title:
			action === OPERATE_ACTION.ADD_FILE
				? t('dialog_create_file')
				: t('dialog_create_folder'),
		prompt: {
			isValid: (val) => val.length > 2,
			type: 'text',
			name: t('dialog_create_title'),
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
	try {
		await axios.put(store.url + '/api/files/' + path);
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.create_file_success')
		});
		await emits('loadChart');
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.create_file_failed') + e.message
		});
	}
};

const createFolder = async (path: string) => {
	try {
		const res = await axios.post(
			store.url + '/api/files/' + path + '?file_type=dir'
		);
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.create_folder_success')
		});
		await emits('loadChart');
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.create_folder_failed') + e.message
		});
	}
};

const renameDialog = (path: string, label: string, action: OPERATE_ACTION) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: t('message.rename_file'),
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
				renameFile(path, label, val);
			}
		})
		.catch((err) => {
			console.log(err);
		});
};

const renameFile = async (path: string, label: string, newname: any) => {
	const newPath = path.replace(label, newname);

	try {
		await axios.patch(
			store.url +
				'/api/files/' +
				path +
				'?action=rename&destination=' +
				newPath,
			{},
			{
				headers: { 'content-type': 'text/plain' }
			}
		);
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.rename_folder_success')
		});
		await emits('loadChart');
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.rename_folder_failed')
		});
	}
};

const deleteFile = async (path: string) => {
	BtDialog.show({
		platform: 'web',
		cancel: true,
		message: t('message.deleteTip'),
		okStyle: {
			background: '#00BE9E',
			color: '#ffffff'
		},
		title: 'Delete'
	})
		.then((val) => {
			if (val) {
				_deleteFile(path);
			}
		})
		.catch((err) => {
			console.log(err);
		});
};

const _deleteFile = async (path: string) => {
	try {
		await axios.delete(store.url + '/api/files/' + path);
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.delete_file_success')
		});
		await emits('loadChart');
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('message.delete_file_failed') + e.message
		});
	}
};
</script>

<style lang="scss" scoped>
.horiz {
	opacity: 0;
}
.hide-horiz {
	opacity: 1;
}
.my-tree-wrapper {
	width: calc(100% - 10px);
	::v-deep(.q-tree__arrow) {
		font-size: 14px;
		border: 1px solid rgba(255, 255, 255, 0);
		color: $ink-1;
	}
	::v-deep(.q-tree__spinner) {
		font-size: 14px;
		border: 1px solid rgba(255, 255, 255, 0);
		color: $ink-1;
	}
}
::v-deep(.selected-node) {
	position: relative;
	&::after {
		content: '';
		position: absolute;
		left: -4px;
		right: 4px;
		top: 0;
		bottom: 0;
		background-color: rgba(34, 111, 255, 0.12);
		border-radius: 4px;
	}
}
.my-tree-header-container {
	width: 100%;
	position: relative;
	border-radius: 4px;
	.my-tree-badge-wrapper {
		margin-left: 4px;
	}
	.my-tree-avatar {
		position: relative;
		& ~ .my-tree-badge-wrapper {
			position: absolute;
			bottom: 0;
			right: 0;
		}
	}
}
::v-deep(.q-tree .q-tree__node-header) {
	border-radius: 4px;
}
::v-deep(.q-tree .q-tree__node-header .q-focus-helper) {
	left: -6px;
	right: 0;
	width: auto;
	border-radius: 4px;
}

::v-deep(.q-focus-helper:has(+ .q-icon.q-tree__arrow)) {
	left: 0;
	right: 0;
	width: auto;
}

.my-tree-header-wrapper {
	padding: 4px 0;
	border-radius: 4px;
	padding-right: 8px;
	width: 100%;
}

::v-deep(.q-tree__node-header:before) {
	border-bottom-left-radius: 7px;
	border-color: $separator-2;
}
::v-deep(
		.q-tree__node--parent > .q-tree__node-collapsible > .q-tree__node-body:after
	) {
	border-color: #e0e0e0;
}
::v-deep(.q-tree__node:after) {
	border-color: $separator-2;
}
::v-deep(.my-menu-before-scroll .q-scrollarea__container) {
	overflow-x: hidden;
}
::v-deep(.my-menu-before-scroll .q-scrollarea__content) {
	min-width: initial;
	width: 100%;
}
::v-deep(.my-menu-before-scroll .q-tree__node-collapsible) {
	overflow-x: hidden;
}
::v-deep(
		.q-tree
			> .q-tree__node--parent
			> .q-tree__node-header
			> .q-tree__node-header-content
	) {
	color: $ink-1 !important;
}
::v-deep(
		.q-tree--dense
			.q-tree__node--parent
			> .q-tree__node-collapsible
			> .q-tree__node-body
	) {
	padding-bottom: 0px;
}
</style>
