<script setup lang="ts">
import {computed, ref} from "vue"
import {main} from "../../../wailsjs/go/models"
import Workspace = main.Workspace

let props = defineProps<{
  isAsking: boolean,
  workspaces: Workspace[],
}>()
let emit = defineEmits<{
  (e: 'switchWorkspace', workspace: Workspace): void
}>()
let searchText = ref('')
let dialog = ref(false)
let foundWorkspaces = computed((): Workspace[] => {
  if (searchText.value === '') {
    return []
  }
  return props.workspaces.filter(v => (getWorkspaceTextForSearch(v).toLowerCase())
      .includes(searchText.value.toLowerCase()))
})

function getWorkspaceTextForSearch(workspace: Workspace) {
  let result = workspace.title + ' ' + workspace.context + ' ' + workspace.input
  result = result.replaceAll('\n', ' ')
  return result
}

function getSlugHTML(haystack: string, needle: string) {
  const window = 100
  const halfWindow = window / 2
  let ix = haystack.toLowerCase().indexOf(needle.toLowerCase())
  let result = ''
  let remainWindow = halfWindow
  if (ix < halfWindow) {
    result += haystack.substring(0, ix)
    remainWindow = window - ix
  } else {
    result += haystack.substring(ix - halfWindow, ix)
  }
  let realNeedle = haystack.substring(ix, ix + needle.length)
  result += realNeedle
  if (ix + needle.length + remainWindow >= haystack.length) {
    result += haystack.substring(ix + needle.length, haystack.length)
  } else {
    result += haystack.substring(ix + needle.length, ix + needle.length + remainWindow)
  }
  result = '...' + result.replaceAll(realNeedle, '<span style="color: red">' + realNeedle + '</span>') + '...'
  return result
}

function goToWorkspace(workspace: Workspace) {
  emit('switchWorkspace', workspace)
  dialog.value = false
}
</script>

<template>
  <div>
    <v-dialog max-width="700" v-model="dialog">
      <v-card>
        <v-card-title>Search Text in Workspaces</v-card-title>
        <v-card-text>
          <v-text-field label="Keyword" color="primary" v-model="searchText"></v-text-field>
          <v-list>
            <v-list-item v-for="workspace in foundWorkspaces" @click="goToWorkspace(workspace)">
              <template #title>{{ workspace.title }}</template>
              <template #subtitle>
                <div v-html="getSlugHTML(getWorkspaceTextForSearch(workspace),searchText)"></div>
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" variant="text" @click="dialog=false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-btn @click="searchText='';dialog=true" :disabled="isAsking" variant="text" class="flex-grow-1" color="primary"
           prepend-icon="mdi-magnify">
      Search
    </v-btn>
  </div>
</template>

<style scoped>

</style>