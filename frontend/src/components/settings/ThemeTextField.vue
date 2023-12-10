<script setup lang="ts">
import {shadeColor} from "../../helper"
import {useTheme} from "vuetify"

let props = defineProps<{
  themeColor: string,
}>()
let emit = defineEmits<{
  (e: 'update:themeColor', val: string): void,
}>()
let theme = useTheme()

function onChangeThemeColor(val: string) {
  if (!checkThemeColor(val)) {
    return
  }
  emit('update:themeColor', val)
  theme.themes.value.light.colors.primary = val
  theme.themes.value.dark.colors.primary = shadeColor(val, -40)
}

function checkThemeColor(val: string): boolean {
  return /^#[0-9A-Fa-f]{6}$/i.test(val)
}
</script>

<template>
  <v-text-field label="Theme Color" color="primary" :model-value="themeColor"
                @update:model-value="onChangeThemeColor"
                hint="Must be a valid hex color; default value: #FF9800"></v-text-field>
</template>

<style scoped>

</style>