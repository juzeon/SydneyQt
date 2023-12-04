<script setup lang="ts">

import {v4 as uuidV4} from 'uuid'
import Scaffold from "../components/Scaffold.vue"
import {useRouter} from "vue-router"
import {useSettings} from "../composables"
import {computed, onMounted, ref} from "vue"
import {useTheme} from "vuetify"
import {main} from "../../wailsjs/go/models"
import Preset = main.Preset
import OpenAIBackend = main.OpenAIBackend

let theme = useTheme()
let router = useRouter()
let {config, fetch: fetchingSettings} = useSettings()
let loading = ref(true)
onMounted(() => {
  loading.value = true
  fetchingSettings().then(() => {
    loading.value = false
    activePreset.value = config.value.presets[0]
    activeOpenaiBackendName.value = config.value.open_ai_backends[0].name
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

let activePreset = ref<Preset>()

function addPreset() {
  let preset = <Preset>{
    name: 'New Preset ' + uuidV4(),
    content: '[system](#additional_instructions)\n',
  }
  config.value.presets.push(preset)
  activePreset.value = preset
}

function deletePreset(preset: Preset) {
  if (preset.name === 'Sydney') {
    return
  }
  config.value.presets = config.value.presets.filter(v => v.name !== preset.name)
  if (preset === activePreset.value) {
    activePreset.value = config.value.presets[0]
  }
}

let renamePresetName = ref('')
let renamePresetInstance = ref<Preset>()
let renamePresetDialog = ref(false)
let renamePresetError = ref('')

function renamePreset(preset: Preset) {
  if (preset.name === 'Sydney') {
    return
  }
  renamePresetInstance.value = preset
  renamePresetError.value = ''
  renamePresetName.value = preset.name
  renamePresetDialog.value = true
}

function confirmRenamePreset() {
  if (config.value.presets.find(v => v.name === renamePresetName.value)) {
    renamePresetError.value = 'Preset name already exists'
    return
  }
  renamePresetInstance.value!.name = renamePresetName.value
  renamePresetDialog.value = false
}

let activeOpenaiBackendName = ref('')
let renameBackendName = ref('')
let isRenamingBackend = ref(false)
let renameBackendError = ref('')

function onRenameBackend() {
  renameBackendError.value = ''
  renameBackendName.value = activeOpenaiBackendName.value
  isRenamingBackend.value = true
}

function confirmRenameBackend() {
  if (renameBackendName.value === '') {
    return
  }
  if (config.value.open_ai_backends.find(v => v.name === renameBackendName.value) || renameBackendName.value === 'Sydney') {
    renameBackendError.value = 'Backend name already exists.'
    return
  }
  config.value.open_ai_backends.find(v => v.name === activeOpenaiBackendName.value)!.name = renameBackendName.value
  activeOpenaiBackendName.value = renameBackendName.value
  isRenamingBackend.value = false
}

function addOpenaiBackend() {
  let backend = <OpenAIBackend>Object.assign({},
      config.value.open_ai_backends.find(v => v.name === activeOpenaiBackendName.value)!)
  backend.name = 'OpenAI ' + uuidV4().split('-')[0]
  config.value.open_ai_backends.push(backend)
  activeOpenaiBackendName.value = backend.name
}

function deleteOpenaiBackend(backend: OpenAIBackend) {
  if (config.value.open_ai_backends.length <= 1) {
    return
  }
  activeOpenaiBackendName.value = config.value.open_ai_backends[0].name
  config.value.open_ai_backends = config.value.open_ai_backends.filter(v => v !== backend)
}

function checkOpenaiEndpoint(val: string) {
  if (!val.endsWith('/v1')) {
    return 'The endpoint is expected to end with /v1'
  }
  if (!val.startsWith('http')) {
    return 'The endpoint is expected to start with http'
  }
  return true
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
              <v-card class="my-3">
                <v-card-title>Presets</v-card-title>
                <v-card-text>
                  <div class="d-flex" style="max-height: 400px">
                    <div class="d-flex flex-column">
                      <v-list density="compact" width="200" class="flex-grow-1 overflow-y-auto">
                        <v-list-item :active="preset===activePreset" v-for="preset in config.presets">
                          <template #title>
                            <p style="cursor: pointer" class="overflow-x-hidden" @click="activePreset=preset">
                              {{ preset.name }}</p>
                          </template>
                          <template #append>
                            <v-btn icon color="red" variant="text" density="compact" @click="deletePreset(preset)">
                              <v-icon>mdi-delete</v-icon>
                            </v-btn>
                            <v-btn icon color="primary" variant="text" density="compact" @click="renamePreset(preset)">
                              <v-icon>mdi-pencil</v-icon>
                            </v-btn>
                          </template>
                        </v-list-item>
                      </v-list>
                      <v-btn variant="text" color="primary" @click="addPreset">
                        <v-icon>mdi-plus</v-icon>
                        Add
                      </v-btn>
                    </div>
                    <v-textarea rows="15" class="ml-3" v-model="activePreset!.content"></v-textarea>
                  </div>
                </v-card-text>
                <v-dialog max-width="500" v-model="renamePresetDialog">
                  <v-card>
                    <v-card-title>Rename Preset</v-card-title>
                    <v-card-text>
                      <v-text-field label="Rename" :error-messages="renamePresetError" v-model="renamePresetName"
                                    color="primary"></v-text-field>
                    </v-card-text>
                    <v-card-actions>
                      <v-spacer></v-spacer>
                      <v-btn variant="text" color="primary" @click="renamePresetDialog=false">Cancel</v-btn>
                      <v-btn variant="text" color="primary" @click="confirmRenamePreset">Confirm</v-btn>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-card>
              <v-card class="my-3">
                <v-card-title>OpenAI Backends</v-card-title>
                <v-card-subtitle>Allow you to invoke OpenAI ChatGPT API.
                  You can set multiple configurations and choose a backend from the chat page.
                </v-card-subtitle>
                <v-card-text>
                  <div class="d-flex">
                    <v-tabs class="flex-grow-1" show-arrows v-model="activeOpenaiBackendName" color="primary"
                            :disabled="isRenamingBackend">
                      <v-tab v-for="backend in config.open_ai_backends" :value="backend.name">{{ backend.name }}</v-tab>
                    </v-tabs>
                    <v-btn variant="text" icon color="primary" @click="addOpenaiBackend" :disabled="isRenamingBackend">
                      <v-icon>mdi-plus</v-icon>
                    </v-btn>
                  </div>
                  <v-window v-model="activeOpenaiBackendName">
                    <v-window-item v-for="backend in config.open_ai_backends" :value="backend.name">
                      <div class="mx-3 my-3">
                        <div class="d-flex align-center">
                          <v-text-field v-if="isRenamingBackend" label="Name" v-model="renameBackendName"
                                        color="primary" :error-messages="renameBackendError"></v-text-field>
                          <v-text-field v-else label="Name" :model-value="backend.name" color="primary"
                                        :disabled="true"></v-text-field>
                          <div v-if="isRenamingBackend" class="d-flex mx-3 mb-3">
                            <v-btn icon color="primary" variant="text" class="mx-1" @click="confirmRenameBackend">
                              <v-icon>mdi-check</v-icon>
                            </v-btn>
                            <v-btn icon color="primary" variant="text" class="mx-1" @click="isRenamingBackend=false">
                              <v-icon>mdi-close</v-icon>
                            </v-btn>
                          </div>
                          <v-btn v-else icon color="primary" variant="text" class="mx-3 mb-3" @click="onRenameBackend">
                            <v-icon>mdi-pencil</v-icon>
                          </v-btn>
                        </div>
                        <v-text-field label="Endpoint" :rules="[checkOpenaiEndpoint]" v-model="backend.openai_endpoint"
                                      color="primary"></v-text-field>
                        <v-text-field label="Key" v-model="backend.openai_key" color="primary"></v-text-field>
                        <v-text-field label="Model" v-model="backend.openai_short_model" color="primary"></v-text-field>
                        <v-slider label="Temperature" v-model="backend.openai_temperature" min="0" max="2"
                                  step="0.1" color="primary" thumb-label="always"></v-slider>
                        <div class="d-flex my-3">
                          <v-spacer></v-spacer>
                          <v-btn icon color="red" variant="text" :disabled="isRenamingBackend"
                                 @click="deleteOpenaiBackend(backend)">
                            <v-icon>mdi-delete</v-icon>
                          </v-btn>
                        </div>
                      </div>
                    </v-window-item>
                  </v-window>
                </v-card-text>
              </v-card>
            </v-card-text>
          </v-card>
        </v-container>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>

</style>