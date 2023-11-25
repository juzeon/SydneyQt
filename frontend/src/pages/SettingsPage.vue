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

function revokeReplyCountInputRule(v: any) {
  let i = parseInt(v)
  if (isNaN(i)) return false
  return i >= 0
}

function onRevokeReplyCountChanged(v: string) {
  let i = parseInt(v)
  if (isNaN(i) || i < 0) return
  config.value.revoke_reply_count = i
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
                  <v-text-field color="primary" label="Proxy" v-model="config.proxy" v-bind="props"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip
                  text="FQDN for the websocket endpoint of Sydney, without any prefixes (protocol names) or suffixes (slashes)."
                  location="bottom">
                <template #activator="{props}">
                  <v-text-field color="primary" label="Wss Domain" v-model="config.wss_domain"
                                v-bind="props"></v-text-field>
                </template>
              </v-tooltip>
            </v-card-text>
          </v-card>
          <v-card title="Display" class="my-3">
            <v-card-text>
              <v-switch color="primary" label="Dark Mode" :model-value="config.dark_mode"
                        @update:model-value="onDarkModeSwitch"></v-switch>
              <v-switch color="primary" label="No Suggestion" v-model="config.no_suggestion"></v-switch>
              <v-tooltip text="Default: SF" location="bottom">
                <template #activator="{props}">
                  <v-text-field color="primary" v-bind="props" label="Font Family"
                                v-model="config.font_family"></v-text-field>
                </template>
              </v-tooltip>
              <v-slider color="primary" min="10" max="30" step="1" thumb-label="always" v-model="config.font_size"
                        label="Font Size"></v-slider>
              <div class="text-center mb-3" :style="fontStyle">Font Example</div>
              <!-- TODO v-slider tooltip -->
              <v-tooltip text="Height of the textarea of user input, in vh." location="bottom">
                <template #activator="{props}">
                  <v-slider color="primary" v-bind="props" step="1" min="10" max="60" label="User Input Textarea Height"
                            v-model="config.stretch_factor"
                            thumb-label="always"></v-slider>
                </template>
              </v-tooltip>
            </v-card-text>
          </v-card>
          <v-card title="Accessibility" class="my-3">
            <v-card-text>
              <v-tooltip text="Keyboard shortcut to send the user input." location="bottom">
                <template #activator="{props}">
                  <v-select v-bind="props" color="primary" label="Enter Mode" v-model="config.enter_mode"
                            :items="['Enter','Ctrl+Enter']"></v-select>
                </template>
              </v-tooltip>
              <v-tooltip
                  text="Send this text automatically when Bing revokes a message
                  if Revoke Reply Count is larger than zero or set as a suggested response otherwise."
                  location="bottom">
                <template #activator="{props}">
                  <v-text-field v-bind="props" color="primary" label="Revoke Reply Text"
                                v-model="config.revoke_reply_text"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip
                  text="Maximum count for auto-reply when Bing revokes a message.
                  Set this to 0 to disable and show a suggestion only."
                  location="bottom">
                <template #activator="{props}">
                  <v-text-field v-bind="props" color="primary" label="Revoke Reply Count"
                                :model-value="config.revoke_reply_count"
                                @update:model-value="onRevokeReplyCountChanged"
                                :rules="[revokeReplyCountInputRule]"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip text="Whether to send the selected quick response immediately
              rather than append it to the user input textarea if it is empty." location="bottom">
                <template #activator="{props}">
                  <v-switch v-bind="props" label="Disable Straightforward Quick Response" color="primary"
                            v-model="config.disable_direct_quick"></v-switch>
                </template>
              </v-tooltip>
              <!-- TODO bing filter bypass text -->
            </v-card-text>
          </v-card>
        </v-container>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>

</style>