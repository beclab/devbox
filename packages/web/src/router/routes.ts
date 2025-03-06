import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/MainLayout.vue'),
		beforeEnter: (to, _from, next) => {
			if (to.fullPath == '/') {
				return next({ path: '/home' });
			}
			next();
		},
		children: [
			{ path: '/home', component: () => import('pages/HomePage.vue') },
			{
				path: '/create',
				component: () => import('pages/CreatePage.vue')
			},
			{
				path: '/containers',
				component: () => import('pages/ContainerPage.vue')
			},
			{
				path: '/app/:id/:path?',
				component: () => import('../pages/ApplicationPage.vue')
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
