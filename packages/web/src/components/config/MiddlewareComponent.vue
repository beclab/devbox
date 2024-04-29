<template>
	<div class="column q-gutter-md">
		<q-card>
			<q-item>
				<q-item-section>
					<q-item-label>Postgres</q-item-label>
				</q-item-section>
			</q-item>
			<q-input
				filled
				v-model="store.cfg.middleware.postgres.username"
				label="Postgres Username"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>
			<q-input
				filled
				v-model="store.cfg.middleware.postgres.password"
				label="Postgres Password"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>
			<div class="row justify-end">
				<q-btn @click="addPostgresDataBase">New</q-btn>
			</div>
			<div
				v-for="(image, index) of store.cfg.middleware.postgres.databases"
				:key="`ii` + index"
			>
				<q-item class="row justify-end">
					<q-item-section>
						<q-input
							filled
							v-model="store.cfg.middleware.postgres.databases[index].name"
							label="Database Name"
							lazy-rules
							:rules="[
								(val) => (val && val.length > 0) || 'Please input the app name'
							]"
						/>
					</q-item-section>

					<q-item-section>
						<q-checkbox
							v-model="
								store.cfg.middleware.postgres.databases[index].distributed
							"
							>Distributed</q-checkbox
						>
					</q-item-section>

					<q-btn @click="deletePostgresDataBase(index)">Delete</q-btn>
				</q-item>
			</div>
		</q-card>

		<q-card>
			<q-item>
				<q-item-section>
					<q-item-label>Redis</q-item-label>
				</q-item-section>
			</q-item>

			<q-input
				filled
				v-model="store.cfg.middleware.redis.password"
				label="Redis Password"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>
			<div class="row justify-end">
				<q-btn @click="addRedisDataBase">New</q-btn>
			</div>
			<div
				v-for="(image, index) of store.cfg.middleware.redis.databases"
				:key="`ii` + index"
			>
				<q-item class="row justify-end">
					<q-item-section>
						<q-input
							filled
							v-model="store.cfg.middleware.redis.databases[index].name"
							label="Database Name"
							lazy-rules
							:rules="[
								(val) => (val && val.length > 0) || 'Please input the app name'
							]"
						/>
					</q-item-section>

					<q-btn @click="deleteRedisDataBase(index)">Delete</q-btn>
				</q-item>
			</div>
		</q-card>

		<q-card>
			<q-item>
				<q-item-section>
					<q-item-label>Mongodb</q-item-label>
				</q-item-section>
			</q-item>

			<q-input
				filled
				v-model="store.cfg.middleware.mongodb.username"
				label="Mongodb Username"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>

			<q-input
				filled
				v-model="store.cfg.middleware.mongodb.password"
				label="Mongodb Password"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>
			<div class="row justify-end">
				<q-btn @click="addMongodbDataBase">New</q-btn>
			</div>
			<div
				v-for="(image, index) of store.cfg.middleware.mongodb.databases"
				:key="`ii` + index"
			>
				<q-item class="row justify-end">
					<q-item-section>
						<q-input
							filled
							v-model="store.cfg.middleware.mongodb.databases[index].name"
							label="Database Name"
							lazy-rules
							:rules="[
								(val) => (val && val.length > 0) || 'Please input the app name'
							]"
						/>
					</q-item-section>

					<q-btn @click="deleteMongodbDataBase(index)">Delete</q-btn>
				</q-item>
			</div>
		</q-card>

		<q-card>
			<q-item>
				<q-item-section>
					<q-item-label>ZincSearch</q-item-label>
				</q-item-section>
			</q-item>

			<q-input
				filled
				v-model="store.cfg.middleware.zincSearch.username"
				label="ZincSearch username"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>

			<q-input
				filled
				v-model="store.cfg.middleware.zincSearch.password"
				label="ZincSearch Password"
				lazy-rules
				:rules="[
					(val) => (val && val.length > 0) || 'Please input the app name'
				]"
			/>
			<div class="row justify-end">
				<q-btn @click="addZincDataBase">New</q-btn>
			</div>
			<div
				v-for="(image, index) of store.cfg.middleware.zincSearch.indexes"
				:key="`ii` + index"
			>
				<q-item class="row justify-end">
					<q-item-section>
						<q-input
							filled
							v-model="store.cfg.middleware.zincSearch.indexes[index].name"
							label="Index Name"
							lazy-rules
							:rules="[
								(val) => (val && val.length > 0) || 'Please input the app name'
							]"
						/>
					</q-item-section>

					<q-btn @click="deleteZincDataBase(index)">Delete</q-btn>
				</q-item>
			</div>
		</q-card>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, PropType } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';
import { useRoute } from 'vue-router';
import { useDevelopingApps } from '../../stores/app';
import { ApplicationInfo, AppCfg } from '@devbox/core';

const store = useDevelopingApps();

function addPostgresDataBase() {
	store.cfg.middleware.postgres.databases.push({
		name: 'db1',
		distributed: false
	});
}

function deletePostgresDataBase(index: number) {
	store.cfg.middleware.postgres.databases.splice(index, 1);
}

function addRedisDataBase() {
	store.cfg.middleware.redis.databases.push({
		name: 'namespace1'
	});
}

function deleteRedisDataBase(index: number) {
	store.cfg.middleware.redis.databases.splice(index, 1);
}

function addMongodbDataBase() {
	store.cfg.middleware.mongodb.databases.push({
		name: 'db1'
	});
}

function deleteMongodbDataBase(index: number) {
	store.cfg.middleware.mongodb.databases.splice(index, 1);
}

function addZincDataBase() {
	store.cfg.middleware.zincSearch.indexes.push({
		name: 'index name'
	});
}

function deleteZincDataBase(index: number) {
	store.cfg.middleware.zincSearch.indexes.splice(index, 1);
}
</script>
