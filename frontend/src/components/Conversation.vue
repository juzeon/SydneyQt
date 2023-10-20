<script setup lang="ts">
defineProps<{
  id: number,
  title: string,
  createdAt: Date,
}>()
</script>

<template>
  <v-card style="margin: 1px">
    <v-card-title>{{ title }}</v-card-title>
    <v-card-subtitle>{{ createdAt}}</v-card-subtitle>
    <v-card-actions>
      <v-btn density="compact" icon="mdi-delete"></v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>

</style>