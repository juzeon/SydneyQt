<script setup lang="ts">

import {marked} from "marked"
import {BrowserOpenURL} from "../../../wailsjs/runtime"
import {onMounted, ref} from "vue"
import {CheckUpdate} from "../../../wailsjs/go/main/App"
import {main} from "../../../wailsjs/go/models"
import CheckUpdateResult = main.CheckUpdateResult

let versionResult = ref<CheckUpdateResult | undefined>(undefined)
let versionError = ref('')
let versionLoading = ref(false)
let versionDialog = ref(false)

function checkUpdate() {
  versionLoading.value = true
  versionError.value = ''
  versionResult.value = undefined
  CheckUpdate().then(res => {
    versionResult.value = res
    if (versionResult.value.need_update) {
      versionDialog.value = true
    }
  }).catch(err => {
    versionError.value = err.toString()
  }).finally(() => {
    versionLoading.value = false
  })
}

onMounted(() => {
  checkUpdate()
})
</script>

<template>
  <div>
    <div class="d-flex align-center">
      <v-icon size="large">mdi-update</v-icon>
      <div class="ml-3">
        <p v-if="versionLoading">Checking update...</p>
        <p v-if="versionError">Error checking update: {{ versionError }}</p>
        <div v-if="versionResult">
          <p v-if="versionResult.need_update">New update available: {{ versionResult.latest_version }}
            (current: {{ versionResult.current_version }})</p>
          <p v-else>You are using the latest version: {{ versionResult.current_version }}</p>
        </div>
      </div>
      <v-spacer></v-spacer>
      <v-btn variant="text" color="primary" v-if="versionResult?.need_update" @click="versionDialog=true">
        View
      </v-btn>
      <v-btn variant="text" color="primary" :loading="versionLoading" @click="checkUpdate">Re-Check</v-btn>
    </div>
    <v-dialog max-width="500" v-model="versionDialog">
      <v-card :title="'New Version: '+versionResult?.latest_version">
        <v-card-text>
          <div
              v-html="marked.parse(versionResult?.release_note?.replace(/\n+/g,'\n\n')
                      || '*No release note.*') as string"></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary"
                 @click="BrowserOpenURL(versionResult?.release_url ?? '');versionDialog=false">
            Download
          </v-btn>
          <v-btn variant="text" color="primary" @click="versionDialog=false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>