<script setup lang="ts">
import {computed, onUpdated, ref} from "vue"
import {ChatMessage, toChatMessages, unescapeHtml} from "../../helper"
import {marked} from "marked"
import katex from 'katex'
import 'katex/dist/katex.min.css'
import {v4 as uuidV4} from "uuid"
import RichImageBlock from "./rich_blocks/RichImageBlock.vue"
import {main} from "../../../wailsjs/go/models"
import DataReference = main.DataReference

let props = defineProps<{
  context: string,
  customFontStyle: any,
  lockScroll: boolean,
  dataReferences: DataReference[],
}>()
let chatMessages = computed(() => {
  return toChatMessages(props.context)
})
let iconMap = {
  'assistant': 'mdi-account-tie-woman',
  'user': 'mdi-account',
  'system': 'mdi-laptop'
}
let showSystemPrompt = ref(false)

function renderMD(content: string) {
  const renderer = new marked.Renderer()
  let i = 0
  const next_id = () => `__special_katext_id_${i++}__`
  const math_expressions: any = {}

  function replace_math_with_ids(text: string) {
    text = text.replace(/\n+/g, '\n\n')
    text = text.replace(/\|\n\n\|/g, '|\n|')
    text = text.replace(/\$\$([\s\S]+?)\$\$/g, (_match, expression) => {
      expression = expression.replace(/\\+$/gm, '\\\\')
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
  const original_link = renderer.link
  renderer.link = function (href, title, text) {
    if (href === text) {
      return href
    }
    return original_link(href, title, text)
  }
  const original_code = renderer.code
  renderer.code = (code, infostring, escaped) => {
    let content = original_code(code, infostring, escaped)
    let uuid = uuidV4()
    content = '<div id="' + uuid + '" style="position: relative">' + content + '' +
        '<div class="copy-btn" ' +
        'onclick="fallbackCopyTextToClipboard(getTextFromSelector(\'#' + uuid + ' code\'))">Copy</div>' +
        '</div>'
    return content
  }
  let rendered_md_only = marked.parse(content, {renderer: renderer}) as string
  try {
    return rendered_md_only.replace(/(__special_katext_id_\d+__)/g, (_match, capture) => {
      const {type, expression} = math_expressions[capture]
      return katex.renderToString(unescapeHtml(expression), {displayMode: type === 'block', strict: false})
    })
  } catch (e) {
    console.log(e)
  }
  return rendered_md_only
}

interface SourceAttribute {
  index: number,
  link: string,
  title: string,
}

function renderMessage(message: ChatMessage): string {
  let content = message.message
  if (message.type === 'search_result') {
    try {
      let sourceAttributes: SourceAttribute[] = JSON.parse(message.message)
      return renderMD(sourceAttributes.map(v => '\\[' + v.index + '] [' + v.title + '](' + v.link + ')').join('<br>\n'))
    } catch (e) {
      console.log(e)
      return message.message
    }
  }
  if (message.type === 'message' && message.role === 'assistant') {
    let searchResult = findNearestSearchResult(message)
    if (searchResult) {
      try {
        let sourceAttributes: SourceAttribute[] = JSON.parse(searchResult.message)
        for (let src of sourceAttributes) {
          content = content.replaceAll('[^' + src.index + '^]',
              '[[' + src.index + ']](' + src.link + ')')
          content = content.replaceAll('(^' + src.index + '^)',
              '(' + src.link + ')')
        }
      } catch (e) {
        console.log(e)
      }
    }
  }
  return renderMD(content)
}

function findNearestSearchResult(message: ChatMessage): ChatMessage | null {
  let searchResultMessage = chatMessages.value?.[chatMessages.value.findIndex(v => v === message) - 1]
  if (!searchResultMessage || searchResultMessage.role !== 'assistant' || searchResultMessage.type !== 'search_result') {
    return null
  }
  return searchResultMessage
}

function findDataReferenceFromUUID(uuid: string): DataReference | null {
  uuid = uuid.trim()
  let item = props.dataReferences.find(v => v.uuid === uuid)
  if (item) {
    return item
  }
  return null
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
        <v-btn class="ml-3" v-if="message.role==='system' && message.type==='additional_instructions'"
               size="small" variant="text" @click="showSystemPrompt=!showSystemPrompt">
          {{ showSystemPrompt ? 'Hide' : 'Show' }}
        </v-btn>
      </div>
      <div v-if="message.type==='rich_data_reference'">
        <div v-if="!findDataReferenceFromUUID(message.message)"><i>Undefined UUID</i></div>
        <rich-image-block v-else-if="findDataReferenceFromUUID(message.message)!.type==='image'"
                          :custom-font-style="customFontStyle"
                          :data="findDataReferenceFromUUID(message.message)!.data"></rich-image-block>
      </div>
      <div v-else-if="showSystemPrompt || !(message.role==='system' && message.type==='additional_instructions')"
           v-html="renderMessage(message)" class="my-1"></div>
      <div v-else class="text-caption">...(omitted)</div>
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

#my-box pre {
  position: relative;
  overflow: auto;
  margin: 5px 0;
  padding: 1.75rem 0 1.75rem 1rem;
  border-radius: 10px;
  background: black;
  color: white;
}

#my-box pre button {
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

#my-box pre button:hover {
  cursor: pointer;
  background-color: #bcbabb;
}

#my-box table {
  border-collapse: collapse;
}

#my-box th, #my-box td {
  border: 1px solid black;
  padding: 8px;
}

#my-box li {
  margin-left: 20px;
}

#my-box p, table, li, code {
  cursor: text;
}

#my-box a {
  text-decoration: none;
}

#my-box .copy-btn {
  position: absolute;
  right: 5px;
  top: 5px;
  background-color: #999;
  border-radius: 8px;
  font-size: 16px;
  padding: 3px 5px;
  color: #333;
  cursor: pointer;
}

#my-box .copy-btn:hover {
  color: #666;
}
</style>