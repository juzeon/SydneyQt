<script setup lang="ts">
import {sydney} from "../../../wailsjs/go/models"
import duration from "dayjs/plugin/duration"
import dayjs from "dayjs"
import {SaveRemoteJPEGImage} from "../../../wailsjs/go/main/App"
import GenerateImageResult = sydney.GenerateImageResult
import {swal} from "../../helper"

dayjs.extend(duration)
let props = defineProps<{
  imagePacks: GenerateImageResult[],
  customFontStyle: any,
}>()
let emit = defineEmits<{
  (e: 'update:imagePacks', val: GenerateImageResult[]): void
}>()

function saveImage(url: string) {
  SaveRemoteJPEGImage(url).catch(err => {
    swal.error(err)
  })
}
</script>

<template>
  <div class="fill-height overflow-auto"
       :style="{border: 'grey 1px solid',padding: '5px',...customFontStyle}">
    <div v-if="!imagePacks.length">
      <p class="font-italic">No generative images available.</p>
    </div>
    <div v-for="(pack,index) in imagePacks">
      <div class="d-flex align-center">
        <v-icon>mdi-image-multiple</v-icon>
        <p class="text-h6 ml-3">{{ pack.text }}</p>
        <p class="ml-3 text-caption" style="color: #999">Generation took
          {{ dayjs.duration(Math.floor(pack.duration / 1000 / 1000)).asSeconds() }} seconds</p>
      </div>
      <div class="d-flex">
        <v-img v-for="url in pack.image_urls" height="200" width="200" :src="url" class="mx-3"
               style="position: relative">
          <template v-slot:placeholder>
            <div class="d-flex align-center justify-center fill-height">
              <v-progress-circular
                  color="grey-lighten-4"
                  indeterminate
              ></v-progress-circular>
            </div>
          </template>
          <v-btn icon density="compact" color="primary" variant="tonal"
                 style="position: absolute;right: 0;bottom: 0;" @click="saveImage(url)">
            <v-icon>mdi-download</v-icon>
          </v-btn>
        </v-img>
      </div>
      <v-divider class="my-3" v-if="index!==imagePacks.length-1"></v-divider>
    </div>
  </div>
</template>

<style scoped>

</style>