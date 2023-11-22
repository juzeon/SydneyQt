<script setup lang="ts"></script>

<template>
  <div>
    <v-app>
      <v-main>
        <v-app-bar color="primary">
          <slot name="left-top"></slot>
          <p class="text-h5">SydneyQt</p>
          <v-spacer></v-spacer>
          <v-btn icon>
            <v-icon>mdi-cog</v-icon>
          </v-btn>
        </v-app-bar>
        <div style="height: 98%" class="ma-2">
          <slot></slot>
        </div>
      </v-main>
    </v-app>
  </div>
</template>

<style scoped>

</style>