# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

![Static Badge](https://img.shields.io/badge/project-SydneyQt-blue) ![GitHub release (with filter)](https://img.shields.io/github/v/release/juzeon/SydneyQt) ![GitHub all releases](https://img.shields.io/github/downloads/juzeon/SydneyQt/total) ![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/juzeon/SydneyQt/wails.yml) ![GitHub License](https://img.shields.io/github/license/juzeon/SydneyQt)

[![SydneyQt - A desktop client for the jailbroken New Bing AI | Product Hunt](https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=438079&theme=light)](https://www.producthunt.com/posts/sydneyqt?utm_source=badge-featured&utm_medium=badge&utm_souce=badge-sydneyqt)

A cross-platform desktop client for the jailbroken New Bing AI (Sydney ver.) built with Go and [Wails](https://github.com/wailsapp/wails) ([previously](https://github.com/juzeon/SydneyQt/tree/v1) based on Python and Qt).

[简体中文](README_zh.md)

## Features

- Jailbreak New Bing with parameter tweaks and prompt injection.
- Access features in the gray-scale test in advance.
- Resolve CAPTCHA automatically via a local Selenium browser or a Bypass Server.
- Region restriction unlocking with proxy and Cloudflare Workers.
- Edit the chat context freely, including the AI's previous responses.
- Prevent Bing AI's message revoking, and automatically send custom text to continue the generation.
- Revoke and edit your last message.
- Craft, choose and send custom quick responses to the chat.
- Display the rich or plain text of the chat context, supporting LaTeX formulas, tables, codes, etc.
- Chat with webpages you browse.
- Chat with documents you open (including pdf, docx, pptx, txt and md).
- GPT-4 with vision that supports image search.
- Generate images using the latest DALL·E 3 model.
- Use OpenAI ChatGPT API with swichable different configurations.
- Switch between custom prompt presets.
- Responsible and humanized UI designs built with modern web technologies.
- Dark mode.
- Customize settings to your liking.

## Download

You can download binaries from the [release page](https://github.com/juzeon/SydneyQt/releases) for Windows, Linux and macOS, or build it yourself according to the Build section.

Platform information:

- Windows:  SydneyQt-windows-amd64.exe
- Linux:  SydneyQt-linux-amd64
- macOS: SydneyQt.app.zip, SydneyQt.pkg (unsigned)

## Usage

1. Put your `cookies.json` in the same folder as the executable file (`$HOME/Library/Application Support/SydneyQt` for macOS):
   - Install the Cookie-Editor extension for [Chrome](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm) or [Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/) (Recommend Chrome rather than Firefox since we use Chrome's network stack to bypass Bing's firewall and CAPTCHA)
   - Go to `bing.com`
   - Open the extension
   - Grant permission for All sites
   - Click `Export` on the bottom right, then `Export as JSON` (This saves your cookies to clipboard)
   - Paste your cookies into a file `cookies.json`, created in the same directory as the executable file.
   - **Note: make sure you can use the web chat before exporting the cookie.**
2. Run the program.

**(NEW) Give a try to [our new browser extension](https://github.com/juzeon/SydneyQt-browser-extension) with the abilities of automatically resolving CAPTCHA, exporting cookies, etc!**

Please follow the instructions in the next section to solve common issues.

## Common issues

### Proxy

Setting up a proxy is a must for users from mainland China.

1. Go to settings in and try different proxy types. For example: http://127.0.0.1:7890, socks5://127.0.0.1:7890 (assuming 7890 is the port to your proxy here).
2. If you use Clash or a similar proxy software, make sure that domains with the suffix `bing.com` are routed through the proxy. Some proxy providers may add `bing.com` to the direct rule, which means it will bypass the proxy.
3. If this doesn't work either, leave the proxy blank, and try using [Proxifier](https://www.proxifier.com/) or Clash TUN mode.

### Region pollution

*For Chinese users only.*

If the first time you open the Bing website without a proxy, it will redirect you to `cn.bing.com` and pollute your cookies, which means you will no longer access Bing AI with those cookies, even if you use a proxy afterwards. In case of region pollution, configure the proxy rules to make sure Bing will be accessed via proxy first and then clear all cookies from your browser or just open a privacy browsing window and log in your Microsoft account again and export the cookies finally.

### Wss Reverse Proxy

Bing bans specific countries from accessing the Bing AI (to be specific, sydney.bing.com), so in that case you need to set up a wss reverse proxy with Cloudflare Workers. Here are the steps to do that:

<details>
<summary>Click me</summary>

1. Go to [this link](https://dash.cloudflare.com/) and sign in or sign up for a Cloudflare account.
2. In the sidebar, select `Workers & Pages`.
3. On the page that opens, click `Create application`.
4. Choose `Create Worker`.
5. Give your worker a name and click `Deploy`.
6. On the worker detail page, click `Quick edit`.
7. Copy all the code from [here](https://raw.githubusercontent.com/Harry-zklcdc/go-proxy-bingai/master/cloudflare/worker.js) and paste it over the existing code in `worker.js`. Then click `Save and deploy`.
8. Copy the worker domain that looks like `xxxx-xxxx-xxxx.xxxx.workers.dev` (not a URL like `https://xxxx-xxxx-xxxx.xxxx.workers.dev/`, please remove the prefixes and suffixes) and paste it as `Wss Domain` in the settings page. Then click `Save`.
</details>

### Cookie expiration

The cookies you set up before may expire from time to time. You can check the status of your cookies in the chat page of the software. In case of expiration, just redo the cookies importing steps in the Usage section.

### CAPTCHA

Starting from v2.4.0, SydneyQt will launch a local Selenium browser to try resolving the CAPTCHA automatically, and use a [Bypass Server](https://github.com/Harry-zklcdc/go-proxy-bingai#%E4%BA%BA%E6%9C%BA%E9%AA%8C%E8%AF%81%E6%9C%8D%E5%8A%A1%E5%99%A8) instead if configured.

If this does not work, please follow these steps:

1. Check if the cookies have expired. If so, re-importing them.
2. After making sure the cookies are valid, open Bing Web in your browser and sending a random message. You should see a CAPTCHA challenge. If not, verify that the current user matches the cookies.json file. Complete the CAPTCHA and go back to the software. It should work fine now.

If you experience **infinite CAPTCHA loops**, you can try the following steps:

1. Install Bing for mobile on your phone.

2. Log in with your Microsoft account.

3. Send a message to New Bing.

**Make sure your proxy IP does not change.** If you use Clash, disable load-balancing or round-robin modes and stick to one node only. Otherwise you will need to manually solve the CAPTCHA in your browser frequently.

## Build

Environment: Go 1.21+, Node.js 16+

You can follow the development guidelines from [Wails](https://wails.io/docs/gettingstarted/installation/).

Here's the TL;DR version:

1. Install Go and Node.js.
2. Install Wails: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`.
3. Clone the project: `git clone https://github.com/juzeon/SydneyQt`.
4. Run the building command: `wails build`.

### Developer Notes

Use `debug_options_sets.json` to overwrite optionsSets, e.g:

```json
[		
	"fluxsydney",
	"iyxapbing",
	"iycapbing",
	"clgalileoall",
	"gencontentv3",
	"nojbf"
]
```

## Web API

Thanks to [@PeronGH](https://github.com/PeronGH) we now have a Web API. [Check out for more details.](webapi/README.md)

## Screenshots

![](https://public.ptree.top/ShareX/2023/12/04/1701694976/1qwHCtSW7D.png)

![](https://public.ptree.top/ShareX/2023/12/05/1701779864/syd-color.jpg)

![](https://public.ptree.top/ShareX/2023/12/11/1702287078/qUxbdxgRcN.png)

![](https://public.ptree.top/ShareX/2023/12/04/1701694905/sGRMfoZDFY.png)

![](https://public.ptree.top/ShareX/2023/12/04/1701694936/KwoV5xRVCj.png)

![](https://public.ptree.top/ShareX/2023/12/04/1701694957/vRsuaw8lOD.png)

![](https://public.ptree.top/ShareX/2023/12/04/1701696071/u8vwoftQT5.png)

![](https://public.ptree.top/ShareX/2023/12/04/1701695093/457fe0ufJZ.png)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=juzeon/SydneyQt&type=Date)](https://star-history.com/#juzeon/SydneyQt&Date)

## Acknowledgement

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>

<https://github.com/Harry-zklcdc/go-proxy-bingai>