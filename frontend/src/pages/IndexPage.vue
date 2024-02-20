<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from "vue"
import {main, sydney} from "../../wailsjs/go/models"
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"
import {fromChatMessages, generateRandomName, shadeColor, swal, toChatMessages} from "../helper"
import {AskAI, CountToken, GenerateImage, GetConciseAnswer} from "../../wailsjs/go/main/App"
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
import GenerativeImageWindow from "../components/index/GenerativeImageWindow.vue"
import AskOptions = main.AskOptions
import Workspace = main.Workspace
import ChatFinishResult = main.ChatFinishResult
import UploadSydneyImageResult = main.UploadSydneyImageResult
import GenerativeImage = sydney.GenerativeImage
import GenerateImageResult = sydney.GenerateImageResult
import ConciseAnswerReq = main.ConciseAnswerReq

let theme = useTheme()
let navDrawer = ref(true)
let modeList = ['Creative', 'Balanced', 'Precise', 'Designer']
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
  created_at: dayjs().format(),
  use_classic: false,
  gpt_4_turbo: false,
  persistent_input: false,
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
    'font-family': "'" + config.value.font_family + "'!important",
    'font-size': config.value.font_size + 'px!important',
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
let captchaDialog = ref(false)

let askEventMap = {
  "chat_append": (data: string) => {
    if (!replied.value) {
      fixContextLineBreak()
      currentWorkspace.value.context += '[user](#message)\n' +
          (hiddenPrompt.value === '' ? currentWorkspace.value.input : hiddenPrompt.value) + "\n\n"
      if (hiddenPrompt.value === '') {
        if (!currentWorkspace.value.persistent_input) {
          currentWorkspace.value.input = ''
        }
      } else {
        hiddenPrompt.value = ''
      }
      replied.value = true
    }
    currentWorkspace.value.context += data
    if (captchaDialog.value) {
      captchaDialog.value = false
    }
    scrollChatContextToBottom()
  },
  "chat_finish": (result: ChatFinishResult) => {
    console.log('receive chat_finish: ' + JSON.stringify(result))
    fixContextLineBreak()
    isAsking.value = false
    replied.value = false
    hiddenPrompt.value = ''
    if (captchaDialog.value) {
      captchaDialog.value = false
    }
    if (result.success) {
      statusBarText.value = 'Ready.'
      if (!config.value.no_image_removal_after_chat) {
        uploadedImage.value = undefined
      }
      lockScroll.value = false
      if (!config.value.disable_summary_title_generation) {
        generateTitle()
      }
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
  },
  "chat_resolving_captcha": (msg: string) => {
    captchaDialog.value = true
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
      if (!workspace.image_packs) {
        workspace.image_packs = []
      }
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
  currentWorkspace.value.image_packs = []
  suggestedResponses.value = []
}

let chatContextTabIndex = ref(0)

let generativeImageLoading = ref(false)

function generateImage(req: GenerativeImage) {
  generativeImageLoading.value = true
  GenerateImage(req).then(res => {
    if (!currentWorkspace.value.image_packs) {
      currentWorkspace.value.image_packs = []
    }
    currentWorkspace.value.image_packs.push(res)
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    generativeImageLoading.value = false
  })
}

let workspaceNav = ref(null)
let additionalOptionsDialog = ref(false)
let additionalOptionPreview = computed(() => {
  return 'Locale: ' + currentWorkspace.value.locale +
      '; No Search: ' + currentWorkspace.value.no_search +
      '; Use Classic: ' + currentWorkspace.value.use_classic
})

function generateTitle() {
  let workspace = currentWorkspace.value
  if (!/^Chat \w+_\w+$/m.test(workspace.title)) {
    return
  }
  let systemPrompt = '# Role: Title Generator\n' +
      '## Rules:\n' +
      '- Write an extremely concise subtitle for the text provided wrapped with <x-text> tag ' +
      'with no more than a few words.\n' +
      '- The first letter of all words should be capitalized.\n' +
      '- Exclude punctuation.\n' +
      '- Use the same langauge as the user\'s message.\n' +
      '- Write just the title and nothing else. No introduction to yourself. No explanation. Just the title.\n'
  let xContext = fromChatMessages(toChatMessages(workspace.context)
      .filter(v => !(v.role === 'system' && v.type === 'additional_instructions')))
  let req: ConciseAnswerReq
  if (workspace.backend === 'Sydney') {
    req = {
      backend: workspace.backend,
      context: '<x-text>\n' + xContext + '\n</x-text>',
      prompt: systemPrompt,
    }
  } else {
    req = {
      backend: workspace.backend,
      context: systemPrompt,
      prompt: '<x-text>\n' + xContext + '\n</x-text>',
    }
  }
  GetConciseAnswer(req).then(title => {
    workspace.title = title
        .replace(/^#/, '')
        .replace(/<\/?x-text>/g,'').trim()
  }).catch(err => {
    console.log(err)
  })
}
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
                     @update:suggested-responses="arr => suggestedResponses=arr"
                     @scroll-chat-context-to-bottom="scrollChatContextToBottom"></workspace-nav>
      <div class="d-flex flex-column fill-height" v-if="!loading">
        <div class="d-flex align-center top-action-bar mx-2">
          <p class="font-weight-bold">Chat Context:</p>
          <v-spacer></v-spacer>
          <div class="d-flex align-center">
            <v-select v-model="currentWorkspace.backend" :items="backendList" color="primary" label="Backend"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select v-model="currentWorkspace.conversation_style" :disabled="currentWorkspace.backend!=='Sydney'"
                      :items="modeList" color="primary" label="Mode"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select :model-value="currentWorkspace.preset" @update:model-value="onPresetChange"
                      :items="config.presets.map(v=>v.name)" color="primary"
                      label="Preset"
                      density="compact"
                      class="mx-2"></v-select>
            <v-tooltip
                text="Enable the latest gpt-4-turbo model will increase the speed of response,
                        reduce repeatability, but be harder to jailbreak."
                location="bottom">
              <template #activator="{props}">
                <v-switch v-bind="props" v-model="currentWorkspace.gpt_4_turbo" label="GPT-4-Turbo"
                          density="compact"
                          :disabled="currentWorkspace.backend!=='Sydney'" class="mx-2"
                          color="primary"></v-switch>
              </template>
            </v-tooltip>
            <v-tooltip :text="additionalOptionPreview" location="bottom">
              <template #activator="{props}">
                <v-btn @click="additionalOptionsDialog=true" v-bind="props" icon variant="text" color="primary">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-btn>
              </template>
            </v-tooltip>
            <v-dialog max-width="230" v-model="additionalOptionsDialog">
              <v-card title="Additional Options">
                <v-card-text>
                  <div class="d-flex flex-column">
                    <v-select v-model="currentWorkspace.locale" :disabled="currentWorkspace.backend!=='Sydney'"
                              :items="localeList" color="primary" label="Locale"
                              density="compact"></v-select>
                    <v-tooltip text="Note that you will not be able to generate images when No Search is enabled."
                               location="bottom">
                      <template #activator="{props}">
                        <v-switch v-bind="props" v-model="currentWorkspace.no_search" label="No Search"
                                  density="compact"
                                  :disabled="currentWorkspace.backend!=='Sydney'"
                                  color="primary"></v-switch>
                      </template>
                    </v-tooltip>
                    <v-tooltip text="Persist the user input so that it won't be cleared after sent." location="bottom">
                      <template #activator="{props}">
                        <v-switch v-bind="props" v-model="currentWorkspace.persistent_input" label="Persistent Input"
                                  density="compact" color="primary"></v-switch>
                      </template>
                    </v-tooltip>
                    <v-tooltip
                        text="Classic Creative mode will not enable gpt-4-turbo forever.
                          Turn off this to enable gpt-4-turbo for Creative mode when available."
                        location="bottom">
                      <template #activator="{props}">
                        <v-switch v-bind="props" v-model="currentWorkspace.use_classic" label="Use Classic Creative"
                                  density="compact"
                                  :disabled="currentWorkspace.backend!=='Sydney'"
                                  color="primary"></v-switch>
                      </template>
                    </v-tooltip>
                  </div>
                </v-card-text>
                <v-card-actions>
                  <v-spacer></v-spacer>
                  <v-btn variant="text" color="primary" @click="additionalOptionsDialog=false">Done</v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
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
          <v-tab :value="3">Image</v-tab>
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
            <v-window-item :value="3" class="fill-height">
              <generative-image-window :custom-font-style="customFontStyle"
                                       v-model:image-packs="currentWorkspace.image_packs"></generative-image-window>
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
          <v-tooltip text="There are images generating..." location="top">
            <template #activator="{props}">
              <v-scale-transition>
                <v-btn v-bind="props" icon v-if="generativeImageLoading"
                       style="position:absolute;left: 25px;bottom: 25px;" color="primary">
                  <img class="loading-icon"/>
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
        <v-dialog max-width="500" v-model="captchaDialog" :persistent="true">
          <v-card>
            <v-card-text>
              <div class="d-flex justify-center align-center flex-column">
                <v-progress-circular indeterminate color="primary"></v-progress-circular>
                <div class="mt-3">Please wait patiently while we are resolving the CAPTCHA...</div>
              </div>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn variant="text" color="primary" @click="captchaDialog=false">Dismiss</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
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