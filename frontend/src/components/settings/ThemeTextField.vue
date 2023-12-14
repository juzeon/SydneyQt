<script setup lang="ts">
import {shadeColor} from "../../helper"
import {useTheme} from "vuetify"
import {ref} from "vue"

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

let colorDialog = ref(false)

</script>

<template>
  <div class="d-flex align-center">
    <v-text-field label="Theme Color" color="primary" :model-value="themeColor"
                  @update:model-value="onChangeThemeColor"
                  hint="Must be a valid hex color; default value: #FF9800"></v-text-field>
    <v-btn icon variant="tonal" color="primary" class="mb-5 ml-3" @click="colorDialog=true">
      <v-icon>mdi-select-color</v-icon>
    </v-btn>
    <v-dialog v-model="colorDialog" max-width="350">
      <v-card title="Pick a Theme Color">
        <v-card-text>
          <v-color-picker :hide-inputs="true" :model-value="themeColor" @update:model-value="onChangeThemeColor"
                          :show-swatches="true"></v-color-picker>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" variant="text" @click="colorDialog=false">Done</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>

</style>