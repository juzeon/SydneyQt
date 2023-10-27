<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref, watch} from "vue"
import {GetConfig, SetConfig} from "../../wailsjs/go/main/Settings"
import {main} from "../../wailsjs/go/models"
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"
import {swal} from "../helper"
import {AskAI, CountToken} from "../../wailsjs/go/main/App"
import {AskTypeOpenAI, AskTypeSydney} from "../constants"
import Config = main.Config
import AskOptions = main.AskOptions

let modeList = ['Creative', 'Balanced', 'Precise']
let backendList = computed(() => {
  return ['Sydney', ...config.value.open_ai_backends.map(v => v.name)]
})
let localeList = ['zh-CN', 'en-US']
let config = ref<main.Config>(new Config())
let loading = ref(true)
let currentWorkspace = ref({
  id: 1,
  context: '',
  input: '',
  backend: 'Sydney',
  locale: 'zh-CN',
  preset: 'Sydney',
  conversation_style: 'Creative',
})
let tokenCount = ref(0)
let fetchingTokenCount = ref(0)
watch(config, value => {
  console.log('config changed, set config to new value:')
  console.log(value)
  SetConfig(config.value)
}, {deep: true})
watch(currentWorkspace, async (value, oldValue) => {
  console.log('currentWorkspace changed')
  if (value.context !== oldValue.context) {
    tokenCount.value = await CountToken(value.context)
  }
}, {deep: true})

async function updateFromSettings() {
  loading.value = true
  config.value = await GetConfig()
  let workspace = config.value.workspaces?.find(v => v.id === config.value.current_workspace_id)
  if (workspace) {
    currentWorkspace.value = workspace
  } else {
    currentWorkspace.value.context = config.value.presets.find(v => v.name === 'Sydney')?.content ?? ''
    config.value.workspaces = [currentWorkspace.value]
    config.value.current_workspace_id = 1
  }
  tokenCount.value = await CountToken(currentWorkspace.value.context)
  loading.value = false
  setTimeout(() => {
    scrollChatContextToBottom()
  }, 0)
}

let suggestedResponses = ref<string[]>([])
let isAsking = ref(false)
let replied = ref(false)

let askEventMap = {
  "chat_alert": (data: string) => {
    swal.error(data)
  },
  "chat_append": (data: string) => {
    let scrollBottom = false
    if (!replied.value) {
      console.log('first reply')
      fixContextLineBreak()
      currentWorkspace.value.context += '[user](#message)\n' + currentWorkspace.value.input + "\n\n"
      scrollBottom = true
      currentWorkspace.value.input = ''
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
  "chat_finish": () => {
    fixContextLineBreak()
    doListeningEvents(true)
    isAsking.value = false
    replied.value = false
  },
  "chat_suggested_responses": (data: string[]) => {
    suggestedResponses.value = data
  },
  "chat_token": (data: number) => {
    fetchingTokenCount.value = data
  },
  "chat_message_revoke": (replyDeep: number) => {
    if (config.value.revoke_reply_text != '' && replyDeep < config.value.revoke_reply_count) {
      // TODO send continue instruction
    }
  },
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

async function startAsking() {
  isAsking.value = true
  doListeningEvents()
  let askOptions = new AskOptions()
  askOptions.chat_context = currentWorkspace.value.context
  askOptions.type = currentWorkspace.value.backend === 'Sydney' ? AskTypeSydney : AskTypeOpenAI
  askOptions.prompt = currentWorkspace.value.input
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
    console.log('user-input not focused')
    return
  }
  if (isAsking.value) {
    return
  }
  console.log('handle focused key press for user-input')
  if (config.value.enter_mode === 'Enter' && (event.key == 'Enter' || event.key == 'NumpadEnter')) {
    event.preventDefault()
    startAsking()
  } else if ((event.key == 'Enter' || event.key == 'NumpadEnter') && (event.ctrlKey && event.metaKey)) {
    startAsking()
  }
}

onMounted(() => {
  updateFromSettings()
  window.addEventListener('keypress', handleKeyPress, true)
})
onUnmounted(() => {
  window.removeEventListener('keypress', handleKeyPress, true)
})


</script>

<template>
  <div style="height: 100%" class="d-flex flex-column" v-if="!loading">
    <div class="d-flex align-center">
      <p class="mb-5">Chat Context:</p>
      <v-spacer></v-spacer>
      <div class="d-flex">
        <v-select v-model="currentWorkspace.backend" :items="backendList" color="primary" label="Backend"
                  density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="currentWorkspace.conversation_style" :items="modeList" color="primary" label="Mode"
                  density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="currentWorkspace.locale" :items="localeList" color="primary" label="Locale" density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="currentWorkspace.preset" :items="config.presets.map(v=>v.name)" color="primary"
                  label="Preset"
                  density="compact"
                  class="mx-2"></v-select>
      </div>
      <v-btn color="primary" class="mb-5 ml-2" :disabled="isAsking">Reset</v-btn>
    </div>
    <div class="flex-grow-1">
      <textarea id="chat-context" class="input-textarea" v-model="currentWorkspace.context"></textarea>
    </div>
    <div class="my-2 d-flex">
      <p>Follow-up User Input:</p>
      <v-spacer></v-spacer>
      <v-btn color="primary" density="compact" class="mx-1" :disabled="isAsking">Image</v-btn>
      <v-btn color="primary" density="compact" class="mx-1" :disabled="isAsking">Document</v-btn>
      <v-btn color="primary" density="compact" class="mx-1" :disabled="isAsking">Browse</v-btn>
      <v-btn color="primary" density="compact" class="mx-1" :disabled="isAsking">Revoke</v-btn>
      <v-menu>
        <template #activator="{props}">
          <v-btn color="primary" density="compact" append-icon="mdi-menu-down" v-bind="props" class="mx-1"
                 :disabled="isAsking">Quick
          </v-btn>
        </template>
        <v-list density="compact">
          <v-list-item>test</v-list-item>
        </v-list>
      </v-menu>
      <v-btn color="primary" density="compact" class="mx-1" v-if="isAsking" @click="stopAsking">Stop</v-btn>
      <v-btn color="primary" density="compact" class="mx-1" v-else @click="startAsking">Send</v-btn>
    </div>
    <div style="height: 20vh">
      <textarea id="user-input" class="input-textarea" v-model="currentWorkspace.input"></textarea>
    </div>
  </div>
</template>

<style scoped>
.input-textarea {
  height: 99%;
  width: 100%;
  border: grey 1px dashed;
  resize: none;
  padding: 5px
}
</style>