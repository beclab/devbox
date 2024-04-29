import { defineStore } from 'pinia';
import { MenuLabelType, DocumenuType } from '@devbox/core';
import { useDevelopingApps } from './app';

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
					label: MenuLabel.DEVBOX,
					key: MenuLabel.DEVBOX,
					icon: '',
					children: [
						{
							label: MenuLabel.HOME,
							key: MenuLabel.HOME,
							icon: 'sym_r_home'
						},
						{
							label: MenuLabel.CONTAINERS,
							key: MenuLabel.CONTAINERS,
							icon: 'sym_r_deployed_code'
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
					label: MenuLabel.APPLICATIONS,
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
					message: 'Quick start',
					link: 'https://www.baidu.com/'
				},
				{
					id: 2,
					message: 'Application chart specification guideline',
					link: 'https://www.baidu.com/'
				},
				{
					id: 3,
					message: 'Learn about submission process',
					link: 'https://www.baidu.com/'
				},
				{
					id: 4,
					message: 'How to submit application to the App Market',
					link: 'https://www.baidu.com/'
				},
				{
					id: 5,
					message: 'How to manage and maintain application in the App Market',
					link: 'https://www.baidu.com/'
				},
				{
					id: 6,
					message: 'About helm charts format',
					link: 'https://www.baidu.com/'
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
					icon: 'sym_o_grid_view'
				});
			}
		}
	}
});
