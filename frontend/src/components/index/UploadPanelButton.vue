<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {SelectUploadFile, UploadSydneyImage} from "../../../wailsjs/go/main/App"
import {swal} from "../../helper"

let uploading = ref(false)

let globalProps = defineProps<{
  isAsking: boolean,
  modelValue: any | undefined,
  type: string,// file, image
}>()
let emit = defineEmits<{
  (e: 'update:modelValue', val: any | undefined): void
}>()
let typeIconMap: any = {
  'image': 'mdi-file-image',
  'file': 'mdi-attachment'
}

function upload() {
  switch (globalProps.type) {
    case 'image':
      uploadImage()
      break
    case 'file':
      selectFile()
      break
  }
}

function selectFile() {
  SelectUploadFile().then(res => {
    emit('update:modelValue', res)
  }).catch(err => {
    swal.error(err)
  })
}

function uploadImage() {
  uploading.value = true
  UploadSydneyImage().then(res => {
    if (res.canceled) {
      return
    }
    emit('update:modelValue', res)
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    uploading.value = false
  })
}
</script>

<template>
  <div style="position: relative">
    <v-hover v-slot="{ isHovering, props }">
      <v-hover v-slot="{isHovering:subHovering,props:subProps}">
        <v-fade-transition>
          <v-card v-show="(isHovering || subHovering) && modelValue" v-bind="subProps"
                  style="position: absolute;bottom: 24px;right: 32px;">
            <v-card-text>
              <div v-if="type==='image'">
                <img v-if="modelValue" style="max-width: 200px;max-height: 400px"
                     :src="modelValue.base64_url" alt="img"/>
              </div>
              <div v-else-if="type==='file'">
                <div>{{ modelValue }}</div>
              </div>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn variant="text" color="primary" @click="upload">
                <v-icon>mdi-file-replace</v-icon>
                Replace
              </v-btn>
              <v-btn variant="text" color="red" @click="emit('update:modelValue',undefined)">
                <v-icon>mdi-close</v-icon>
                Remove
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-fade-transition>
      </v-hover>
      <user-input-tool-button @click="upload" :bindings="modelValue?props:undefined"
                              :tooltip="'Select your '+type+' to upload'"
                              :icon="typeIconMap[type]" :color="modelValue?'green':undefined"
                              :disabled="isAsking" :loading="uploading"></user-input-tool-button>
    </v-hover>
  </div>
</template>

<style scoped>

</style>