import {ref, Ref, watch} from "vue"
import {main} from "../wailsjs/go/models"
import {GetConfig, SetConfig} from "../wailsjs/go/main/Settings"
import Config = main.Config

interface UseSettingsResult {
    config: Ref<Config>
    fetch: () => Promise<void>
}

export function useSettings(): UseSettingsResult {
    console.log('call useSettings')
    let config = ref(new Config())
    watch(config, value => {
        console.log('watcher: config has changed, pushing:')
        console.log(value)
        SetConfig(config.value).then(() => {
            console.log('watcher: config pushed')
        })
    }, {deep: true})
    let fetch = async () => {
        console.log('fetching settings')
        config.value = await GetConfig()
    }
    return {
        config, fetch
    }
}