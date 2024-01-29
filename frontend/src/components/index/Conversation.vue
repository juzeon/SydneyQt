<script setup lang="ts">
import dayjs from "dayjs"
import LocalizedFormat from "dayjs/plugin/localizedFormat"
import ActionIconButton from "./ActionIconButton.vue"
import {computed} from "vue"
import {useTheme} from "vuetify"

dayjs.extend(LocalizedFormat)

let props = defineProps<{
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
  (e: 'share'): void
}>()

let theme = useTheme()
let cardBg = computed(() => {
  let obj: any = {}
  if (theme.current.value.dark) {
    obj['bg-grey-darken-3'] = props.active
  } else {
    obj['bg-grey-lighten-3'] = props.active
  }
  return obj
})
</script>

<template>
  <v-card class="ma-3" :class="cardBg">
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
      <action-icon-button @click="emit('share')" icon="mdi-share-variant" text="Share"></action-icon-button>
      <action-icon-button @click="emit('export')" icon="mdi-export" text="Export"></action-icon-button>
      <action-icon-button @click="emit('edit')" icon="mdi-pencil" text="Edit"></action-icon-button>
      <action-icon-button @click="emit('delete')" icon="mdi-delete" text="Delete"
                          :disabled="disabled ?? false"></action-icon-button>
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