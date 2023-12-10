<script setup lang="ts">
import {ref} from "vue"

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

function moveQuickResponse(item: string, isUp: boolean) {
  let index = props.quick.findIndex(v => v === item)
  if (index === -1 || props.quick.length <= 1) {
    return
  }
  if (isUp) {
    if (index === 0) return
    let tmp = props.quick[index]
    props.quick[index] = props.quick[index - 1]
    props.quick[index - 1] = tmp
  } else {
    if (index === props.quick.length - 1) return
    let tmp = props.quick[index]
    props.quick[index] = props.quick[index + 1]
    props.quick[index + 1] = tmp
  }
}

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
        <v-list-item v-for="(item,ix) in quick" :title="item">
          <template #prepend>
            <p style="color: #999" class="mr-2 text-no-wrap overflow-x-hidden">{{ ix + 1 }}.</p>
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
            <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                   @click="moveQuickResponse(item,true)">
              <v-icon>mdi-menu-up</v-icon>
            </v-btn>
            <v-btn variant="text" icon density="compact" class="mx-1" color="primary"
                   @click="moveQuickResponse(item,false)">
              <v-icon>mdi-menu-down</v-icon>
            </v-btn>
          </template>
        </v-list-item>
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