<script setup lang="ts">

import Conversation from "./Conversation.vue"
import SearchWorkspaceButton from "./SearchWorkspaceButton.vue"
import {main, sydney} from "../../../wailsjs/go/models"
import {generateRandomName} from "../../helper"
import dayjs from "dayjs"
import {computed, ref} from "vue"
import Workspace = main.Workspace
import Preset = main.Preset
import GenerateImageResult = sydney.GenerateImageResult

let props = defineProps<{
  modelValue: boolean,
  workspaces: Workspace[],
  currentWorkspace: Workspace,
  isAsking: boolean,
  presets: Preset[],
}>()
let sortedWorkspaces = computed(() => {
  return props.workspaces.sort((a, b) => {
    return b.id - a.id
  }) ?? []
})
let emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void
  (e: 'onReset'): void,
  (e: 'update:workspaces', workspaces: Workspace[]): void
  (e: 'update:currentWorkspace', workspace: Workspace): void
  (e: 'update:suggestedResponses', arr: string[]): void
}>()

function onDeleteWorkspace(workspace: Workspace) {
  if (sortedWorkspaces.value.length <= 1) {
    workspace.title = 'Chat ' + generateRandomName()
    workspace.input = ''
    workspace.created_at = dayjs().format()
    emit('onReset')
    return
  }
  emit('update:workspaces', props.workspaces.filter(v => v.id !== workspace.id))
  if (workspace.id === props.currentWorkspace.id) {
    switchWorkspace(sortedWorkspaces.value[0])
  }
}

let editWorkspaceDialog = ref(false)
let editWorkspaceTitle = ref('')
let editWorkspaceIndex = ref(0)

function onEditWorkspace(workspace: Workspace) {
  editWorkspaceTitle.value = workspace.title
  editWorkspaceIndex.value = workspace.id
  editWorkspaceDialog.value = true
}

function confirmEditWorkspace() {
  if (editWorkspaceTitle.value === '') {
    return
  }
  let workspace = sortedWorkspaces.value.find(v => v.id === editWorkspaceIndex.value)!
  workspace.title = editWorkspaceTitle.value
  editWorkspaceDialog.value = false
}


function addWorkspace() {
  let nextID = sortedWorkspaces.value[0].id + 1
  let workspace = <Workspace>{
    id: nextID,
    title: 'Chat ' + generateRandomName(),
    created_at: dayjs().format(),
    no_search: props.currentWorkspace.no_search,
    backend: props.currentWorkspace.backend,
    context: props.presets.find(v => v.name === props.currentWorkspace.preset)?.content ?? '',
    conversation_style: props.currentWorkspace.conversation_style,
    input: '',
    locale: props.currentWorkspace.locale,
    preset: props.currentWorkspace.preset,
    image_packs: <GenerateImageResult[]>[],
  }
  props.workspaces.push(workspace)
  switchWorkspace(workspace)
}

function switchWorkspace(workspace: Workspace) {
  emit('update:currentWorkspace', workspace)
  emit('update:suggestedResponses', [])
}
</script>

<template>
  <div>
    <v-navigation-drawer :model-value="modelValue" @update:model-value="val => emit('update:modelValue',val)"
                         :permanent="true">
      <div class="d-flex flex-column fill-height">
        <v-list class="overflow-y-auto flex-grow-1">
          <v-list-item v-for="workspace in sortedWorkspaces">
            <conversation :title="workspace.title" :created-at="workspace.created_at"
                          :active="workspace.id===currentWorkspace.id" :disabled="isAsking"
                          @delete="onDeleteWorkspace(workspace)" @edit="onEditWorkspace(workspace)"
                          @click="switchWorkspace(workspace)"></conversation>
          </v-list-item>
        </v-list>
        <div class="d-flex ma-3">
          <v-btn :disabled="isAsking" @click="addWorkspace" variant="text" class="flex-grow-1" color="primary"
                 prepend-icon="mdi-plus">
            Add
          </v-btn>
          <search-workspace-button @switch-workspace="switchWorkspace" :is-asking="isAsking"
                                   :workspaces="sortedWorkspaces"></search-workspace-button>
        </div>
      </div>
    </v-navigation-drawer>
    <v-dialog max-width="500" v-model="editWorkspaceDialog">
      <v-card>
        <v-card-text>
          <v-text-field color="primary" label="Workspace Title" v-model="editWorkspaceTitle"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" @click="editWorkspaceDialog=false">Cancel</v-btn>
          <v-btn variant="text" color="primary" @click="confirmEditWorkspace">Confirm</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>