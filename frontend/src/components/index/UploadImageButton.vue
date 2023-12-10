<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ref} from "vue"
import {UploadSydneyImage} from "../../../wailsjs/go/main/App"
import {swal} from "../../helper"
import {main} from "../../../wailsjs/go/models"
import UploadSydneyImageResult = main.UploadSydneyImageResult

let uploadingImage = ref(false)

let globalProps = defineProps<{
  isAsking: boolean,
  modelValue: UploadSydneyImageResult | undefined
}>()
let emit = defineEmits<{
  (e: 'update:modelValue', val: UploadSydneyImageResult | undefined): void
}>()

function uploadImage() {
  uploadingImage.value = true
  UploadSydneyImage().then(res => {
    if (res.canceled) {
      return
    }
    emit('update:modelValue', res)
  }).catch(err => {
    swal.error(err)
  }).finally(() => {
    uploadingImage.value = false
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
              <img v-if="modelValue" style="max-width: 200px;max-height: 400px"
                   :src="modelValue.base64_url" alt="img"/>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn variant="text" color="primary" @click="uploadImage">
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
      <user-input-tool-button @click="uploadImage" :bindings="modelValue?props:undefined"
                              tooltip="Upload an image"
                              icon="mdi-file-image" :color="modelValue?'green':undefined"
                              :disabled="isAsking" :loading="uploadingImage"></user-input-tool-button>
    </v-hover>
  </div>
</template>

<style scoped>

</style>