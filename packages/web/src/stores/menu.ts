import { defineStore } from 'pinia';
import { MenuLabelType, DocumenuType } from '@devbox/core';
import { useDevelopingApps } from './app';
import { i18n } from '../boot/i18n';

const store = useDevelopingApps();

export enum MenuLabel {
	DEVBOX = 'DevBox',
	HOME = 'Home',
	CONTAINERS = 'Containers',
	HELP = 'Help',
	APPLICATIONS = 'Applications'
}

export type DataState = {
	homeMenu: MenuLabelType[];
	applicationMenu: MenuLabelType[];
	documentList: DocumenuType[];
	currentItem: string;
	appCurrentItem: string;
};

export const useMenuStore = defineStore('menu', {
	state() {
		return {
			homeMenu: [
				{
					label: i18n.global.t(`enums.${MenuLabel.DEVBOX}`),
					key: MenuLabel.DEVBOX,
					icon: '',
					children: [
						{
							label: i18n.global.t(`enums.${MenuLabel.HOME}`),
							key: MenuLabel.HOME,
							icon: 'sym_r_home',
							muted: true
						},
						{
							label: i18n.global.t(`enums.${MenuLabel.CONTAINERS}`),
							key: MenuLabel.CONTAINERS,
							icon: 'sym_r_deployed_code',
							muted: true
						}
						// {
						//   label: MenuLabel.HELP,
						//   key: MenuLabel.HELP,
						//   icon: 'sym_o_inbox_customize',
						// },
					]
				}
			],
			applicationMenu: [
				{
					label: i18n.global.t(`enums.${MenuLabel.APPLICATIONS}`),
					key: MenuLabel.APPLICATIONS,
					icon: '',
					children: []
				}
			],

			currentItem: MenuLabel.HOME,
			appCurrentItem: 'files',
			documentList: [
				{
					id: 1,
					message: 'DevBox Tutorial',
					link: 'https://docs.olares.xyz/developer/develop/tutorial/devbox.html'
				},
				{
					id: 2,
					message: 'Introduction to Basic Concepts of Olares Application',
					link: 'https://docs.olares.xyz/manual/concepts/application.html'
				},
				{
					id: 3,
					message: 'Learn about Olares Application Chart',
					link: 'https://docs.olares.xyz/developer/develop/package/chart.html'
				},
				{
					id: 4,
					message: 'Configuration Guideline for OlaresManifest',
					link: 'https://docs.olares.xyz/developer/develop/package/manifest.html'
				},
				{
					id: 5,
					message: 'Olares Market protocol overview',
					link: 'https://docs.olares.xyz/manual/tasks/install-uninstall-update.html'
				},
				{
					id: 6,
					message: 'How to sbmit an application to the Olares Market',
					link: 'https://docs.olares.xyz/developer/develop/submit/'
				}
			]
		} as DataState;
	},
	getters: {
		menuList(state) {
			return [...state.homeMenu, ...state.applicationMenu];
		}
	},
	actions: {
		updateApplications() {
			this.applicationMenu[0].children = [];
			for (const app of store.apps) {
				this.applicationMenu[0].children.push({
					label: app.appName,
					key: `/app/${app.id}`,
					icon: 'sym_r_grid_view'
				});
			}
		}
	}
});
