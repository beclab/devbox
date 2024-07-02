<template>
	<div class="q-pa-md">
		<q-table title="Home" :rows="store.apps" :columns="columns" row-key="name">
			<template v-slot:top-right>
				<!-- <q-btn label="Save" class="col-1" flat color="primary" @click="onSave()" /> -->
				<q-btn
					label="Create new APP"
					class="col-1"
					flat
					color="primary"
					@click="onCreate()"
				/>
			</template>

			<template v-slot:header="props">
				<q-tr :props="props">
					<q-th v-for="col in props.cols" :key="col.name" :props="props">
						{{ col.label }}
					</q-th>
					<q-th auto-width />
				</q-tr>
			</template>

			<template v-slot:body="props">
				<q-tr :props="props">
					<q-td v-for="col in props.cols" :key="col.name" :props="props">
						{{ col.value }}
					</q-td>
					<q-td auto-width>
						<q-btn
							size="sm"
							color="accent"
							round
							dense
							@click="onEdit(props.row)"
							icon="edit"
							style="margin-right: 10px"
						/>
						<q-btn
							size="sm"
							color="green"
							round
							dense
							@click="onCodeBox(props.row)"
							icon="code"
							style="margin-right: 10px"
						/>
						<q-btn
							size="sm"
							color="red"
							round
							dense
							@click="onPreview(props.row)"
							icon="preview"
						/>
					</q-td>
				</q-tr>
			</template>
		</q-table>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useDevelopingApps } from '../stores/app';
//import { AppInfo }  from '../types/appinfo';
import axios from 'axios';
import { ApplicationInfo } from '../../../core/src/index';

const router = useRouter();
const $q = useQuasar();

const store = useDevelopingApps();

const columns = [
	{
		name: 'name',
		required: true,
		label: 'App Name',
		align: 'left',
		field: (row: any) => row.appName,
		format: (val: any, row: any) => `${val}`,
		sortable: true
	},
	{
		name: 'env',
		align: 'center',
		label: 'Dev Env',
		field: 'devEnv',
		sortable: true
	},
	{
		name: 'create_time',
		align: 'left',
		label: 'Create Time',
		field: 'createTime',
		sortable: true
	}
];

onMounted(async () => {
	//
});

function onCreate() {
	router.push({ path: '/create' });
}

const tryToConnect = async (url: string) => {
	try {
		const data = await axios.get(url);
		if (
			data.data ==
			"<h1><a href='https://www.bytetradelab.io/'>Bytetrade</a></h1>"
		) {
			throw data;
		}
	} catch (e: any) {
		console.log(e);
		$q.notify({
			type: 'negative',
			message: 'app is still not installed yet' + e.message
		});
	}
};

async function onEdit(app: ApplicationInfo) {
	router.push({ path: '/app/' + app.id });
}

async function onPreview(app: ApplicationInfo) {
	const testUrl = 'https://' + app.ide;
	await tryToConnect(testUrl);
	if (window.top == window) {
		window.open('//' + app.entrance, '_blank');
	} else {
		await store.openApplication({
			appid: app.id,
			path: ''
		});
	}
}

async function onCodeBox(app: ApplicationInfo) {
	const testUrl = 'https://' + app.ide;
	await tryToConnect(testUrl);
	if (window.top == window) {
		window.open('//' + app.ide, '_blank');
	} else {
		await store.openApplication({
			appid: app.id,
			path: '/proxy/3000/'
		});
	}
}
</script>
