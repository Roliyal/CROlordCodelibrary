globalThis.global ||= globalThis;

import { createApp } from "vue";
import App from "./App.vue";
import { initRum } from "./rum";

initRum();

createApp(App).mount("#app");
