<script setup lang="ts">
import {computed} from "vue"
import {toChatMessages} from "../helper"

let props = defineProps<{
  context: string,
}>()
let chatMessages = computed(() => {
  return toChatMessages(props.context)
})
</script>

<template>
  <div style="height: 99%">

  </div>
</template>

<style scoped>

</style>