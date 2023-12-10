<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {UploadDocument} from "../../../wailsjs/go/main/App"
import {swal} from "../../helper"

let props = defineProps<{
  isAsking: boolean,
}>()
let emit = defineEmits<{
  (e: 'appendBlockToCurrentWorkspace', val: string): void
}>()
let uploadingDocument = ref(false)

function uploadDocument() {
  uploadingDocument.value = true
  UploadDocument().then(res => {
    if (res.canceled) {
      return
    }
    emit('appendBlockToCurrentWorkspace', '[user](#document_context_' + res.ext?.substring(1) + '_file)\n' + res.text)
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