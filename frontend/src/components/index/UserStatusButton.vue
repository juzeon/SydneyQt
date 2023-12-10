<script setup lang="ts">
import {onMounted, ref} from "vue"
import {GetUser} from "../../../wailsjs/go/main/App"

let currentUser = ref('')
let currentError = ref('')
let loading = ref(false)

function refresh() {
  loading.value = true
  currentUser.value = ''
  currentError.value = ''
  GetUser().then(username => {
    currentUser.value = username
    console.log('GetUser success: ' + username)
  }).catch(err => {
    currentError.value = err.toString()
    console.log('GetUser error: ' + err)
  }).finally(() => {
    loading.value = false
  })
}

onMounted(() => {
  refresh()
})
</script>

<template>
  <div>
    <v-tooltip :text="loading?'Loading...':(currentUser?'User: '+currentUser : 'Error: '+currentError)"
               location="bottom">
      <template #activator="{props}">
        <v-btn icon v-bind="props" :loading="loading" @click="refresh">
          <v-icon v-if="currentUser" color="green">mdi-account</v-icon>
          <v-icon v-else color="red">mdi-alert</v-icon>
        </v-btn>
      </template>
    </v-tooltip>
  </div>
</template>

<style scoped>

</style>