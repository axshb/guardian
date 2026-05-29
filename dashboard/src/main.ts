import { createApp } from 'vue';
import { createPinia } from 'pinia';

import App from './App.vue';
import router from './router';
import ToastContainer from './components/ToastContainer.vue';
import './style.css';
import './themes/index.css';

const app = createApp(App);

app.use(createPinia());
app.use(router);
app.component('ToastContainer', ToastContainer);

app.mount('#app');
