<template>
	<div class="column">
		<div class="text-h6 text-ink-1">{{ t('enums.METADATA') }}</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_icon') }} *
			</div>
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
					class="text-ink-2 q-mt-sm"
					style="font-size: 11px; text-indent: 10px; line-height: 1"
				>
					{{ t('config_metadata_icon_hint') }}
				</div>
			</div>
		</div>

		<div class="form-item row" style="margin-top: 20px">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_apptitle') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_metadata_apptitle_hint')"
					v-model="store.cfg.metadata.title"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_metadata_apptitle_rules')
					]"
					class="form-item-input"
					input-class="text-ink-2"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_versionname') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_metadata_versionname_hint')"
					v-model="store.cfg.metadata.versionName"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_metadata_versionname_rules')
					]"
					class="form-item-input"
					input-class="text-ink-2"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_categories') }} *
			</div>
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
					class="form-item-input"
					:hint="t('config_metadata_categories_hint')"
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_metadata_categories_rules')
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
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_shortdesc') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					:hint="t('config_metadata_shortdesc_hint')"
					v-model="store.cfg.metadata.description"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_metadata_shortdesc_rules')
					]"
					class="form-item-input"
					input-class="text-ink-2"
					counter
					maxlength="100"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('config_metadata_fulldesc') }} *
			</div>
			<div class="form-item-value">
				<q-input
					dense
					borderless
					no-error-icon
					autogrow
					type="textarea"
					:hint="t('config_metadata_fulldesc_hint')"
					v-model="store.cfg.spec.fullDescription"
					lazy-rules
					:rules="[
						(val) =>
							(val && val.length > 0) || t('config_metadata_fulldesc_rules')
					]"
					class="form-item-input"
					input-class="text-ink-2"
					counter
					maxlength="4000"
				>
				</q-input>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useDevelopingApps } from '../../stores/app';
import { useI18n } from 'vue-i18n';

import UploadIcon from '../common/UploadIcon.vue';

const categoryOptions = [
	'Productivity',
	'Utilities',
	'Entertainment',
	'Social Network',
	'Blockchain'
];

const { t } = useI18n();
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
		background: $background-hover;
		color: $ink-2;
		font-size: 12px;
	}
}
</style>
