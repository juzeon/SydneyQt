<script setup lang="ts">
import {computed, onUpdated} from "vue"
import {toChatMessages} from "../../helper"
import {marked} from "marked"
import katex from 'katex'

let props = defineProps<{
  context: string,
  customFontStyle: any,
  lockScroll: boolean,
}>()
let chatMessages = computed(() => {
  return toChatMessages(props.context)
})
let iconMap = {
  'assistant': 'mdi-account-tie-woman',
  'user': 'mdi-account',
  'system': 'mdi-laptop'
}

function renderMD(content: string) {
  const renderer = new marked.Renderer()
  let i = 0
  const next_id = () => `__special_katext_id_${i++}__`
  const math_expressions: any = {}

  function replace_math_with_ids(text: string) {
    text = text.replace(/\n+/g, '\n\n')
    text = text.replace(/\|\n\n\|/g, '|\n|')
    text = text.replace(/\$\$([\s\S]+?)\$\$/g, (_match, expression) => {
      const id = next_id()
      math_expressions[id] = {type: 'block', expression}
      return id
    })
    text = text.replace(/\$([^\n\s]+?)\$/g, (_match, expression) => {
      const id = next_id()
      math_expressions[id] = {type: 'inline', expression}
      return id
    })
    return text
  }

  const original_listitem = renderer.listitem
  renderer.listitem = function (text, task, checked) {
    return original_listitem(replace_math_with_ids(text), task, checked)
  }
  const original_paragraph = renderer.paragraph
  renderer.paragraph = function (text) {
    return original_paragraph(replace_math_with_ids(text))
  }
  const original_tablecell = renderer.tablecell
  renderer.tablecell = function (content, flags) {
    return original_tablecell(replace_math_with_ids(content), flags)
  }
  const original_text = renderer.text
  renderer.text = function (text) {
    return original_text(replace_math_with_ids(text))
  }
  const rendered_md_only = marked.parse(content, {renderer: renderer}) as string
  try {
    return rendered_md_only.replace(/(__special_katext_id_\d+__)/g, (_match, capture) => {
      const {type, expression} = math_expressions[capture]
      return katex.renderToString(expression, {displayMode: type === 'block'})
    })
  } catch (e) {
    console.log(e)
  }
  return rendered_md_only
}

onUpdated(() => {
  if (props.lockScroll) {
    return
  }
  let myBox = document.getElementById('my-box')
  if (myBox) {
    myBox.scrollTop = myBox.scrollHeight
  }
})
</script>

<template>
  <div id="my-box" class="fill-height overflow-auto"
       :style="{border: 'grey 1px solid',padding: '5px',...customFontStyle}">
    <div v-for="(message,index) in chatMessages">
      <div class="d-flex align-center text-h6">
        <v-icon>{{ iconMap?.[message.role as keyof typeof iconMap] ?? 'mdi-account-multiple' }}</v-icon>
        <p class="ml-3" style="text-transform: uppercase!important;">{{ message.role }}</p>
        <p class="ml-3 text-caption" style="color: #999">{{ message.type }}</p>
      </div>
      <div v-html="renderMD(message.message)" class="my-1"></div>
      <v-divider class="my-3" v-if="index!==chatMessages.length-1"></v-divider>
    </div>
  </div>
</template>

<style>
*,
*:before,
*:after {
  box-sizing: border-box;
}

pre {
  position: relative;
  overflow: auto;
  margin: 5px 0;
  padding: 1.75rem 0 1.75rem 1rem;
  border-radius: 10px;
  background: black;
  color: white;
}

pre button {
  position: absolute;
  top: 5px;
  right: 5px;

  font-size: 0.9rem;
  padding: 0.15rem;
  background-color: #b1b1b1;
  border: ridge 1px #7b7b7c;
  border-radius: 5px;
  text-shadow: #e8e8e8 0 0 2px;
}

pre button:hover {
  cursor: pointer;
  background-color: #bcbabb;
}

table {
  border-collapse: collapse;
}

th, td {
  border: 1px solid black;
  padding: 8px;
}

</style>