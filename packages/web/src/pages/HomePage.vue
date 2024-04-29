<template>
	<div class="wrap">
		<div class="container">
			<div class="header">
				<div class="text-h3 title">Welcome to DevBox</div>
				<div class="text-subtitle1 text-grey-6 q-mt-sm subtitle">
					An easy way to develop in Terminus
				</div>
			</div>
			<div class="row">
				<div class="col-6 q-pr-lg" style="padding-right: 40px">
					<div class="text-h5 text-grey-10">Start</div>

					<div
						class="text-subtitle1 text-white create-btn create-applation-btn row items-center justify-start"
						@click="onCreate"
					>
						<span class="add_wrap">
							<q-icon name="sym_r_add" size="20px" color="green" />
						</span>
						<span> Create a new application </span>
					</div>
					<input
						ref="uploadInput"
						type="file"
						style="display: none"
						accept=".tgz"
						@change="uploadFile"
					/>
					<div
						class="text-subtitle1 text-grey-6 create-btn create-con-btn row items-center justify-start"
						@click="uploadChat"
					>
						<span class="add_wrap">
							<q-icon name="sym_r_add" size="20px" color="grey-5" />
						</span>
						<span> Upload Chart </span>
					</div>

					<div class="text-h5 text-grey-10">Recent</div>

					<div class="recent-app column items-center justify-center">
						<img src="../assets/recent-icon.png" />
						<p class="text-grey-5 text-body1">
							The application you recently<br />
							developed will be displayed here.
						</p>
					</div>

					<!-- <div
            class="create-btn create-app-btn row items-center justify-between"
          >
            <div class="app-content row items-center justify-start">
              <img src="../assets/app-1.png" />
              <div>
                <div class="text-subtitle2 text-grey-10">Agora</div>
                <div class="text-body3 text-grey-5">
                  Last opened at 2023-03-02 12:11:23
                </div>
              </div>
            </div>

            <div class="app-status row items-center justify-center">
              <i></i>
              <span class="text-grey-8">Running</span>
            </div>
          </div> -->
				</div>
				<div class="col-6" style="padding-left: 40px">
					<div class="text-h5 text-grey-10">Documents</div>
					<document-link
						v-for="(item, index) in menuStore.documentList"
						:key="index"
						:data="item"
					></document-link>
					<div class="link-more text-body1 text-grey-5">
						Visit <span>Terminus developer community</span> for more documents
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../stores/app';
import { useMenuStore } from '../stores/menu';

import DocumentLink from '../components/common/DocumentLink.vue';

const $q = useQuasar();
const router = useRouter();
const store = useDevelopingApps();
const menuStore = useMenuStore();
const uploadInput = ref();

onMounted(async () => {
	//
});

function onCreate() {
	router.push({ path: '/create' });
}

const uploadChat = () => {
	uploadInput.value.click();
};

async function uploadFile(event: any) {
	console.log(event);
	const file = event.target.files[0];
	if (file) {
		const { status, message, appId } = await upload_dev_file(file);
		if (status) {
			$q.notify(message);
			if (appId) {
				await store.getApps();
				await router.push({ path: '/app/' + appId });
				await menuStore.updateApplications();
				menuStore.currentItem = '/app/' + appId;
			}
		} else {
			$q.notify(message);
		}
	} else {
		console.log('file selected failure');
	}
}

async function upload_dev_file(
	file: any
): Promise<{ status: boolean; message: string; appId?: number }> {
	try {
		const formData = new FormData();
		formData.append('chart', file);
		const res: { appId: number } = await axios.post(
			store.url + '/api/command/upload-app-archive',
			formData,
			{
				headers: { 'Content-Type': 'multipart/form-data' }
			}
		);

		return {
			status: true,
			message: 'upload chart success',
			appId: res.appId
		};
	} catch (e: any) {
		return { status: false, message: e.message };
	}
}
</script>

<style lang="scss" scoped>
.wrap {
	width: 100%;
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 44px;
	.container {
		width: 100%;
		max-width: 1280px;
		height: 100vh;
		.header {
			margin-top: 56px;
			margin-bottom: 20px;
			padding-bottom: 32px;
			border-bottom: 1px solid #ebebeb;
		}
		.create-btn {
			width: 100%;
			height: 72px;
			border-radius: 12px;
			padding: 0 20px;
			box-sizing: border-box;
			cursor: pointer;
			.add_wrap {
				width: 32px;
				height: 32px;
				line-height: 32px;
				display: inline-block;
				text-align: center;
				border-radius: 16px;
				margin-right: 12px;
			}

			.app-content {
				img {
					width: 32px;
					height: 32px;
					border-radius: 8px;
					margin-right: 12px;
				}
			}
			.app-status {
				i {
					width: 12px;
					height: 12px;
					display: inline-block;
					border-radius: 6px;
					background-color: green;
					margin-right: 8px;
				}
			}
		}
		.create-applation-btn {
			margin-top: 12px;
			background-color: rgba(0, 190, 158, 1);
			.add_wrap {
				background-color: white;
			}
		}
		.create-con-btn {
			margin-top: 20px;
			background-color: white;
			border: 1px solid rgba(235, 235, 235, 1);
			margin-bottom: 32px;
			.add_wrap {
				background-color: rgba(246, 246, 246, 1);
			}
		}
		.create-app-btn {
			margin-top: 20px;
			background-color: white;
			border: 1px solid rgba(235, 235, 235, 1);
			.add_wrap {
				background-color: rgba(246, 246, 246, 1);
			}
		}

		.recent-app {
			width: 100%;
			min-height: 300px;
			img {
				width: 120px;
				height: 120px;
			}
		}

		.link-more {
			margin-top: 32px;
			span {
				color: #00967d;
			}
		}
	}
}
</style>
