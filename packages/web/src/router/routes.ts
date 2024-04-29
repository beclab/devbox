import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/MainLayout.vue'),
		beforeEnter: (to, from, next) => {
			if (to.fullPath == '/') {
				return next({ path: '/home' });
			}
			next();
		},
		children: [
			{ path: '/home', component: () => import('pages/HomePage.vue') },
			{ path: '/list', component: () => import('pages/ListPage.vue') },
			{
				path: '/create',
				component: () => import('pages/CreatePage.vue')
			},
			{
				path: '/containers',
				component: () => import('pages/ContainerPage.vue')
			},

			// {
			//   path: '/list',
			//   name: 'list',
			//   component: () => import('pages/ListPage.vue'),
			// },
			{
				path: '/app/:id',
				component: () => import('pages/ApplicationPage.vue')
			},
			// {
			//   path: '/edit',
			//   name: 'edit',
			//   component: () => import('pages/EditPage.vue'),
			// },
			{
				path: '/help',
				name: 'help',
				component: () => import('pages/HelpPage.vue')
			}
		]
	},

	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/ErrorNotFound.vue')
	}
];

export default routes;
