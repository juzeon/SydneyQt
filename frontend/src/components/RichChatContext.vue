<script setup lang="ts">
import {computed} from "vue"
import {toChatMessages} from "../helper"

let props = defineProps<{
  context: string,
}>()
let chatMessages = computed(() => {
  return toChatMessages(props.context)
})
let iconMap = {
  'assistant': 'mdi-account-tie-woman',
  'user': 'mdi-account',
  'system': 'mdi-laptop'
}
</script>

<template>
  <div class="fill-height overflow-y-auto">
    <div v-for="message in chatMessages">
      <v-icon>{{ iconMap?.[message.role] ?? 'mdi-account-multiple' }}</v-icon>
      <div>{{ message.message }}</div>
    </div>
  </div>
</template>

<style scoped>

</style>