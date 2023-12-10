<script setup lang="ts">
import {ref} from "vue"
import {v4 as uuidV4} from "uuid"
import {main} from "../../../wailsjs/go/models"
import OpenAIBackend = main.OpenAIBackend

let props = defineProps<{
  open_ai_backends: OpenAIBackend[],
}>()
let emit = defineEmits<{
  (e: 'update:open_ai_backends', arr: OpenAIBackend[]): void
}>()
let activeOpenaiBackendName = ref(props.open_ai_backends[0].name)
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
  if (props.open_ai_backends.find(v => v.name === renameBackendName.value) || renameBackendName.value === 'Sydney') {
    renameBackendError.value = 'Backend name already exists.'
    return
  }
  props.open_ai_backends.find(v => v.name === activeOpenaiBackendName.value)!.name = renameBackendName.value
  activeOpenaiBackendName.value = renameBackendName.value
  isRenamingBackend.value = false
}

function addOpenaiBackend() {
  let backend = <OpenAIBackend>Object.assign({},
      props.open_ai_backends.find(v => v.name === activeOpenaiBackendName.value)!)
  backend.name = 'OpenAI ' + uuidV4().split('-')[0]
  props.open_ai_backends.push(backend)
  activeOpenaiBackendName.value = backend.name
}

function deleteOpenaiBackend(backend: OpenAIBackend) {
  if (props.open_ai_backends.length <= 1) {
    return
  }
  activeOpenaiBackendName.value = props.open_ai_backends[0].name
  emit('update:open_ai_backends',props.open_ai_backends.filter(v => v !== backend))
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

function onChangeOpenAIMaxTokens(val: string) {
  let i = parseInt(val)
  if (isNaN(i)) {
    return
  }
  if (i < 0) {
    return
  }
  let backend = props.open_ai_backends.find(v => v.name === activeOpenaiBackendName.value)
  if (!backend) {
    return
  }
  backend.max_tokens = i
}
</script>

<template>
  <v-card class="my-3">
    <v-card-title>OpenAI Backends</v-card-title>
    <v-card-subtitle>Allow you to invoke OpenAI ChatGPT API.
      You can set multiple configurations and choose a backend from the chat page.
    </v-card-subtitle>
    <v-card-text>
      <div class="d-flex">
        <v-tabs class="flex-grow-1" show-arrows v-model="activeOpenaiBackendName" color="primary"
                :disabled="isRenamingBackend">
          <v-tab v-for="backend in open_ai_backends" :value="backend.name">{{ backend.name }}</v-tab>
        </v-tabs>
        <v-btn variant="text" icon color="primary" @click="addOpenaiBackend" :disabled="isRenamingBackend">
          <v-icon>mdi-plus</v-icon>
        </v-btn>
      </div>
      <v-window v-model="activeOpenaiBackendName">
        <v-window-item v-for="backend in open_ai_backends" :value="backend.name">
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
            <v-slider label="Frequency Penalty" v-model="backend.frequency_penalty"
                      color="primary" min="-2" max="2" step="0.1" thumb-label="always"></v-slider>
            <v-slider label="Presence Penalty" v-model="backend.presence_penalty"
                      color="primary" min="-2" max="2" step="0.1" thumb-label="always"></v-slider>
            <v-text-field label="Max Tokens" color="primary" :model-value="backend.max_tokens"
                          @update:model-value="onChangeOpenAIMaxTokens"
                          hint="No limitation on purpose: 0"></v-text-field>
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
</template>

<style scoped>

</style>