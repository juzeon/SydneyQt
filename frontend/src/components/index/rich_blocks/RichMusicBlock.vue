<script setup lang="ts">
import {sydney} from "../../../../wailsjs/go/models"
import dayjs from "dayjs"
import duration from "dayjs/plugin/duration"
import {ref} from "vue"
import {SaveRemoteFile} from "../../../../wailsjs/go/main/App"
import {swal} from "../../../helper"
import GenerateMusicResult = sydney.GenerateMusicResult

dayjs.extend(duration)

let props = defineProps<{
  data: GenerateMusicResult,
  customFontStyle: any,
}>()
let previewDialog = ref(false)
let saving = ref(false)

function saveFile(type: string) {
  let errHandle = (err: any) => {
    swal.error(err)
  }
  let finalize = () => {
    saving.value = false
  }
  saving.value = true
  if (type === 'audio') {
    SaveRemoteFile('mp3', props.data.title, props.data.music_url).catch(errHandle).finally(finalize)
  } else {
    SaveRemoteFile('mp4', props.data.title, props.data.video_url).catch(errHandle).finally(finalize)
  }
}
</script>

<template>
  <div :style="{'font-family':customFontStyle['font-family']}">
    <div class="d-flex mt-2">
      <div>
        <img :src="data.cover_img_url" alt="Cover Image" class="d-block">
        <div class="d-flex justify-center">
          <v-btn @click="previewDialog=true" variant="tonal" color="primary" class="mt-1">Preview</v-btn>
        </div>
      </div>
      <div class="ml-3">
        <div class="d-flex align-center">
          <div style="font-size: 18px" class="d-flex align-center">
            <v-icon class="mr-2">mdi-music</v-icon>
            <p><b>{{ data.title }}</b></p>
          </div>
          <p class="text-caption ml-3" style="color: #999">Musical style: {{ data.musical_style }}
            (generation took {{
              dayjs.duration(Math.floor(data.duration / 1000 / 1000)).asSeconds()
            }} seconds)</p>
        </div>
        <div v-html="data.lyrics.replaceAll('\n', '<br/>')"></div>
      </div>
    </div>
    <v-dialog max-width="400" v-model="previewDialog">
      <v-card title="Music Preview">
        <v-card-text class="d-flex justify-center">
          <video style="max-height: 550px" controls :src="data.video_url"></video>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" :loading="saving" @click="saveFile('video')">Save Video File</v-btn>
          <v-btn variant="text" color="primary" :loading="saving" @click="saveFile('audio')">Save Audio File</v-btn>
          <v-btn variant="text" color="primary" @click="previewDialog=false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>