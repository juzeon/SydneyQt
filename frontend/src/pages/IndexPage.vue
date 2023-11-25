<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from "vue"
import {main} from "../../wailsjs/go/models"
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"
import {swal} from "../helper"
import {AskAI, CountToken} from "../../wailsjs/go/main/App"
import {AskTypeOpenAI, AskTypeSydney} from "../constants"
import Scaffold from "../components/Scaffold.vue"
import Conversation from "../components/Conversation.vue"
import {useSettings} from "../composables"
import {useTheme} from "vuetify"
import UserInputToolButton from "../components/UserInputToolButton.vue"
import AskOptions = main.AskOptions
import Workspace = main.Workspace

let theme = useTheme()
let navDrawer = ref(false)
let modeList = ['Creative', 'Balanced', 'Precise']
let backendList = computed(() => {
  return ['Sydney', ...config.value.open_ai_backends.map(v => v.name)]
})
let localeList = ['zh-CN', 'en-US']
let loading = ref(true)
let currentWorkspace = ref(<Workspace>{
  id: 1,
  context: '',
  input: '',
  backend: 'Sydney',
  locale: 'zh-CN',
  preset: 'Sydney',
  conversation_style: 'Creative',
  no_search: false,
})
let chatContextTokenCount = ref(0)
let userInputTokenCount = ref(0)
let fetchingTokenCount = ref(0)
watch(currentWorkspace, async () => {
  chatContextTokenCount.value = await CountToken(currentWorkspace.value.context)
  userInputTokenCount.value = await CountToken(currentWorkspace.value.input)
}, {deep: true})
let statusTokenCountText = computed(() => {
  return 'Chat Context: ' + chatContextTokenCount.value + ' tokens; User Input: ' + userInputTokenCount.value + ' tokens'
})
let statusBarText = ref('Ready.')
let {config, fetch: fetchSettings} = useSettings()
let textareaStyle = computed(() => {
  return {
    'font-family': "'" + config.value.font_family + "'",
    'font-size': config.value.font_size + 'px',
  }
})

let hiddenPrompt = ref('')
watch(hiddenPrompt, value => {
  console.log('hiddenPrompt changed: ' + value)
})
let suggestedResponses = ref<string[]>([])
let isAsking = ref(false)
let replied = ref(false)

let askEventMap = {
  "chat_alert": (data: string) => {
    swal.error(data)
    statusBarText.value = data
  },
  "chat_append": (data: string) => {
    let scrollBottom = false
    if (!replied.value) {
      console.log('first reply')
      fixContextLineBreak()
      currentWorkspace.value.context += '[user](#message)\n' +
          (hiddenPrompt.value === '' ? currentWorkspace.value.input : hiddenPrompt.value) + "\n\n"
      scrollBottom = true
      if (hiddenPrompt.value === '') {
        currentWorkspace.value.input = ''
      } else {
        hiddenPrompt.value = ''
      }
      replied.value = true
    }
    let chatContextElem = document.getElementById('chat-context')
    if (chatContextElem) {
      // if (Math.abs(chatContextElem.scrollTop - chatContextElem.scrollHeight) < 500) {
      scrollBottom = true
      // }
      currentWorkspace.value.context += data
      if (scrollBottom) {
        chatContextElem.scrollTop = chatContextElem.scrollHeight
      }
    }
  },
  "chat_finish": (success: boolean) => {
    fixContextLineBreak()
    doListeningEvents(true)
    isAsking.value = false
    replied.value = false
    hiddenPrompt.value = ''
    if (success) {
      statusBarText.value = 'Ready.'
    }
  },
  "chat_suggested_responses": (data: string) => {
    suggestedResponses.value = JSON.parse(data)
  },
  "chat_token": (data: number) => {
    fetchingTokenCount.value = data
    statusBarText.value = 'Fetching the response, ' + fetchingTokenCount.value + ' tokens received currently.'
  },
  "chat_message_revoke": (replyDeep: number) => {
    if (config.value.revoke_reply_text != '' && replyDeep < config.value.revoke_reply_count) {
      // TODO send continue instruction
      // should first check the user input, if exist, append to the chat context
    }
  },
  "chat_conversation_created": () => {
    statusBarText.value = 'Fetching the response...'
  }
}

function scrollChatContextToBottom() {
  let element = document.getElementById('chat-context')
  if (element) {
    element.scrollTop = element.scrollHeight
  }
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
  for (let [event, func] of Object.entries(askEventMap)) {
    if (isUnregister) {
      EventsOff(event)
    } else {
      EventsOn(event, func)
    }
  }
}

async function startAsking(prompt: string = '') {
  if (isAsking.value) {
    swal.error('An active conversation has already launched.')
    return
  }
  suggestedResponses.value = []
  isAsking.value = true
  statusBarText.value = 'Creating the conversation...'
  doListeningEvents()
  let askOptions = new AskOptions()
  askOptions.chat_context = currentWorkspace.value.context
  askOptions.type = currentWorkspace.value.backend === 'Sydney' ? AskTypeSydney : AskTypeOpenAI
  if (prompt === '') {
    askOptions.prompt = currentWorkspace.value.input
  } else {
    hiddenPrompt.value = prompt
    askOptions.prompt = hiddenPrompt.value
  }
  askOptions.reply_deep = 0
  askOptions.openai_backend = ''
  askOptions.image_url = ''
  await AskAI(askOptions)
}

function stopAsking() {
  EventsEmit('chat_stop')
  isAsking.value = false
}

function handleKeyPress(event: KeyboardEvent) {
  if (document.getElementById('user-input') !== document.activeElement) {
    return
  }
  if (isAsking.value) {
    return
  }
  console.log('handle focused key press for user-input')
  if (config.value.enter_mode === 'Enter' && (event.key == 'Enter' || event.key == 'NumpadEnter')) {
    if (!event.shiftKey) {
      event.preventDefault()
      startAsking()
    }
  } else if ((event.keyCode == 10 || event.keyCode == 13) && (event.ctrlKey || event.metaKey)) {
    startAsking()
  }
}

onMounted(() => {
  loading.value = true
  fetchSettings().then(async () => {
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
  window.removeEventListener('keypress', handleKeyPress, true)
})

function onPresetChange(newValue: string) {
  if (currentWorkspace.value.context.trim()
      === config.value.presets.find(v => v.name === currentWorkspace.value.preset)?.content.trim()) {
    currentWorkspace.value.context = config.value.presets.find(v => v.name === newValue)?.content ?? ''
  }
  currentWorkspace.value.preset = newValue
}

</script>

<template>
  <scaffold>
    <template #left-top>
      <v-btn icon @click="navDrawer=!navDrawer">
        <v-icon>mdi-menu</v-icon>
      </v-btn>
    </template>
    <template #default>
      <v-navigation-drawer v-model="navDrawer">
        <v-list>
          <v-list-item>
            <conversation :id="1" title="New Chat" :created-at="new Date()"></conversation>
          </v-list-item>
        </v-list>
      </v-navigation-drawer>
      <div style="height: 100%" class="d-flex flex-column" v-if="!loading">
        <div class="d-flex align-center">
          <p class="mb-5 font-weight-bold">Chat Context:</p>
          <v-spacer></v-spacer>
          <div class="d-flex">
            <v-select v-model="currentWorkspace.backend" :items="backendList" color="primary" label="Backend"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select v-model="currentWorkspace.conversation_style" :items="modeList" color="primary" label="Mode"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select v-model="currentWorkspace.locale" :items="localeList" color="primary" label="Locale"
                      density="compact"
                      class="mx-2"></v-select>
            <v-select :model-value="currentWorkspace.preset" @update:model-value="onPresetChange"
                      :items="config.presets.map(v=>v.name)" color="primary"
                      label="Preset"
                      density="compact"
                      class="mx-2"></v-select>
            <v-switch v-model="currentWorkspace.no_search" label="No Search" density="compact"
                      color="primary" class="mx-2 mt-1"></v-switch>
          </div>
          <v-btn color="primary" class="mb-5 ml-2" variant="tonal" :disabled="isAsking"
                 @click="currentWorkspace.context=config.presets.find(v=>v.name===currentWorkspace.preset)?.content ?? ''">
            Reset
          </v-btn>
        </div>
        <div class="flex-grow-1">
          <textarea :style="textareaStyle" id="chat-context" class="input-textarea"
                    v-model="currentWorkspace.context"></textarea>
        </div>
        <div class="d-flex" v-if="!config.no_suggestion">
          <div style="font-size: 12px;height: 20px" class="overflow-x-hidden text-no-wrap">
            <v-chip style="cursor: pointer" v-for="item in suggestedResponses" density="compact" color="primary"
                    variant="outlined" @click="startAsking(item)"
                    class="ml-3">{{ item }}
            </v-chip>
          </div>
        </div>
        <div class="my-1 d-flex">
          <p class="font-weight-bold">Follow-up User Input:</p>
          <v-spacer></v-spacer>
          <user-input-tool-button tooltip="Upload an image" icon="mdi-file-image"
                                  :disabled="isAsking"></user-input-tool-button>
          <user-input-tool-button tooltip="Upload a document (.pdf/.docx/.pptx)" icon="mdi-file-document"
                                  :disabled="isAsking"></user-input-tool-button>
          <user-input-tool-button tooltip="Browse a webpage" icon="mdi-web"
                                  :disabled="isAsking"></user-input-tool-button>
          <user-input-tool-button tooltip="Revoke the latest user message" icon="mdi-backspace"
                                  :disabled="isAsking"></user-input-tool-button>
          <v-menu>
            <template #activator="{props}">
              <v-btn color="primary" density="compact" variant="tonal" append-icon="mdi-menu-down"
                     v-bind="props" class="mx-1"
                     :disabled="isAsking">
                Quick
              </v-btn>
            </template>
            <v-list density="compact">
              <v-list-item>test</v-list-item>
            </v-list>
          </v-menu>
          <v-btn color="primary" density="compact" variant="tonal" class="mx-1" v-if="isAsking" @click="stopAsking"
                 append-icon="mdi-stop">Stop
          </v-btn>
          <v-btn color="primary" density="compact" variant="tonal" class="mx-1" v-else @click="startAsking"
                 append-icon="mdi-send">
            Send
          </v-btn>
        </div>
        <div :style="{height:config.stretch_factor+'vh'}">
          <textarea :style="textareaStyle" id="user-input" class="input-textarea"
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