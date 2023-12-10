<script setup lang="ts">
import dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime"

dayjs.extend(relativeTime)

defineProps<{
  disabled?: boolean
  active?: boolean,
  title: string,
  createdAt: string,
}>()
let emit = defineEmits<{
  (e: 'delete'): void,
  (e: 'edit'): void,
  (e: 'click'): void
}>()
</script>

<template>
  <v-card style="margin: 1px">
    <div @click="disabled?()=>{}:emit('click')" style="cursor: pointer">
      <v-card-title :class="{'font-weight-bold':active}">{{ title }}</v-card-title>
      <v-card-subtitle>{{ dayjs(createdAt).format() }}</v-card-subtitle>
    </div>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('edit')" color="primary" density="compact" icon="mdi-pencil"></v-btn>
      <v-btn @click="emit('delete')" color="primary" :disabled="disabled ?? false" density="compact"
             icon="mdi-delete"></v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>

</style>