<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {FetchWebpage} from "../../../wailsjs/go/main/App"

let props = defineProps<{
  isAsking: boolean,
}>()
let emit = defineEmits<{
  (e: 'appendBlockToCurrentWorkspace', val: string): void
}>()

let webpageFetchDialog = ref(false)
let webpageFetchURL = ref('')
let webpageFetching = ref(false)
let webpageFetchError = ref('')

function fetchWebpage() {
  webpageFetching.value = true
  webpageFetchError.value = ''
  FetchWebpage(webpageFetchURL.value).then(res => {
    let text = '[user](#webpage_context)\n'
    if (res.title === '') {
      text += JSON.stringify(res.content)
    } else {
      text += JSON.stringify(res)
    }
    text += '\n\n'
    emit('appendBlockToCurrentWorkspace', text)
    webpageFetching.value = false
    webpageFetchDialog.value = false
    webpageFetchURL.value = ''
  }).catch(err => {
    webpageFetchError.value = err.toString()
  }).finally(() => {
    webpageFetching.value = false
  })
}
</script>

<template>
  <div>
    <user-input-tool-button tooltip="Fetch a webpage" icon="mdi-web" @click="webpageFetchDialog=true"
                            :disabled="isAsking" :loading="webpageFetching"></user-input-tool-button>
    <v-dialog v-model="webpageFetchDialog" max-width="500" :persistent="true">
      <v-card>
        <v-card-title>Enter a URL to fetch</v-card-title>
        <v-card-text>
          <v-text-field :error-messages="webpageFetchError" label="URL" v-model="webpageFetchURL"
                        color="primary" @keydown.enter="fetchWebpage"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" :disabled="webpageFetching"
                 @click="webpageFetchURL='';webpageFetchError='';webpageFetchDialog=false">
            Cancel
          </v-btn>
          <v-btn variant="text" color="primary" :loading="webpageFetching" @click="fetchWebpage">Fetch</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>