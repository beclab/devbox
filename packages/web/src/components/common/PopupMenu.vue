<template>
	<q-menu class="menu-wrap">
		<q-list dense padding>
			<q-item
				class="row items-center justify-start"
				clickable
				v-close-popup
				v-for="item in items"
				:key="item.name"
				@click="handleEvent(item.name)"
			>
				<q-icon :name="item.icon" size="20px" class="q-mr-sm" color="ink-2" />
				<q-item-section class="menuName text-ink-2">
					{{ item.label }}</q-item-section
				>
			</q-item>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { OPERATE_ACTION } from '../../types/constants';

const props = defineProps({
	items: {
		type: Object,
		required: false
	},
	path: {
		type: String,
		required: false
	},
	label: {
		type: String,
		required: false
	}
});

const emits = defineEmits(['handleEvent']);

const handleEvent = async (action: OPERATE_ACTION) => {
	emits('handleEvent', action, props.path, props.label);
};
</script>
