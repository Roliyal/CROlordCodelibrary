/*
import { createApp } from 'vue'
import App from './App.vue'
import 'bootstrap/dist/css/bootstrap.min.css';

createApp(App).mount('#app')
*/

import { createApp } from 'vue';
import App from './App.vue';

const environment = process.env.VUE_APP_ENV;
if (environment === 'gray') {
    import('./gray/css/main.css');
} else {
    import('./base/css/main.css');
}

createApp(App).mount('#app');
