<script setup lang="ts">
import {useRouter} from "vue-router"

let router = useRouter()
</script>

<template>
  <div>
    <v-app>
      <v-main>
        <v-app-bar color="primary">
          <div class="mr-1">
            <slot name="left-top"></slot>
          </div>
          <p class="text-h5">SydneyQt</p>
          <v-spacer></v-spacer>
          <slot name="right-top-prepend"></slot>
          <slot name="right-top">
            <v-btn icon @click="router.push('/settings')">
              <v-icon>mdi-cog</v-icon>
            </v-btn>
          </slot>
        </v-app-bar>
        <div style="height: calc(100vh - 64px)" class="pa-2">
          <slot></slot>
        </div>
      </v-main>
    </v-app>
  </div>
</template>

<style scoped>

</style>