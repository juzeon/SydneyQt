<script setup lang="ts">
import {ref} from "vue"
import {v4 as uuidV4} from "uuid"
import {main} from "../../../wailsjs/go/models"
import Preset = main.Preset

let props = defineProps<{
  presets: Preset[]
}>()
let emit = defineEmits<{
  (e: 'update:presets', arr: Preset[]): void
}>()
let activePreset = ref<Preset>(props.presets[0])

function addPreset() {
  let preset = <Preset>{
    name: 'New Preset ' + uuidV4(),
    content: '[system](#additional_instructions)\n',
  }
  props.presets.push(preset)
  activePreset.value = preset
}

function deletePreset(preset: Preset) {
  if (preset.name === 'Sydney') {
    return
  }
  emit('update:presets', props.presets.filter(v => v.name !== preset.name))
  if (preset === activePreset.value) {
    activePreset.value = props.presets[0]
  }
}

let renamePresetName = ref('')
let renamePresetInstance = ref<Preset>()
let renamePresetDialog = ref(false)
let renamePresetError = ref('')

function renamePreset(preset: Preset) {
  if (preset.name === 'Sydney') {
    return
  }
  renamePresetInstance.value = preset
  renamePresetError.value = ''
  renamePresetName.value = preset.name
  renamePresetDialog.value = true
}

function confirmRenamePreset() {
  if (props.presets.find(v => v.name === renamePresetName.value)) {
    renamePresetError.value = 'Preset name already exists'
    return
  }
  renamePresetInstance.value!.name = renamePresetName.value
  renamePresetDialog.value = false
}
</script>

<template>
  <v-card class="my-3">
    <v-card-title>Presets</v-card-title>
    <v-card-text>
      <div class="d-flex" style="max-height: 400px">
        <div class="d-flex flex-column">
          <v-list density="compact" width="200" class="flex-grow-1 overflow-y-auto">
            <v-list-item :active="preset===activePreset" v-for="preset in presets">
              <template #title>
                <p style="cursor: pointer" class="overflow-x-hidden" @click="activePreset=preset">
                  {{ preset.name }}</p>
              </template>
              <template #append>
                <v-btn icon color="red" variant="text" density="compact" @click="deletePreset(preset)">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
                <v-btn icon color="primary" variant="text" density="compact" @click="renamePreset(preset)">
                  <v-icon>mdi-pencil</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>
          <v-btn variant="text" color="primary" @click="addPreset">
            <v-icon>mdi-plus</v-icon>
            Add
          </v-btn>
        </div>
        <v-textarea rows="15" class="ml-3" v-model="activePreset.content"></v-textarea>
      </div>
    </v-card-text>
    <v-dialog max-width="500" v-model="renamePresetDialog">
      <v-card>
        <v-card-title>Rename Preset</v-card-title>
        <v-card-text>
          <v-text-field label="Rename" :error-messages="renamePresetError" v-model="renamePresetName"
                        color="primary"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" @click="renamePresetDialog=false">Cancel</v-btn>
          <v-btn variant="text" color="primary" @click="confirmRenamePreset">Confirm</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<style scoped>

</style>