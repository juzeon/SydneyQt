<script setup lang="ts">

import Scaffold from "../components/Scaffold.vue"
import {useRouter} from "vue-router"
import {useSettings} from "../composables"
import {computed, onMounted, ref} from "vue"
import {useTheme} from "vuetify"

let theme = useTheme()
let router = useRouter()
let {config, fetch: fetchingSettings} = useSettings()
let loading = ref(true)
onMounted(() => {
  loading.value = true
  fetchingSettings().then(() => {
    loading.value = false
  })
})
let fontStyle = computed(() => {
  return {
    'font-family': "'" + config.value.font_family + "'",
    'font-size': config.value.font_size + 'px',
  }
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
      <div class="overflow-auto" v-if="!loading">
        <v-container class="d-flex flex-column">
          <p class="text-h4 mb-3">Settings</p>
          <v-card title="Network" class="my-3">
            <v-card-text>
              <v-tooltip text="Enter a HTTP proxy (e.g. http://127.0.0.1:7890). Leave blank to disable proxy."
                         location="bottom">
                <template #activator="{props}">
                  <v-text-field label="Proxy" v-model="config.proxy" v-bind="props"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip
                  text="FQDN for the websocket endpoint of Sydney, without any prefixes (protocol names) or suffixes (slashes)."
                  location="bottom">
                <template #activator="{props}">
                  <v-text-field label="Wss Domain" v-model="config.wss_domain" v-bind="props"></v-text-field>
                </template>
              </v-tooltip>
            </v-card-text>
          </v-card>
          <v-card title="Display" class="my-3">
            <v-card-text>
              <v-switch label="Dark Mode" :model-value="config.dark_mode"
                        @update:model-value="onDarkModeSwitch"></v-switch>
              <v-switch label="No Suggestion" v-model="config.no_suggestion"></v-switch>
              <v-tooltip text="Default: SF" location="bottom">
                <template #activator="{props}">
                  <v-text-field v-bind="props" label="Font Family" v-model="config.font_family"></v-text-field>
                </template>
              </v-tooltip>
              <v-slider min="10" max="30" step="1" thumb-label="always" v-model="config.font_size"
                        label="Font Size"></v-slider>
              <div class="text-center mb-3" :style="fontStyle">Font Example</div>
              <!-- TODO v-slider tooltip -->
              <v-tooltip text="Height of the textarea of user input, in vh." location="bottom">
                <template #activator="{props}">
                  <v-slider v-bind="props" step="1" min="10" max="60" label="User Input Textarea Height"
                            v-model="config.stretch_factor"
                            thumb-label="always"></v-slider>
                </template>
              </v-tooltip>
            </v-card-text>
          </v-card>
          <v-card title="Accessibility" class="my-3">
            <v-card-text>
              <v-select label="Enter Mode" v-model="config.enter_mode"
                        :items="['Enter','Ctrl+Enter']"></v-select>
            </v-card-text>
          </v-card>
        </v-container>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>

</style>