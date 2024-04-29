<template>
	<div class="column">
		<div class="text-h6 text-grey-10">Details</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Upgrade Description
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					type="textarea"
					hint="Describe what is new in this upgraded version."
					v-model="store.cfg.spec.upgradeDescription"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || 'Please input the upgrade description'
					]"
					color="teal-4"
					input-class="form-item-input"
					counter
					maxlength="4000"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Developer *</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					hint="The name of developer of this app."
					v-model="store.cfg.spec.developer"
					lazy-rules
					:rules="[
						(val) => (val && val.length > 0) || 'Please input the developer'
					]"
					color="teal-4"
					input-class="form-item-input"
					counter
					maxlength="30"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Submitter *</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					hint="The name of submitter who submits this app to the app market."
					v-model="store.cfg.spec.submitter"
					lazy-rules
					:rules="[
						(val) => (val && val.length > 0) || 'Please input the submitter'
					]"
					color="teal-4"
					input-class="form-item-input"
					counter
					maxlength="30"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Featured Image
			</div>
			<div class="form-item-value">
				<upload-icon
					:default-img="store.cfg.spec.featuredImage"
					:max-size="8 * 1024 * 1024"
					accept=".jpg, .png, .webp"
					:acceptW="1440"
					:acceptH="900"
					message="Drag and drop a JPEG, PNG or WEBP file here to upload"
					@uploaded="uploaded"
				/>
				<div
					class="text-grey-7 q-mt-sm"
					style="font-size: 11px; text-indent: 10px; line-height: 1"
				>
					Upload a featured image for the app. The image must be in JPEG, PNG or
					WEBP format, up to 8MB each, with a size of 1440x900 px.
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Promote Image *
			</div>
			<div class="form-item-value">
				<upload-files
					:default-imgs="store.cfg.spec.promoteImage"
					:max-size="8 * 1024 * 1024"
					accept=".png, .webp, .jpg, jpeg"
					:acceptW="1440"
					:acceptH="900"
					:maxfiles="8"
					@uploaded="uploaded"
					@deleteDefaultImg="deleteDefaultImg"
				/>
				<div class="text-grey-7 q-mt-sm" style="font-size: 11px">
					Upload 2-8 app screenshots for promotion. Screenshots must be in JPEG,
					PNG or WEBP format, up to 8MB each, with a size of 1440x900 px.
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Document</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Add a link to the documents or user manual for your app."
					v-model="store.cfg.spec.doc"
					lazy-rules
					color="teal-4"
					class="form-item-input"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Website</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Add a link to your official website, if you have one."
					v-model="store.cfg.spec.website"
					lazy-rules
					color="teal-4"
					class="form-item-input"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Legal Note</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Add a link to the legal notes that you want to display on the app market."
					v-model="store.cfg.spec.legal"
					lazy-rules
					color="teal-4"
					class="form-item-input"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">License</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Add a link to your app's license agreement."
					v-model="store.cfg.spec.license"
					lazy-rules
					color="teal-4"
					class="form-item-input"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Source Code</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Add a link to your app's source code."
					v-model="store.cfg.spec.sourceCode"
					lazy-rules
					color="teal-4"
					class="form-item-input"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Support Client
			</div>

			<div class="form-item-value">
				<div class="row items-center justify-between">
					<div class="text-subtitle2 text-grey-8">
						Add links to your app clients on other platforms.
					</div>
					<q-btn-dropdown
						borderless
						flat
						no-caps
						color="teal-8"
						style="border: 1px solid rgba(235, 235, 235, 1); border-radius: 8px"
						label="Add Clients"
						dropdown-icon="sym_r_keyboard_arrow_down"
					>
						<q-list>
							<q-item
								clickable
								v-close-popup
								@click="addClient(option)"
								v-for="option in supportClient"
								:key="option"
								:disable="option.url ? true : false"
							>
								<q-item-section>
									<q-item-label>{{ option.label }}</q-item-label>
								</q-item-section>
							</q-item>
						</q-list>
					</q-btn-dropdown>
				</div>
				<template v-for="(item, index) in supportClient" :key="index">
					<ClientCard
						:data="item"
						v-if="item.url"
						@editClient="editClient"
						@deleteClient="deleteClient"
					/>
				</template>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../../stores/app';

import UploadFiles from '../common/UploadFiles.vue';
import UploadIcon from '../common/UploadIcon.vue';
import ClientCard from '../common/ClientCard.vue';
import DialogEditClient from '../dialog/DialogEditClient.vue';
import DialogConfirm from '../dialog/DialogConfirm.vue';

const store = useDevelopingApps();
const $q = useQuasar();

const supportClient = ref([
	{
		label: 'Android',
		name: 'android',
		type: false,
		url: '',
		hint: 'Android mobile apps in the Google Play'
	},
	{
		label: 'iOS',
		name: 'ios',
		type: false,
		url: '',
		hint: 'iPhone/iPad apps in the App Store'
	},
	{
		label: 'Edge',
		name: 'edge',
		type: false,
		url: '',
		hint: 'Edge Extension in the Edge Addons'
	},
	{
		label: 'Mac',
		name: 'mac',
		type: false,
		url: '',
		hint: 'Mac apps in the Mac App Store'
	},
	{
		label: 'Windows',
		name: 'windows',
		type: false,
		url: '',
		hint: 'Download link for windows client'
	},
	{
		label: 'Linux',
		name: 'linux',
		type: false,
		url: '',
		hint: 'Download link for linux client'
	}
]);

const editClient = (data) => {
	$q.dialog({
		component: DialogEditClient,
		componentProps: {
			data
		}
	})
		.onOk((data) => {
			store.cfg.spec.supportClient[data.name] = data.url;
			updateSupportClient();
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const deleteClient = ({ name, label }) => {
	const title = `Delete ${label}`;
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: title,
			message: 'Are you sure to delete this client?'
		}
	})
		.onOk((data) => {
			console.log('OK', data);
			store.cfg.spec.supportClient[name] = '';
			updateSupportClient();
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const addClient = (option) => {
	$q.dialog({
		component: DialogEditClient,
		componentProps: {
			data: option,
			mode: 'create'
		}
	})
		.onOk((data) => {
			console.log('OK', data);
			for (let i = 0; i < supportClient.value.length; i++) {
				const element = supportClient.value[i];
				if (element.name === data.name) {
					store.cfg.spec.supportClient[data.name] = data.url;
				}
			}

			updateSupportClient();
		})
		.onCancel(() => {
			console.log('Cancel');
		})
		.onDismiss(() => {
			console.log('Called on OK or Cancel');
		});
};

const updateSupportClient = () => {
	if (!store.cfg.spec.supportClient) {
		return false;
	}
	for (let i = 0; i < supportClient.value.length; i++) {
		const element = supportClient.value[i];
		element.url = store.cfg.spec.supportClient[element.name];
		if (store.cfg.spec.supportClient[element.name]) {
			element.type = true;
		} else {
			element.type = false;
		}
	}
};

const uploaded = (url: string) => {
	if (!store.cfg.spec.promoteImage) {
		store.cfg.spec.promoteImage = [];
	}
	store.cfg.spec.promoteImage.push(url);
};

const deleteDefaultImg = (url: string) => {
	const index = store.cfg.spec.promoteImage.findIndex((item) => item === url);
	store.cfg.spec.promoteImage.splice(index, 1);
};

onMounted(() => {
	updateSupportClient();
});
</script>

<style lang="scss" scoped>
.form-item {
	margin-top: 20px;
	.form-item-key {
		width: 160px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
