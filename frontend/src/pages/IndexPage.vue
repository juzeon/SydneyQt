<script setup lang="ts">

</script>

<template>
  <div style="height: 100%" class="d-flex flex-column">
    <div class="d-flex align-center">
      <p class="mb-5">Chat Context:</p>
      <v-spacer></v-spacer>
      <div class="d-flex">
        <v-select color="primary" label="Backend" density="compact" class="mx-2"></v-select>
        <v-select color="primary" label="Mode" density="compact" class="mx-2"></v-select>
        <v-select color="primary" label="Locale" density="compact" class="mx-2"></v-select>
        <v-select color="primary" label="Preset" density="compact" class="mx-2"></v-select>
      </div>
      <v-btn color="primary" class="mb-5 ml-2">Reset</v-btn>
    </div>
    <div class="flex-grow-1" style="height: 50vh;overflow: auto">
      <v-textarea no-resize auto-grow :rows="20"></v-textarea>
    </div>
    <div class="mb-3">Tool bar</div>
    <div style="overflow: auto">
      <v-textarea :rows="5" no-resize></v-textarea>
    </div>
  </div>
</template>

<style scoped>

</style>