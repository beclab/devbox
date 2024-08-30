import { boot } from 'quasar/wrappers';
import { Notify } from 'quasar';
import axios, { AxiosInstance, AxiosResponse } from 'axios';

declare module '@vue/runtime-core' {
	interface ComponentCustomProperties {
		$axios: AxiosInstance;
	}
}

// Be careful when using SSR for cross-request state pollution
// due to creating a Singleton instance here;
// If any client changes this (global) instance, it might be a
// good idea to move this instance creation inside of the
// "export default () => {}" function below (which runs individually
// for each client)
const api = axios.create({ baseURL: 'https://api.example.com' });

export default boot(({ app }) => {
	// for use inside Vue files (Options API) through this.$axios and this.$api

	app.config.globalProperties.$axios = axios;
	// ^ ^ ^ this will allow you to use this.$axios (for Vue Options API form)
	//       so you won't necessarily have to import axios in each vue file

	app.config.globalProperties.$api = api;
	// ^ ^ ^ this will allow you to use this.$api (for Vue Options API form)
	//       so you can easily perform requests against your app's API

	app.config.globalProperties.$axios.interceptors.response.use(
		(response: AxiosResponse) => {
			//const store = useAccountStore();
			// if (
			//   response.config.url &&
			//   response.config.url.indexOf(store.space.url) !== -1 &&
			//   response.config.method === 'post'
			// ) {
			//   const data = response.data;

			//   if (data && data.code === 401) {
			//     return response;
			//   }

			//   if (data.code !== 200) {
			//     throw new Error(data.message);
			//   }

			//   return data;
			// } else {

			console.log('response.config.method ' + response.config.method);
			if (response.config.url && response.config.url.endsWith('.md')) {
				return response;
			}
			//  else if (response.config.method?.toLowerCase() == 'put') {
			//   return response.data;
			// }
			else {
				const data = response.data;

				if (!data || (data.code != 200 && data.code != 0)) {
					if (data.msg || data.message || data.data.result) {
						Notify.create({
							type: 'negative',
							message: data.msg || data.message || data.data.result
						});
					}

					throw Error(data.msg);
				}

				// if (data.message) {
				// 	Notify.create({
				// 		message: data.message
				// 	});
				// }

				return data.data;
			}
		}
		//  }
	);
});

export { api };
