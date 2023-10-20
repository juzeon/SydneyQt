<script lang="ts" setup>
import {ref} from "vue"

let navDrawer = ref(true)
</script>

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
        <div style="height: 98%" class="ma-2">
          <router-view></router-view>
        </div>
      </v-main>
    </v-app>
  </div>
</template>