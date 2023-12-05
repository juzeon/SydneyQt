import {createApp} from 'vue'
import App from './App.vue'
import './style.css'

// Vuetify
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import {createVuetify} from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import {aliases, mdi} from "vuetify/iconsets/mdi"
import IndexPage from "./pages/IndexPage.vue"
import * as VueRouter from 'vue-router'
import SettingsPage from "./pages/SettingsPage.vue"

const vuetify = createVuetify({
    components,
    directives,
    icons: {
        defaultSet: 'mdi',
        aliases,
        sets: {
            mdi,
        },
    },
    theme: {
        defaultTheme: 'light',
        themes: {
            light: {
                colors: {
                    primary: '#FF9800',
                }
            },
            dark: {
                dark: true,
                colors: {
                    primary: '#754600',
                }
            }
        }
    }
})
const routes = [
    {path: '/', component: IndexPage},
    {path: '/settings', component: SettingsPage}
]
const router = VueRouter.createRouter({
    history: VueRouter.createWebHashHistory(),
    routes,
})
// @ts-ignore
createApp(App).use(router).use(vuetify).mount('#app')
