<template>
	<div class="column">
		<div class="text-h6 text-grey-10">Metadata</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Icon *</div>
			<div class="form-item-value">
				<upload-icon
					:default-img="store.cfg.metadata.icon"
					:max-size="512 * 1024"
					accept=".png, .webp"
					:acceptW="256"
					:acceptH="256"
					@uploaded="uploaded"
				/>
				<div
					class="text-grey-7 q-mt-sm"
					style="font-size: 11px; text-indent: 10px; line-height: 1"
				>
					Your app icon appears in the Terminus Market. The app's icon must be
					in PNG or WEBP format, up to 512 KB, with a size of 256x256 px.
				</div>
			</div>
		</div>

		<div class="form-item row" style="margin-top: 20px">
			<div class="form-item-key text-subtitle2 text-grey-10">App Title *</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Your app title appears in the app market."
					v-model="store.cfg.metadata.title"
					lazy-rules
					:rules="[
						(val) => (val && val.length > 0) || 'Please input the app title'
					]"
					color="teal-4"
					class="form-item-input"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Version Name *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					hint="Your app's version displayed in the Terminus Market. Please specify in the SemVer 2.0.0 format."
					v-model="store.cfg.metadata.versionName"
					lazy-rules
					:rules="[
						(val) => (val && val.length > 0) || 'Please input the version name'
					]"
					color="teal-4"
					class="form-item-input"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">Categories *</div>
			<div class="form-item-value">
				<q-select
					dense
					borderless
					v-model="store.cfg.metadata.categories"
					:options="categoryOptions"
					use-input
					use-chips
					multiple
					no-error-icon
					max-values="2"
					hide-dropdown-icon
					input-debounce="0"
					@new-value="createPort"
					class="form-item-input"
					hint="Used to display your app on different category pages in the Terminus Market."
					:rules="[
						(val) => (val && val.length > 0) || 'Please input categories'
					]"
				>
					<template v-slot:selected-item="scope">
						<q-chip
							square
							icon-remove="sym_r_close"
							removable
							@remove="scope.removeAtIndex(scope.index)"
							:tabindex="scope.tabindex"
							class="q-ma-none tagChip"
						>
							{{ scope.opt }}
						</q-chip>
					</template>
				</q-select>
				<!-- <q-select
          dense
          borderless
          v-model="store.cfg.metadata.target"
          :options="openOptions"
          emit-value
          map-options
          class="form-item-input q-mt-md"
        >
        </q-select> -->
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Short Description *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					hint="A short description appears below app title in the Terminus Market."
					v-model="store.cfg.metadata.description"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || 'Please input the short description'
					]"
					color="teal-4"
					class="form-item-input"
					counter
					maxlength="100"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-grey-10">
				Full Description *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					autogrow
					type="textarea"
					hint="A full description of your app."
					v-model="store.cfg.spec.fullDescription"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || 'Please input the full description'
					]"
					color="teal-4"
					class="form-item-input"
					counter
					maxlength="4000"
				>
				</q-input>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, PropType } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
import { useRoute } from 'vue-router';
import { useDevelopingApps } from '../../stores/app';
import { ApplicationInfo, AppCfg } from '@devbox/core';

import UploadIcon from '../common/UploadIcon.vue';

// const props = defineProps({
//   app: {
//     type: Object as PropType<ApplicationInfo>,
//     required: true,
//   },
// });

const categoryOptions = [
	'Productivity',
	'Utilities',
	'Entertainment',
	'Social Network',
	'Blockchain'
];

const store = useDevelopingApps();

const uploaded = (url) => {
	store.cfg.metadata.icon = url;
};
</script>
<style lang="scss" scoped>
.form-item {
	margin-top: 40px;
	.form-item-key {
		width: 160px;
		height: 40px;
		line-height: 40px;
	}
	.form-item-value {
		flex: 1;
	}
	.tagChip {
		margin-right: 4px;
		border-radius: 10px;
		background: rgba(246, 246, 246, 1);
		color: rgba(31, 24, 20, 1);
		font-size: 12px;
	}
}
</style>
