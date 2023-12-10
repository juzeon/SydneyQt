<script setup lang="ts">

import Scaffold from "../components/Scaffold.vue"
import {useRouter} from "vue-router"
import {useSettings} from "../composables"
import {computed, onMounted, ref} from "vue"
import UpdateCard from "../components/settings/UpdateCard.vue"
import OpenAIBackendsCard from "../components/settings/OpenAIBackendCard.vue"
import PresetCard from "../components/settings/PresetCard.vue"
import QuickResponseCard from "../components/settings/QuickResponseCard.vue"
import ThemeTextField from "../components/settings/ThemeTextField.vue"
import {useTheme} from "vuetify"

let theme = useTheme()
let router = useRouter()
let {config, fetch: fetchingSettings} = useSettings()
let loading = ref(true)
onMounted(() => {
  loading.value = true
  fetchingSettings().then(() => {
    console.log('get settings ok')
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
      <div v-if="!loading" class="fill-height overflow-y-auto">
        <v-container class="d-flex flex-column">
          <p class="text-h4 mb-3">Settings</p>
          <v-card title="Application" class="my-3">
            <v-card-text>
              <update-card></update-card>
            </v-card-text>
          </v-card>
          <v-card title="Network" class="my-3">
            <v-card-text>
              <v-tooltip
                  text="Enter a proxy URL in http, https or socks5 (e.g. http://127.0.0.1:7890). Leave blank to disable proxy."
                  location="bottom">
                <template #activator="{props}">
                  <v-text-field color="primary" label="Proxy" v-model="config.proxy" v-bind="props"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip
                  text="FQDN for the websocket endpoint of Sydney, without any prefixes (protocol names) or suffixes (slashes)."
                  location="top">
                <template #activator="{props}">
                  <v-text-field color="primary" label="Wss Domain" v-model="config.wss_domain"
                                v-bind="props" hint="Default: sydney.bing.com"></v-text-field>
                </template>
              </v-tooltip>
              <v-tooltip text="URL for creating the Bing conversation and gaining the conversation id."
                         location="bottom">
                <template #activator="{props}">
                  <v-autocomplete color="primary" label="Create Conversation URL" v-bind="props"
                                  :items="['https://edgeservices.bing.com/edgesvc/turing/conversation/create',
                        'https://www.bing.com/turing/conversation/create',
                        'https://copilot.microsoft.com/turing/conversation/create']"
                                  v-model="config.create_conversation_url"></v-autocomplete>
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
                                v-model="config.font_family" hint="Default: SF"></v-text-field>
                </template>
              </v-tooltip>
              <v-slider color="primary" min="10" max="30" step="1" thumb-label="always" v-model="config.font_size"
                        label="Font Size" hint="Default: 16"></v-slider>
              <div class="text-center mb-3" :style="fontStyle">Font Example</div>
              <v-tooltip text="Height of the textarea of user input, in vh." location="bottom">
                <template #activator="{props}">
                  <v-slider color="primary" v-bind="props" step="1" min="10" max="60" label="User Input Textarea Height"
                            v-model="config.stretch_factor"
                            thumb-label="always" hint="Default: 20"></v-slider>
                </template>
              </v-tooltip>
              <theme-text-field v-model:theme-color="config.theme_color"></theme-text-field>
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
              <!-- TODO bing filter bypass text -->
              <v-tooltip text="Whether to send the selected quick response immediately
              rather than append it to the user input textarea if it is empty." location="bottom">
                <template #activator="{props}">
                  <v-switch v-bind="props" label="Disable Straightforward Quick Response" color="primary"
                            v-model="config.disable_direct_quick"></v-switch>
                </template>
              </v-tooltip>
              <v-tooltip text="Whether to remove the uploaded image after successfully receiving Bing's response."
                         location="bottom">
                <template #activator="{props}">
                  <v-switch v-bind="props" label="No Image Removal After Chat" color="primary"
                            v-model="config.no_image_removal_after_chat"></v-switch>
                </template>
              </v-tooltip>
            </v-card-text>
          </v-card>
          <v-card title="Templates" class="my-3">
            <v-card-text>
              <quick-response-card v-model:quick="config.quick"></quick-response-card>
              <preset-card v-model:presets="config.presets"></preset-card>
              <OpenAIBackendsCard v-model:open_ai_backends="config.open_ai_backends"></OpenAIBackendsCard>
            </v-card-text>
          </v-card>
        </v-container>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>

</style>