<template>
	<q-card-actions class="row justify-end items-center q-mt-md q-mb-sm">
		<q-btn
			v-if="showCancel"
			clickable
			dense
			no-caps
			flat
			type="reset"
			class="but-cancel row justify-center items-center"
		>
			{{ cancelText }}
		</q-btn>

		<q-btn
			clickable
			dense
			no-caps
			flat
			type="submit"
			class="but-creat row justify-center items-center"
		>
			{{ okText }}
		</q-btn>
	</q-card-actions>
</template>

<script lang="ts" setup>
defineProps({
	okText: {
		type: String,
		default: 'Sumbit',
		required: false
	},
	cancelText: {
		type: String,
		default: 'Cancel',
		required: false
	},
	showCancel: {
		type: Boolean,
		default: true,
		required: false
	}
});
</script>

<style scoped lang="scss">
.but-creat {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	background: rgba(0, 190, 158, 1);
	color: #ffffff;
	padding: 4px 12px !important;
}
.but-cancel {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	border: 1px solid #e0e0e0;
	color: #4d4742;
	padding: 4px 12px !important;
}
</style>
