<script lang="ts" setup>
import {BrowserOpenURL} from "../wailsjs/runtime"

document.body.addEventListener('click', function (e) {
  // @ts-ignore
  if (e.target && e.target.nodeName == 'A' && e.target.href) {
    // @ts-ignore
    const url = e.target.href
    if (
        !url.startsWith('http://#') &&
        !url.startsWith('file://') &&
        !url.startsWith('http://wails.localhost:')
    ) {
      e.preventDefault()
      BrowserOpenURL(url)
    }
  }
})
</script>

<template>
  <div>
    <router-view></router-view>
  </div>
</template>