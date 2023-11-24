<script setup lang="ts">

import Scaffold from "../components/Scaffold.vue"
import {useRouter} from "vue-router"
import {useSettings} from "../composables"
import {onMounted} from "vue"
import {useTheme} from "vuetify"

let theme = useTheme()
let router = useRouter()
let {config, fetch: fetchingSettings} = useSettings()
onMounted(() => {
  fetchingSettings()
})

function onDarkModeSwitch() {
  config.value.dark_mode = !config.value.dark_mode
  theme.global.name.value = config.value.dark_mode ? 'dark' : 'light'
}
</script>

<template>
  <scaffold>
    <template #left-top>
      <v-btn icon @click="router.push('/')">
        <v-icon>mdi-arrow-left</v-icon>
      </v-btn>
    </template>
    <template #right-top>
      <div></div>
    </template>
    <template #default>
      <div class="overflow-auto">
        <v-container class="d-flex flex-column">
          <p class="text-h5 mb-3">Settings</p>
          <v-select label="Enter Mode" v-model="config.enter_mode"
                    :items="['Enter','Ctrl+Enter']"></v-select>
          <v-tooltip text="Enter a HTTP proxy (e.g. http://127.0.0.1:7890). Leave blank to disable proxy."
                     location="bottom">
            <template #activator="{props}">
              <v-text-field label="Proxy" v-model="config.proxy" v-bind="props"></v-text-field>
            </template>
          </v-tooltip>
          <v-switch label="Dark Mode" :model-value="config.dark_mode" @update:model-value="onDarkModeSwitch"></v-switch>
        </v-container>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>

</style>