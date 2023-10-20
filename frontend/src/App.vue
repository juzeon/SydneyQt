<template>
  <div>
    <v-app>
      <v-main>
        <v-app-bar color="primary">
          <v-btn icon class="mr-1" @click="navDrawer=!navDrawer">
            <v-icon>mdi-menu</v-icon>
          </v-btn>
          <p class="text-h5">SydneyQt</p>
        </v-app-bar>
        <v-navigation-drawer v-model="navDrawer">
          <v-list>
            <v-list-item title="Navigation drawer"></v-list-item>
          </v-list>
        </v-navigation-drawer>
        <v-container class="d-flex">
          <v-btn color="primary">hello</v-btn>
        </v-container>
      </v-main>
    </v-app>
  </div>
</template>

<script lang="ts" setup>
import {ref} from "vue"
let navDrawer = ref(true)
</script>