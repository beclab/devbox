import { defineStore } from 'pinia';
import { Notify } from 'quasar';
import axios from 'axios';
import { CONFIG_TAB } from './../types/constants';
import {
	CreateApplicationConfig,
	ApplicationInfo,
	Container,
	OpenApplicationConfig,
	AppCfg,
	AppContainer,
	MenuTab
} from '@devbox/core';

export type DevelopingApps = {
	url: string;
	apps: ApplicationInfo[];
	containers: Container[];
	cfg: AppCfg | null;
	configMenu: MenuTab[];
	notif: any;
};

export const useDevelopingApps = defineStore('app', {
	state() {
		return {
			url: '',
			apps: [],
			containers: [],
			cfg: null,
			configMenu: [
				{
					name: CONFIG_TAB.METADATA,
					label: 'Metadata',
					status: 0
				},
				{
					name: CONFIG_TAB.DETAILS,
					label: 'Details',
					status: 0
				},
				{
					name: CONFIG_TAB.SPACE,
					label: 'Specs',
					status: 0
				},
				{
					name: CONFIG_TAB.PERMISSIONS,
					label: 'Permissions',
					status: 0
				},
				{
					name: CONFIG_TAB.OPTIONS,
					label: 'Options',
					status: 0
				}
			],
			notif: null
		} as DevelopingApps;
	},
	actions: {
		setUrl(new_url: string) {
			this.url = new_url;
			if (new_url) {
				localStorage.setItem('url', new_url);
			}
		},
		async getApps() {
			try {
				this.apps = await axios.get(this.url + '/api/command/list-app');
			} catch (e) {
				console.log(e);
			}
		},

		async getMyContainers() {
			try {
				this.containers = await axios.get(this.url + '/api/list-my-containers');
				if (!this.containers) this.containers = [];
			} catch (e) {
				console.log(e);
			}
		},

		async createApplication(config: CreateApplicationConfig) {
			try {
				const data: any = await axios.post(
					this.url + '/api/command/create-app',
					config
				);

				return data.appId;
			} catch (e) {
				console.log(e);
			}
		},

		async openApplication(config: OpenApplicationConfig) {
			try {
				const data: any = await axios.post(
					this.url + '/api/command/open-application',
					config
				);
				return data.appId;
			} catch (e) {
				console.log(e);
			}
		},

		async getAppCfg(app_name: string): Promise<AppCfg> {
			try {
				const data: AppCfg = await axios.get(
					this.url + '/api/app-cfg?app=' + app_name
				);
				return data;
			} catch (e) {
				console.log(e);
			}
		},

		async getAppContainer(app_name: string): Promise<Container[]> {
			this.notif = Notify.create({
				type: 'ongoing',
				color: 'grey-4',
				textColor: 'grey-9',
				message: 'Data Loading...'
			});

			try {
				const data: AppContainer[] = await axios.get(
					this.url + '/api/list-app-containers?app=' + app_name
				);
				const res: any = [];
				for (const d of data) {
					const f = this.containers.find(
						(c) =>
							c.podSelector === d.podSelector &&
							c.containerName == d.containerName &&
							c.appName == app_name
					);
					if (f) {
						res.push({
							image: d.image,
							...f
						});
					} else {
						res.push({
							...d
						});
					}
				}
				this.notif();
				return res;
			} catch (e) {
				console.log(e);
				this.notif();
				return [];
			}
		},

		async bindContainer({
			containerId,
			appId,
			podSelector,
			containerName,
			devEnv,
			devContainerName
		}: {
			containerId?: number;
			appId: number;
			podSelector: string;
			containerName: string;
			devEnv?: string;
			devContainerName: string;
		}): Promise<void> {
			try {
				await axios.post(this.url + '/api/bind-container', {
					containerId,
					appId,
					podSelector,
					containerName,
					devEnv,
					devContainerName
				});
			} catch (e) {
				console.log(e);
			}
		},

		async unbindContainer({
			containerId,
			appId,
			podSelector,
			containerName
		}: {
			containerId: number;
			appId: number;
			podSelector: string;
			containerName: string;
		}): Promise<void> {
			try {
				const data = await axios.post(this.url + '/api/unbind-container', {
					containerId,
					appId,
					podSelector,
					containerName
				});
			} catch (e) {
				console.log(e);
			}
		},

		async putAppCfg(app_name: string, cfg: AppCfg): Promise<AppCfg> {
			await axios.post(this.url + '/api/app-cfg?app=' + app_name, cfg);
		},

		async getAppState(app_name: string): Promise<AppCfg> {
			const data: AppCfg = await axios.get(
				this.url + '/api/app-state?app=' + app_name
			);

			return data;
		},

		async getAppStatus(app_name: string): Promise<AppCfg> {
			const data: AppCfg = await axios.get(
				this.url + '/api/app-status?app=' + app_name
			);

			return data;
		}
	}
});
