<script setup lang="ts">
defineProps<{
  tooltip?: string,
  icon: string,
  disabled?: boolean,
  bindings?: any,
}>()
let emit = defineEmits<{
  (e: 'click'): void
}>()
</script>

<template>
  <v-tooltip :text="tooltip ?? ''" location="top">
    <template #activator="{props}">
      <v-btn v-bind="{...props,...(bindings ?? {})}" variant="tonal" color="primary" density="compact" class="mx-1"
             :disabled="disabled ?? false" @click="emit('click')">
        <v-icon>{{ icon }}</v-icon>
      </v-btn>
    </template>
  </v-tooltip>
</template>

<style scoped>

</style>