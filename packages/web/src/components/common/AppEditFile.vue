<template>
	<div class="files-right-header row items-center justify-between">
		<div class="row items-center justify-start">
			<img
				class="q-mr-sm"
				src="../../assets/icon-txt.svg"
				style="width: 12px"
			/>
			<span>{{ fileInfo.name }}</span>
			<span
				class="statusIcon q-ml-sm"
				:style="{
					background: isEditing ? '#FFC46D' : 'rgba(41, 204, 95, 1)'
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
			@change="changeCode"
			:value="fileInfo.code"
			@mount="editorMount"
		/>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { FilesCodeType } from '../../types/types';

defineProps({
	isEditing: {
		type: Boolean,
		default: false,
		required: false
	},
	fileInfo: {
		type: Object as PropType<FilesCodeType>,
		required: true
	}
});

const emits = defineEmits(['onSaveFile', 'changeCode', 'editorMount']);

const onSaveFile = () => {
	emits('onSaveFile');
};

const changeCode = (value) => {
	console.log('changeCode', value);
	emits('changeCode', value);
};

const editorMount = (value) => {
	console.log('editorMount', value);
	emits('editorMount');
};
</script>

<style lang="scss" scoped>
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
</style>

<style lang="scss" scoped>
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
