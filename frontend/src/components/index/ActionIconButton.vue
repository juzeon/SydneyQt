<script setup lang="ts">
let props = defineProps<{
  icon: string,
  text: string,
  disabled?: boolean,
}>()
let emit = defineEmits<{
  (e: 'click'): void
}>()
</script>

<template>
  <v-tooltip location="bottom" :text="text">
    <template #activator="{props}">
      <v-btn @click="emit('click')" color="primary" density="compact" :icon="icon" v-bind="props"
             :disabled="disabled ?? false"></v-btn>
    </template>
  </v-tooltip>
</template>

<style scoped>

</style>