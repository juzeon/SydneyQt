<script setup lang="ts">

import Scaffold from "../components/Scaffold.vue"
import {useRouter} from "vue-router"
import {useSettings} from "../composables"
import {computed, onMounted, ref} from "vue"
import {useTheme} from "vuetify"
import {shadeColor} from "../helper"
import UpdateCard from "../components/settings/UpdateCard.vue"
import OpenAIBackendsCard from "../components/settings/OpenAIBackendCard.vue"
import PresetCard from "../components/settings/PresetCard.vue"

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

let quickRespEditMode = ref('create')
let quickRespEditText = ref('')
let quickRespEditDialog = ref(false)
let quickRespEditError = ref('')
let quickRespEditIndex = ref(-1)

function moveQuickResponse(item: string, isUp: boolean) {
  let index = config.value.quick.findIndex(v => v === item)
  if (index === -1 || config.value.quick.length <= 1) {
    return
  }
  if (isUp) {
    if (index === 0) return
    let tmp = config.value.quick[index]
    config.value.quick[index] = config.value.quick[index - 1]
    config.value.quick[index - 1] = tmp
  } else {
    if (index === config.value.quick.length - 1) return
    let tmp = config.value.quick[index]
    config.value.quick[index] = config.value.quick[index + 1]
    config.value.quick[index + 1] = tmp
  }
}

function createQuickResponse() {
  quickRespEditMode.value = 'create'
  quickRespEditText.value = ''
  quickRespEditDialog.value = true
  quickRespEditError.value = ''
}

function editQuickResponse(index: number) {
  quickRespEditMode.value = 'edit'
  quickRespEditText.value = config.value.quick[index]
  quickRespEditIndex.value = index
  quickRespEditDialog.value = true
  quickRespEditError.value = ''
}

function confirmQuickResponse() {
  if (config.value.quick.find(v => v === quickRespEditText.value)) {
    quickRespEditError.value = 'This Quick Response already exists.'
    return
  }
  if (quickRespEditMode.value === 'create') {
    config.value.quick.push(quickRespEditText.value.trim())
  } else {
    config.value.quick[quickRespEditIndex.value] = quickRespEditText.value.trim()
  }
  quickRespEditDialog.value = false
}

function onChangeThemeColor(val: string) {
  if (!checkThemeColor(val)) {
    return
  }
  config.value.theme_color = val
  theme.themes.value.light.colors.primary = val
  theme.themes.value.dark.colors.primary = shadeColor(val, -40)
}

function checkThemeColor(val: string): boolean {
  return /^#[0-9A-Fa-f]{6}$/i.test(val)
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
              <v-text-field label="Theme Color" color="primary" :model-value="config.theme_color"
                            @update:model-value="onChangeThemeColor"
                            hint="Must be a valid hex color; default value: #FF9800"></v-text-field>
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
              <v-card class="my-3">
                <v-card-title>Quick Responses</v-card-title>
                <v-card-text>
                  <v-list density="compact">
                    <v-list-item v-for="(item,ix) in config.quick" :title="item">
                      <template #prepend>
                        <p style="color: #999" class="mr-2 text-no-wrap overflow-x-hidden">{{ ix + 1 }}.</p>
                      </template>
                      <template #append>
                        <v-btn variant="text" icon density="compact" class="mx-1" color="red"
                               @click="config.quick=config.quick.filter(v=>v!==item)">
                          <v-icon>mdi-delete</v-icon>
                        </v-btn>
                        <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                               @click="editQuickResponse(ix)">
                          <v-icon>mdi-pencil</v-icon>
                        </v-btn>
                        <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                               @click="moveQuickResponse(item,true)">
                          <v-icon>mdi-menu-up</v-icon>
                        </v-btn>
                        <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                               @click="moveQuickResponse(item,false)">
                          <v-icon>mdi-menu-down</v-icon>
                        </v-btn>
                      </template>
                    </v-list-item>
                  </v-list>
                  <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn variant="text" prepend-icon="mdi-plus" color="primary" @click="createQuickResponse">
                      Add
                    </v-btn>
                  </v-card-actions>
                </v-card-text>
                <v-dialog max-width="500" v-model="quickRespEditDialog">
                  <v-card :title="quickRespEditMode==='create'?'Create a Quick Response':'Edit the Quick Response'">
                    <v-card-text>
                      <v-text-field :error-messages="quickRespEditError" label="Quick Response"
                                    v-model="quickRespEditText"
                                    color="primary"></v-text-field>
                    </v-card-text>
                    <v-card-actions>
                      <v-spacer></v-spacer>
                      <v-btn variant="text" color="primary" @click="quickRespEditDialog=false">Cancel</v-btn>
                      <v-btn variant="text" color="primary" @click="confirmQuickResponse">Confirm</v-btn>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-card>
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