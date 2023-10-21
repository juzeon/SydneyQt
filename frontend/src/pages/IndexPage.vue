<script setup lang="ts">
import {computed, onMounted, ref, watch} from "vue"
import {GetConfig, SetConfig} from "../../wailsjs/go/main/Settings"
import {main} from "../../wailsjs/go/models"
import Config = main.Config

let modeList = ['Creative', 'Balanced', 'Precise']
let backendList = computed(() => {
  return ['Sydney', ...config.value.open_ai_backends.map(v => v.name)]
})
let localeList = ['zh-CN', 'en-US']
let config = ref<main.Config>(new Config())
let loading = ref(true)
watch(config, value => {
  console.log('new value:')
  console.log(value)
  SetConfig(config.value)
}, {deep: true})

async function updateFromSettings() {
  loading.value = true
  config.value = await GetConfig()
  loading.value = false
}

onMounted(() => {
  updateFromSettings()
})
</script>

<template>
  <div style="height: 100%" class="d-flex flex-column" v-if="!loading">
    <div class="d-flex align-center">
      <p class="mb-5">Chat Context:</p>
      <v-spacer></v-spacer>
      <div class="d-flex">
        <v-select v-model="config.backend" :items="backendList" color="primary" label="Backend" density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="config.conversation_style" :items="modeList" color="primary" label="Mode" density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="config.locale" :items="localeList" color="primary" label="Locale" density="compact"
                  class="mx-2"></v-select>
        <v-select v-model="config.last_preset" :items="config.presets.map(v=>v.name)" color="primary" label="Preset"
                  density="compact"
                  class="mx-2"></v-select>
      </div>
      <v-btn color="primary" class="mb-5 ml-2">Reset</v-btn>
    </div>
    <div class="flex-grow-1" style="height: 50vh;overflow: auto">
      <v-textarea class="ma-0" no-resize auto-grow :rows="20"></v-textarea>
    </div>
    <div class="my-2 d-flex">
      <p>Follow-up User Input:</p>
      <v-spacer></v-spacer>
      <v-btn color="primary" density="compact" class="mx-1">Image</v-btn>
      <v-btn color="primary" density="compact" class="mx-1">Document</v-btn>
      <v-btn color="primary" density="compact" class="mx-1">Browse</v-btn>
      <v-btn color="primary" density="compact" class="mx-1">Revoke</v-btn>
      <v-menu>
        <template #activator="{props}">
          <v-btn color="primary" density="compact" append-icon="mdi-menu-down" v-bind="props" class="mx-1">Quick</v-btn>
        </template>
        <v-list density="compact">
          <v-list-item>test</v-list-item>
        </v-list>
      </v-menu>
      <v-btn color="primary" density="compact" class="mx-1">Send</v-btn>
    </div>
    <div style="overflow: auto">
      <v-textarea class="ma-0" :rows="5" no-resize></v-textarea>
    </div>
  </div>
</template>

<style scoped>

</style>