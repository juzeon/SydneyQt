<script setup lang="ts">

import UserInputToolButton from "./UserInputToolButton.vue"
import {ChatMessage, swal, toChatMessages} from "../../helper"
import {main} from "../../../wailsjs/go/models"
import Workspace = main.Workspace

let props = defineProps<{
  isAsking: boolean,
  currentWorkspace: Workspace,
}>()

function getChatMessages(): ChatMessage[] {
  let ctx = props.currentWorkspace.context
  return toChatMessages(ctx)
}

function setChatMessages(arr: ChatMessage[]) {
  props.currentWorkspace.context = arr.map(v => `[${v.role}](#${v.type})\n${v.message}`).join('\n\n') + '\n\n'
}

function handleRevoke() {
  let arr = getChatMessages()
  let usersArr = arr.filter(v => v.role === 'user' && v.type === 'message')
  if (usersArr.length < 1) {
    swal.error('Nothing to revoke')
    return
  }
  props.currentWorkspace.input = usersArr[usersArr.length - 1].message
  // @ts-ignore
  setChatMessages(arr.slice(0, arr.findLastIndex(v => v === usersArr[usersArr.length - 1])))
}
</script>

<template>
  <user-input-tool-button tooltip="Revoke the latest user message" icon="mdi-undo" @click="handleRevoke"
                          :disabled="isAsking"></user-input-tool-button>
</template>

<style scoped>

</style>