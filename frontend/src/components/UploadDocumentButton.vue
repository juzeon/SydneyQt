<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {UploadDocument} from "../../wailsjs/go/main/App"
import {swal} from "../helper"
import {main} from "../../wailsjs/go/models"
import Workspace = main.Workspace

let props = defineProps<{
  currentWorkspace: Workspace,
  isAsking: boolean,
}>()
let emit = defineEmits<{
  (e: 'fixContextLineBreak'): void
  (e: 'scrollChatContextToBottom'): void
}>()
let uploadingDocument = ref(false)

function uploadDocument() {
  uploadingDocument.value = true
  UploadDocument().then(res => {
    if (res.canceled) {
      return
    }
    emit('fixContextLineBreak')
    props.currentWorkspace.context += '[user](#document_context_' + res.ext?.substring(1) + '_file)\n' + res.text
    emit('scrollChatContextToBottom')
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    uploadingDocument.value = false
  })
}
</script>

<template>
  <user-input-tool-button @click="uploadDocument" tooltip="Upload a document (.pdf/.docx/.pptx)"
                          icon="mdi-file-document"
                          :loading="uploadingDocument"
                          :disabled="isAsking"></user-input-tool-button>
</template>

<style scoped>

</style>