import { createApp } from 'vue';
import App from './App.vue';
import 'bootstrap/dist/css/bootstrap.min.css';

const app = createApp(App);

app.config.globalProperties.$version = process.env.VUE_APP_VERSION;

app.mount('#app');
