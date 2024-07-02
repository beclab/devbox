<template>
	<div class="upload-file">
		<q-uploader
			ref="IconUploader"
			flat
			auto-upload
			hide-upload-btn
			:url="uploadUrl"
			:max-file-size="maxSize"
			:accept="accept"
			field-name="upload"
			@uploaded="uploaded"
			@uploading="uploading"
			:filter="checkFileNum"
			@rejected="onRejected"
			style="width: 100%; height: 276px"
		>
			<template v-slot:header="scope">
				<div
					class="upload-default column items-center justify-center"
					v-if="scope.files && scope.files.length <= 0 && !showDefaultImg"
				>
					<img src="../../assets/default-img.png" />
					<div class="tips text-ink-1 text-subtitle">
						{{ message || t('upload_icon_nofi') }}
					</div>
					<div class="upload-btn text-teal-8">
						<q-uploader-add-trigger />
						<q-icon name="sym_r_upgrade" size="16px" />
						<span class="text-body3">{{ t('upload') }}</span>
					</div>
				</div>

				<div
					class="replace-btn text-teal-8 cursor-pointer"
					@click="scope.removeQueuedFiles"
					v-else
				>
					<q-uploader-add-trigger />
					<q-icon name="sym_r_refresh" size="16px" />
					<span class="text-body3">{{ t('replace') }}</span>
				</div>
			</template>

			<template v-slot:list="scope">
				<div class="upload-list" v-if="!showDefaultImg">
					<template v-for="(file, index) in scope.files" :key="file.__key">
						<div class="item-img" v-if="index === scope.files.length - 1">
							<img :src="file.__img.src" />

							<div
								class="img-delete cursor-pointer"
								@click="scope.removeFile(file)"
							>
								<q-icon name="sym_r_delete" size="16px" color="teal-6" />
							</div>

							<div class="uploading" v-if="scope.isUploading">
								<img src="../../assets/uploading.svg" />
							</div>
						</div>
					</template>
				</div>
				<div class="upload-list" v-else>
					<div class="item-img">
						<img :src="defaultImg" />

						<div class="img-delete cursor-pointer" @click="deleteDefaultImg">
							<q-icon name="sym_r_delete" size="16px" color="teal-6" />
						</div>

						<div class="uploading" v-if="scope.isUploading">
							<img src="../../assets/uploading.svg" />
						</div>
					</div>
				</div>
			</template>
		</q-uploader>
	</div>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { defineProps, defineEmits, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	defaultImg: {
		type: String,
		required: false
	},
	maxSize: {
		type: [Number, String],
		required: false
	},
	accept: {
		type: String,
		required: false
	},
	acceptW: {
		type: [Number, String],
		required: false
	},
	acceptH: {
		type: [Number, String],
		required: false
	},
	message: {
		type: String,
		required: false
	}
});

const emit = defineEmits(['uploaded']);
const uploadUrl = ref(`${process.env.UPLOAD}/upload`);

const { t } = useI18n();
const $q = useQuasar();
const IconUploader = ref();
const showDefaultImg = ref(true);

if (props.defaultImg) {
	showDefaultImg.value = true;
} else {
	showDefaultImg.value = false;
}

function checkFileNum(files: any[]) {
	console.log('files', files);
	return files.filter(async (file) => {
		return validSize(file);
	});
}

const validSize = (file: Blob | MediaSource) => {
	new Promise<void>(function (resolve, reject) {
		let width = Number(props.acceptW) * 1;
		let height = Number(props.acceptH) * 1;
		let _URL = window.URL || window.webkitURL;
		let img = new Image();
		img.onload = function () {
			let valid = img.width === width && img.height === height;

			console.log('validvalid', valid);
			valid ? resolve() : reject();
		};
		img.src = _URL.createObjectURL(file);
	}).then(
		() => {
			return true;
		},
		() => {
			console.log('IconUploaderIconUploader', IconUploader.value);
			IconUploader.value.abort();
			setTimeout(() => {
				IconUploader.value.reset();
			}, 500);
			const message = `The uploaded image size can only be ${props.acceptW}px * ${props.acceptH}px! Please select again!`;

			$q.notify({
				message: message,
				type: 'warning'
			});
			return false;
		}
	);
};

const onRejected = (rejectedEntries) => {
	$q.notify({
		type: 'warning',
		message: `${rejectedEntries.length} file(s) did not pass validation constraints`
	});
};

function uploaded(info: any) {
	const xhr = JSON.parse(info.xhr.response);
	if (xhr.code === 0) {
		emit('uploaded', xhr.url);
	}
}

const uploading = (info: any) => {
	showDefaultImg.value = false;
	console.log('uploadinguploading', info);
};

const deleteDefaultImg = () => {
	showDefaultImg.value = false;
};
</script>

<style lang="scss" scoped>
.upload-file {
	width: 100%;
	height: 276px;
	border-radius: 12px;
	overflow: hidden;
	border: 1px solid $separator;
	position: relative;

	.upload-default {
		height: 276px;
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		margin: auto;
		img {
			width: 48px;
			height: 48px;
		}
		.tips {
			margin-top: 12px;
			margin-bottom: 20px;
		}
		.upload-btn {
			width: 72px;
			height: 24px;
			border-radius: 4px;
			border: 1px solid var(--Teal-08, #00967d);
			display: flex;
			align-items: center;
			justify-content: center;
			z-index: 1;
		}
	}

	.replace-btn {
		position: absolute;
		top: 230px;
		left: 20px;
		width: 72px;
		height: 24px;
		border-radius: 4px;
		border: 1px solid var(--Teal-08, #00967d);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1;
	}

	.upload-list {
		padding: 20px;
		display: flex;
		align-items: center;
		justify-content: center;

		.item-img {
			width: 160px;
			height: 160px;
			display: flex;
			align-items: center;
			justify-content: center;
			position: relative;
			img {
				width: 100%;
				height: 100%;
			}
			.img-delete {
				width: 24px;
				height: 24px;
				border-radius: 12px;
				background: #ffffff;
				box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
				position: absolute;
				right: -12px;
				top: -12px;
				z-index: 1;
				display: flex;
				align-items: center;
				justify-content: center;
			}

			.uploading {
				width: 100%;
				height: 100%;
				position: absolute;
				top: 0;
				left: 0;
				background: rgba(0, 0, 0, 0.6);
				display: flex;
				align-items: center;
				justify-content: center;
				img {
					width: 40px;
					height: 40px;
					animation: rotateAnimation 1.5s linear infinite;
				}
			}
		}
	}
}

@keyframes rotateAnimation {
	0% {
		transform: rotate(360deg);
	}
	100% {
		transform: rotate(0deg);
	}
}
</style>
