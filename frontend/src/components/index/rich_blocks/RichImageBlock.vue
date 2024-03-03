<script setup lang="ts">
import {sydney} from "../../../../wailsjs/go/models"
import dayjs from "dayjs"
import {SaveRemoteJPEGImage} from "../../../../wailsjs/go/main/App"
import {swal} from "../../../helper"
import duration from "dayjs/plugin/duration"
import GenerateImageResult = sydney.GenerateImageResult

dayjs.extend(duration)

let props = defineProps<{
  data: GenerateImageResult,
  customFontStyle: any,
}>()

function saveImage(url: string) {
  SaveRemoteJPEGImage(url).catch(err => {
    swal.error(err)
  })
}

</script>

<template>
  <div>
    <div class="d-flex align-center text-caption mx-3 my-1" style="color: #999">
      <v-icon>mdi-image-multiple</v-icon>
      <p class="ml-3" :style="{'font-family':customFontStyle['font-family']}">
        Prompt: {{ data.text }} (generation took
        {{ dayjs.duration(Math.floor(data.duration / 1000 / 1000)).asSeconds() }} seconds)</p>
    </div>
    <div class="d-flex">
      <v-img v-for="url in data.image_urls" height="200" width="200" :src="url" class="mx-3"
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
  </div>
</template>

<style scoped>

</style>