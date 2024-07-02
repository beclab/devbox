<template>
	<div class="column">
		<div class="text-h6 text-ink-1">{{ t('enums.CONFIG_TAB.DETAILS') }}</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_upgrade_desc') }}
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
							(val && val.length > 0) || t('config_details_upgrade_rules')
					]"
					input-class="form-item-input text-ink-2"
					counter
					maxlength="4000"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_developer') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_developer_hint')"
					v-model="store.cfg.spec.developer"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_details_developer_rules')
					]"
					input-class="form-item-input text-ink-2"
					counter
					maxlength="30"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_submitter') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_submitter_hint')"
					v-model="store.cfg.spec.submitter"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_details_submitter_rules')
					]"
					input-class="form-item-input text-ink-2"
					counter
					maxlength="30"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_featimage') }}
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
					class="text-ink-2 q-mt-sm"
					style="font-size: 11px; text-indent: 10px; line-height: 1"
				>
					{{ t('config_details_featimage_hint') }}
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_promotemage') }} *
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
				<div class="text-ink-2 q-mt-sm" style="font-size: 11px">
					{{ t('config_details_promotemage_hint') }}
				</div>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_document') }}
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_document_hint')"
					v-model="store.cfg.spec.doc"
					lazy-rules
					class="form-item-input"
					input-class="text-ink-2"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_website') }}
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_website_hint')"
					v-model="store.cfg.spec.website"
					lazy-rules
					class="form-item-input"
					input-class="text-ink-2"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_legalnote') }}
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_legalnote_hint')"
					v-model="store.cfg.spec.legal"
					lazy-rules
					class="form-item-input"
					input-class="text-ink-2"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_license') }}
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_license_hint')"
					v-model="store.cfg.spec.license"
					lazy-rules
					class="form-item-input"
					input-class="text-ink-2"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_sourcecode') }}
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_details_sourcecode_hint')"
					v-model="store.cfg.spec.sourceCode"
					lazy-rules
					class="form-item-input"
					input-class="text-ink-2"
					placeholder="https://"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_details_supportclient') }}
			</div>

			<div class="form-item-value">
				<div class="row items-center justify-between">
					<div class="text-subtitle2 text-ink-2">
						{{ t('config_details_supportclient_desc') }}
					</div>
					<q-btn-dropdown
						class="client-dropdown"
						borderless
						flat
						no-caps
						color="teal-pressed"
						:label="t('config.btn.addClients')"
						dropdown-icon="sym_r_keyboard_arrow_down"
					>
						<q-list flat>
							<q-item
								clickable
								v-close-popup
								@click="addClient(option)"
								v-for="option in supportClient"
								:key="option"
								:disable="option.url ? true : false"
							>
								<q-item-section>
									<q-item-label class="text-ink-2">{{
										option.label
									}}</q-item-label>
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
import { useI18n } from 'vue-i18n';
import { useDevelopingApps } from '../../stores/app';

import UploadFiles from '../common/UploadFiles.vue';
import UploadIcon from '../common/UploadIcon.vue';
import ClientCard from '../common/ClientCard.vue';
import DialogEditClient from '../dialog/DialogEditClient.vue';
import DialogConfirm from '../dialog/DialogConfirm.vue';

const { t } = useI18n();
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
			console.log('supportClient', supportClient.value);
			console.log('store.cfg.spec', store.cfg.spec);

			if (!store.cfg.spec.supportClient) {
				store.cfg.spec.supportClient = {};
			}

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

.client-dropdown {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
