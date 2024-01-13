<script setup lang="ts">
import dayjs from "dayjs"
import LocalizedFormat from "dayjs/plugin/localizedFormat"

dayjs.extend(LocalizedFormat)

defineProps<{
  disabled?: boolean
  active?: boolean,
  title: string,
  createdAt: string,
}>()
let emit = defineEmits<{
  (e: 'delete'): void,
  (e: 'edit'): void,
  (e: 'click'): void,
  (e: 'export'): void
}>()
</script>

<template>
  <v-card style="margin: 1px" :class="{'bg-grey-lighten-3':active}">
    <v-card-text>
      <div @click="disabled?()=>{}:emit('click')" style="cursor: pointer">
        <p :class="{'font-weight-bold':active,'conversation-title':true}">{{ title }}</p>
        <p style="color: #999;font-size: 14px" :class="{'font-weight-bold':active}"
           class="text-no-wrap overflow-hidden">{{
            dayjs(createdAt).format('llll')
          }}</p>
      </div>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn @click="emit('export')" color="primary" density="comfortable" icon="mdi-export"></v-btn>
      <v-btn @click="emit('edit')" color="primary" density="compact" icon="mdi-pencil"></v-btn>
      <v-btn @click="emit('delete')" color="primary" :disabled="disabled ?? false" density="compact"
             icon="mdi-delete"></v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.conversation-title {
  font-size: 16px;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>