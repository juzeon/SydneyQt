<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from "vue"
import {main, sydney} from "../../wailsjs/go/models"
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"
import {generateRandomName, shadeColor, swal} from "../helper"
import {AskAI, CountToken, GenerateImage} from "../../wailsjs/go/main/App"
import {AskTypeOpenAI, AskTypeSydney} from "../constants"
import Scaffold from "../components/Scaffold.vue"
import {useSettings} from "../composables"
import {useTheme} from "vuetify"
import dayjs from "dayjs"
import RichChatContext from "../components/index/RichChatContext.vue"
import UserStatusButton from "../components/index/UserStatusButton.vue"
import WorkspaceNav from "../components/index/WorkspaceNav.vue"
import UploadImageButton from "../components/index/UploadImageButton.vue"
import UploadDocumentButton from "../components/index/UploadDocumentButton.vue"
import FetchWebpageButton from "../components/index/FetchWebpageButton.vue"
import RevokeButton from "../components/index/RevokeButton.vue"
import AskOptions = main.AskOptions
import Workspace = main.Workspace
import ChatFinishResult = main.ChatFinishResult
import UploadSydneyImageResult = main.UploadSydneyImageResult
import GenerativeImage = sydney.GenerativeImage
import GenerateImageResult = sydney.GenerateImageResult

let theme = useTheme()
let navDrawer = ref(true)
let modeList = ['Creative', 'Balanced', 'Precise']
let backendList = computed(() => {
  return ['Sydney', ...config.value.open_ai_backends.map(v => v.name)]
})
let localeList = ['zh-CN', 'en-US']
let loading = ref(true)
let currentWorkspace = ref(<Workspace>{
  id: 1,
  title: 'Chat ' + generateRandomName(),
  context: '',
  input: '',
  backend: 'Sydney',
  locale: 'zh-CN',
  preset: 'Sydney',
  conversation_style: 'Creative',
  no_search: false,
  image_packs: <GenerateImageResult[]>[],
  created_at: dayjs().format()
})

let chatContextTokenCount = ref(0)
let userInputTokenCount = ref(0)
let fetchingTokenCount = ref(0)
watch(currentWorkspace, async () => {
  chatContextTokenCount.value = await CountToken(currentWorkspace.value.context)
  userInputTokenCount.value = await CountToken(currentWorkspace.value.input)
  config.value.current_workspace_id = currentWorkspace.value.id
}, {deep: true})
let statusTokenCountText = computed(() => {
  return 'Chat Context: ' + chatContextTokenCount.value + ' tokens; User Input: ' + userInputTokenCount.value + ' tokens'
})
let statusBarText = ref('Ready.')
let {config, fetch: fetchSettings} = useSettings()
let customFontStyle = computed(() => {
  return {
    'font-family': "'" + config.value.font_family + "'",
    'font-size': config.value.font_size + 'px',
  }
})

let hiddenPrompt = ref('')
let replyDeep = ref(0)
watch(hiddenPrompt, value => {
  console.log('hiddenPrompt changed: ' + value)
})
let suggestedResponses = ref<string[]>([])
let isAsking = ref(false)
let replied = ref(false)
let lockScroll = ref(false)

let askEventMap = {
  "chat_append": (data: string) => {
    if (!replied.value) {
      fixContextLineBreak()
      currentWorkspace.value.context += '[user](#message)\n' +
          (hiddenPrompt.value === '' ? currentWorkspace.value.input : hiddenPrompt.value) + "\n\n"
      if (hiddenPrompt.value === '') {
        currentWorkspace.value.input = ''
      } else {
        hiddenPrompt.value = ''
      }
      replied.value = true
    }
    currentWorkspace.value.context += data
    scrollChatContextToBottom()
  },
  "chat_finish": (result: ChatFinishResult) => {
    console.log('receive chat_finish: ' + JSON.stringify(result))
    fixContextLineBreak()
    isAsking.value = false
    replied.value = false
    hiddenPrompt.value = ''
    if (result.success) {
      statusBarText.value = 'Ready.'
      if (!config.value.no_image_removal_after_chat) {
        uploadedImage.value = undefined
      }
      lockScroll.value = false
    } else {
      console.log('error type: ' + result.err_type)
      console.log('error message: ' + result.err_msg)
      statusBarText.value = result.err_msg
      switch (result.err_type) {
        case 'others':
        case 'message_filtered':
          // should first check the user input, if existed, append to the chat context
          swal.error(result.err_msg)
          statusBarText.value = result.err_msg
          break
        case 'message_revoke':
          if (config.value.revoke_reply_text !== '' && replyDeep.value < config.value.revoke_reply_count) {
            statusBarText.value = ''
            startAsking({
              prompt: config.value.revoke_reply_text,
              replyDeep: replyDeep.value + 1,
              statusBarText: 'Recreating the conversation with Revoke Reply Text automatically...'
            })
          } else {
            swal.error(result.err_msg)
            if (config.value.revoke_reply_text !== '') {
              suggestedResponses.value = [config.value.revoke_reply_text]
            }
          }
          break
      }
    }
  },
  "chat_suggested_responses": (data: string) => {
    suggestedResponses.value = JSON.parse(data)
  },
  "chat_token": (data: number) => {
    fetchingTokenCount.value = data
    statusBarText.value = 'Fetching the response, ' + fetchingTokenCount.value + ' tokens received currently.'
  },
  "chat_conversation_created": () => {
    statusBarText.value = 'Fetching the response...'
  },
  "chat_generate_image": (req: GenerativeImage) => {
    generateImage(req)
  }
}

function scrollChatContextToBottom() {
  if (lockScroll.value) {
    return
  }
  setTimeout(() => {
    let element = document.getElementById('chat-context')
    if (element) {
      element.scrollTop = element.scrollHeight
    }
  }, 0)
}

function fixContextLineBreak() {
  if (currentWorkspace.value.context.trim() == '') {
    return
  }
  if (!currentWorkspace.value.context.endsWith("\n\n")) {
    if (currentWorkspace.value.context.endsWith("\n")) {
      currentWorkspace.value.context += "\n"
    } else {
      currentWorkspace.value.context += "\n\n"
    }
  }
}

function doListeningEvents(isUnregister: boolean = false) {
  isUnregister ? console.log('unregister chat listener') : console.log('register chat listener')
  for (let [event, func] of Object.entries(askEventMap)) {
    if (isUnregister) {
      EventsOff(event)
    } else {
      EventsOn(event, func)
    }
  }
}

interface StartAskingArgs {
  prompt?: string,
  replyDeep?: number,
  statusBarText?: string
}

async function startAsking(args: StartAskingArgs = {}) {
  if (isAsking.value) {
    swal.error('An active conversation has already launched.')
    return
  }
  console.log('startAsking is called with: ' + JSON.stringify(args))
  suggestedResponses.value = []
  isAsking.value = true
  statusBarText.value = args.statusBarText ? args.statusBarText : 'Creating the conversation...'
  let askOptions = new AskOptions()
  askOptions.chat_context = currentWorkspace.value.context
  askOptions.type = currentWorkspace.value.backend === 'Sydney' ? AskTypeSydney : AskTypeOpenAI
  if (!args.prompt) {
    askOptions.prompt = currentWorkspace.value.input
  } else {
    hiddenPrompt.value = args.prompt
    askOptions.prompt = hiddenPrompt.value
  }
  replyDeep.value = args.replyDeep !== undefined ? args.replyDeep : 0
  askOptions.openai_backend = currentWorkspace.value.backend
  askOptions.image_url = uploadedImage.value?.bing_url ?? ''
  await AskAI(askOptions)
}

function applyQuickResponse(text: string) {
  if (currentWorkspace.value.input.trim() === '' && !config.value.disable_direct_quick) {
    startAsking({prompt: text, statusBarText: 'Creating the conversation with: ' + text})
    return
  }
  if (!currentWorkspace.value.input.endsWith('\n') && currentWorkspace.value.input.trim() !== '') {
    currentWorkspace.value.input += '\n'
  }
  currentWorkspace.value.input += text
}

function stopAsking() {
  EventsEmit('chat_stop')
}

let uploadedImage = ref<UploadSydneyImageResult | undefined>()

function handleKeyPress(event: KeyboardEvent) {
  if (document.getElementById('user-input') !== document.activeElement) {
    return
  }
  if (isAsking.value) {
    return
  }
  if (config.value.enter_mode === 'Enter' && (event.key == 'Enter' || event.key == 'NumpadEnter')) {
    if (!event.shiftKey) {
      event.preventDefault()
      startAsking()
    }
  } else if ((event.keyCode == 10 || event.keyCode == 13) && (event.ctrlKey || event.metaKey)) {
    startAsking()
  }
}

function appendBlockToCurrentWorkspace(blockText: string) {
  fixContextLineBreak()
  currentWorkspace.value.context += blockText
  scrollChatContextToBottom()
}

onMounted(() => {
  loading.value = true
  doListeningEvents()
  fetchSettings().then(async () => {
    theme.themes.value.light.colors.primary = config.value.theme_color
    theme.themes.value.dark.colors.primary = shadeColor(config.value.theme_color, -40)
    theme.global.name.value = config.value.dark_mode ? 'dark' : 'light'
    let workspace = config.value.workspaces?.find(v => v.id === config.value.current_workspace_id)
    if (workspace) {
      currentWorkspace.value = workspace
    } else {
      currentWorkspace.value.context = config.value.presets.find(v => v.name === 'Sydney')?.content ?? ''
      config.value.workspaces = [currentWorkspace.value]
      config.value.current_workspace_id = 1
    }
    chatContextTokenCount.value = await CountToken(currentWorkspace.value.context)
    loading.value = false
    setTimeout(() => {
      scrollChatContextToBottom()
    }, 0)
  })
  window.addEventListener('keypress', handleKeyPress, true)
})
onUnmounted(() => {
  doListeningEvents(true)
  window.removeEventListener('keypress', handleKeyPress, true)
})

function onPresetChange(newValue: string) {
  if (currentWorkspace.value.context.trim()
      === config.value.presets.find(v => v.name === currentWorkspace.value.preset)?.content.trim()) {
    currentWorkspace.value.context = config.value.presets.find(v => v.name === newValue)?.content ?? ''
  }
  currentWorkspace.value.preset = newValue
  suggestedResponses.value = []
}

function onReset() {
  currentWorkspace.value.context = config.value.presets.find(v => v.name === currentWorkspace.value.preset)?.content ?? ''
  suggestedResponses.value = []
}

let chatContextTabIndex = ref(0)

let generativeImageLoading = ref(false)
let generativeImageError = ref('')

function generateImage(req: GenerativeImage) {
  generativeImageLoading.value = true
  generativeImageError.value = ''
  GenerateImage(req).then(res => {
    if (!currentWorkspace.value.image_packs) {
      currentWorkspace.value.image_packs = []
    }
    currentWorkspace.value.image_packs.push(res)
  }).catch(err => {
    generativeImageError.value = err.toString()
  }).finally(() => {
    generativeImageLoading.value = false
  })
}

let workspaceNav = ref(null)

</script>

<template>
  <scaffold>
    <template #left-top>
      <v-btn icon @click="navDrawer=!navDrawer">
        <v-icon>mdi-menu</v-icon>
      </v-btn>
    </template>
    <template #right-top-prepend>
      <user-status-button></user-status-button>
    </template>
    <template #default>
      <workspace-nav v-if="!loading" :is-asking="isAsking" v-model="navDrawer"
                     v-model:current-workspace="currentWorkspace"
                     v-model:workspaces="config.workspaces" :presets="config.presets" @on-reset="onReset"
                     @update:suggested-responses="arr => suggestedResponses=arr"></workspace-nav>
      <div class="d-flex flex-column fill-height" v-if="!loading">
        <div class="d-flex align-center top-action-bar mx-2">
          <p class="font-weight-bold">Chat Context:</p>
          <v-spacer></v-spacer>
          <div class="d-flex">
            <v-select v-model="currentWorkspace.backend" :items="backendList" color="primary" label="Backend"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select v-model="currentWorkspace.conversation_style" :disabled="currentWorkspace.backend!=='Sydney'"
                      :items="modeList" color="primary" label="Mode"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select v-model="currentWorkspace.locale" :disabled="currentWorkspace.backend!=='Sydney'"
                      :items="localeList" color="primary" label="Locale"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select :model-value="currentWorkspace.preset" @update:model-value="onPresetChange"
                      :items="config.presets.map(v=>v.name)" color="primary"
                      label="Preset"
                      density="compact"
                      class="mx-2"></v-select>
            <v-switch v-model="currentWorkspace.no_search" label="No Search" density="compact"
                      :disabled="currentWorkspace.backend!=='Sydney'"
                      color="primary" class="mx-2 mt-1"></v-switch>
          </div>
          <v-btn class="ml-2" variant="tonal" :disabled="isAsking" color="primary"
                 @click="onReset">
            <v-icon>mdi-reload</v-icon>
            Reset
          </v-btn>
        </div>
        <v-tabs v-model="chatContextTabIndex" density="compact" color="primary" class="mb-1 flex-shrink-0">
          <v-tab :value="0">Plain</v-tab>
          <v-tab :value="1">Rich</v-tab>
        </v-tabs>
        <div class="flex-grow-1" style="min-height: 0;position: relative"><!-- This is to enable the scroll bar -->
          <v-window v-model="chatContextTabIndex" class="fill-height">
            <v-window-item :value="0" class="fill-height">
              <textarea :style="customFontStyle" id="chat-context" class="input-textarea"
                        v-model="currentWorkspace.context"></textarea>
            </v-window-item>
            <v-window-item :value="1" class="fill-height">
              <rich-chat-context :lock-scroll="lockScroll" :custom-font-style="customFontStyle"
                                 :context="currentWorkspace.context"></rich-chat-context>
            </v-window-item>
          </v-window>
          <v-tooltip :text="lockScroll?'Enable Auto Scrolling':'Disable Auto Scrolling'" location="top">
            <template #activator="{props}">
              <v-scale-transition>
                <v-btn v-bind="props" icon style="position:absolute;right: 25px;bottom: 25px;"
                       v-if="isAsking" @click="lockScroll=!lockScroll"
                       color="primary">
                  <v-icon v-if="lockScroll">mdi-transfer-down</v-icon>
                  <v-icon v-else>mdi-arrow-vertical-lock</v-icon>
                </v-btn>
              </v-scale-transition>
            </template>
          </v-tooltip>
        </div>
        <div class="d-flex" v-if="!config.no_suggestion">
          <div style="font-size: 12px;height: 20px;margin-top: 2px" class="overflow-x-hidden text-no-wrap">
            <v-chip style="cursor: pointer" v-for="item in suggestedResponses" density="compact" color="primary"
                    variant="outlined" @click="startAsking({prompt:item})"
                    class="ml-3">{{ item }}
            </v-chip>
          </div>
        </div>
        <div class="my-1 d-flex">
          <p class="font-weight-bold">Follow-up User Input:</p>
          <v-spacer></v-spacer>
          <upload-image-button :is-asking="isAsking" v-model="uploadedImage"></upload-image-button>
          <upload-document-button :is-asking="isAsking"
                                  @append-block-to-current-workspace="appendBlockToCurrentWorkspace"
          ></upload-document-button>
          <fetch-webpage-button :is-asking="isAsking"
                                @append-block-to-current-workspace="appendBlockToCurrentWorkspace"></fetch-webpage-button>
          <revoke-button :is-asking="isAsking" :current-workspace="currentWorkspace"></revoke-button>
          <v-menu>
            <template #activator="{props}">
              <v-btn color="primary" density="compact" variant="tonal" append-icon="mdi-menu-down"
                     v-bind="props" class="mx-1"
                     :disabled="isAsking">
                Quick
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item density="compact" @click="applyQuickResponse(item)" v-for="item in config.quick">{{
                  item
                }}
              </v-list-item>
            </v-list>
          </v-menu>
          <v-btn color="primary" density="compact" variant="tonal" class="mx-1" v-if="isAsking" @click="stopAsking"
                 append-icon="mdi-stop">Stop
          </v-btn>
          <v-btn color="primary" density="compact" variant="tonal" class="mx-1" v-else @click="startAsking()"
                 append-icon="mdi-send">
            Send
          </v-btn>
        </div>
        <div :style="{height:config.stretch_factor+'vh'}" class="flex-shrink-0">
          <textarea :style="customFontStyle" id="user-input" class="input-textarea"
                    v-model="currentWorkspace.input"></textarea>
        </div>
        <div class="d-flex text-caption">
          <p class="overflow-hidden text-no-wrap">{{ statusBarText }}</p>
          <v-spacer></v-spacer>
          <p class="text-no-wrap ml-2">{{ statusTokenCountText }}</p>
        </div>
      </div>
    </template>
  </scaffold>
</template>

<style scoped>
.input-textarea {
  height: 99%;
  width: 100%;
  border: grey 1px solid;
  resize: none;
  padding: 5px;
}
</style>