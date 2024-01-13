<script setup lang="ts">
import {ref} from "vue"
import Vuedraggable from 'vuedraggable'

let props = defineProps<{
  quick: string[],
}>()
let emit = defineEmits<{
  (e: 'update:quick', arr: string[]): void,
}>()
let quickRespEditMode = ref('create')
let quickRespEditText = ref('')
let quickRespEditDialog = ref(false)
let quickRespEditError = ref('')
let quickRespEditIndex = ref(-1)

function createQuickResponse() {
  quickRespEditMode.value = 'create'
  quickRespEditText.value = ''
  quickRespEditDialog.value = true
  quickRespEditError.value = ''
}

function editQuickResponse(index: number) {
  quickRespEditMode.value = 'edit'
  quickRespEditText.value = props.quick[index]
  quickRespEditIndex.value = index
  quickRespEditDialog.value = true
  quickRespEditError.value = ''
}

function confirmQuickResponse() {
  if (props.quick.find(v => v === quickRespEditText.value)) {
    quickRespEditError.value = 'This Quick Response already exists.'
    return
  }
  if (quickRespEditMode.value === 'create') {
    props.quick.push(quickRespEditText.value.trim())
  } else {
    props.quick[quickRespEditIndex.value] = quickRespEditText.value.trim()
  }
  quickRespEditDialog.value = false
}
</script>

<template>
  <v-card class="my-3">
    <v-card-title>Quick Responses</v-card-title>
    <v-card-text>
      <v-list density="compact">
        <vuedraggable :list="quick" handle=".handle">
          <template #item="{ element:item,index:ix }">
            <v-list-item :title="item">
              <template #prepend>
                <v-icon class="handle mr-1" style="cursor: grab" size="large">mdi-drag</v-icon>
              </template>
              <template #append>
                <v-btn variant="text" icon density="compact" class="mx-1" color="red"
                       @click="emit('update:quick',quick.filter(v=>v!==item))">
                  <v-icon>mdi-delete</v-icon>
                </v-btn>
                <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                       @click="editQuickResponse(ix)">
                  <v-icon>mdi-pencil</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </template>
        </vuedraggable>
      </v-list>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn variant="text" prepend-icon="mdi-plus" color="primary" @click="createQuickResponse">
          Add
        </v-btn>
      </v-card-actions>
    </v-card-text>
    <v-dialog max-width="500" v-model="quickRespEditDialog">
      <v-card :title="quickRespEditMode==='create'?'Create a Quick Response':'Edit the Quick Response'">
        <v-card-text>
          <v-text-field :error-messages="quickRespEditError" label="Quick Response"
                        v-model="quickRespEditText"
                        color="primary"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn variant="text" color="primary" @click="quickRespEditDialog=false">Cancel</v-btn>
          <v-btn variant="text" color="primary" @click="confirmQuickResponse">Confirm</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<style scoped>

</style>