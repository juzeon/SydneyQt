# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

A desktop client for the jailbroken New Bing AI (Sydney ver.) based on Python and Qt.

[简体中文](README_zh.md)

## Features

- Jailbreak New Bing with prompt injection.
- Edit chat context as you wish.
- Prevent message revoking.
- Revoke and edit your last message.
- Choose and send custom quick responses to the chat.
- Use rich and plain text in snapped context, with LaTeX support.
- Chat with webpages you browse.
- Chat with documents you open (including pdf, docx and pptx).
- Send images and ask the AI to do something with them.
- Use OpenAI ChatGPT API.
- Switch between custom prompt presets.
- Dark mode.
- Customize settings to your liking.

## Environment

- Python 3.11+ with pip.
- Windows 10+, macOS or Linux.

## Usage

1. Put your `cookies.json` in the same folder as `main.py`:
   - Install the Cookie-Editor extension for [Chrome](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm) or [Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)
   - Go to `bing.com`
   - Open the extension
   - Click `Export` on the bottom right, then `Export as JSON` (This saves your cookies to clipboard)
   - Paste your cookies into a file `cookies.json`, created in the same directory as `main.py`.
2. Install requirements:

```bash
pip install -r requirements.txt
```

3. Run the program:

```bash
python main.py
```

4. If you see an error message like `200, message='Invalid response status', url=URL('wss://sydney.bing.com/sydney/ChatHub')`, you need to set up a proxy service with Cloudflare Workers. Here are the steps to do that:

<details>
<summary>Click me</summary>

1. Go to [this link](https://dash.cloudflare.com/) and sign in or sign up for a Cloudflare account.
2. In the sidebar, select `Workers & Pages`.
3. On the page that opens, click `Create application`.
4. Choose `Create Worker`.
5. Give your worker a name and click `Deploy`.
6. On the worker detail page, click `Quick edit`.
7. Copy all the code from [here](https://github.com/adams549659584/go-proxy-bingai/blob/master/cloudflare/worker.js) and paste it over the existing code in `worker.js`. Then click `Save and deploy`.
8. Copy the worker domain that looks like `xxxx-xxxx-xxxx.xxxx.workers.dev` (not a URL like `https://xxxx-xxxx-xxxx.xxxx.workers.dev/`, please remove the prefixes and suffixes) and paste it as `Wss Domain` in the settings dialog of SydneyQt. Then click `Save`.
</details>

## Settings

Below is the detailed description on the settings of SydneyQt.

<details>
<summary>Click me</summary>

- Wss Domain: Used to proxy websocket interface, break regional restrictions.
- Proxy: The proxy used to access New Bing, recommended to be an http proxy, such as Clash's 7890 port. If you use a Wss domain that is reverse-proxied by Cloudflare, you may not need a proxy to connect, but since the HTTP GET endpoint for creating conversations is still blocked, you still need a proxy.
- Dark Mode: Imported a custom css from Python Qt to implement dark mode effect, some minor rendering problems may occur on some UI, such as text overflowing buttons, etc.
- Conversation Style: New Bing provides three chat modes, namely Creative, Balanced, and Precise. Among them, Creative and Precise modes are backed by GPT-4, and Balanced mode is backed by GPT-3.5. It is recommended to use Creative mode.
- No Suggestion: New Bing will generate three suggested user responses based on AI's output results. After checking this, the suggestion bar will not be displayed, but AI will still generate suggestions, which means that you have to wait for a while after each round of message sending ends. This can't be turned off even by modifying optionsSets.
- No Search Result: There are currently two ways to disable search: instructing in the jailbreak prompt and automatically adding the "#no_search" keyword after each user-sent message. This option uses the latter.
- Font Family and Size: Font family and size settings for context box and input box.
- Stretch Factor: Used to adjust the placeholder ratio of Chat Context and User Input input box, which is an integer. The larger this value, the higher the Chat Context, and correspondingly, the smaller the User Input height.
- Suggestion on Message Revoke: Due to Microsoft's restrictions, AI may suddenly realize that something is wrong after outputting a piece of content, and then withdraw the message and apologize. Of course, revoking in a third-party client is invalid, at most it means that subsequent content cannot be output. But at the same time it will not generate reply suggestions either. Therefore, the text here is used to replace the suggestion bar display text at this time. The default is `Continue from where you stopped`, instructing AI to continue outputting. Since the new message sent will attach the chat record context in webpage_context, it will not go through external censorship, so AI can continue writing on the content that was just interrupted, unless there is sensitive output again in the continued content.
- Revoke Auto Reply Count: If the value is not 0, then when a message withdrawal is detected, it will automatically send the "message revoke suggestion" text to let AI continue writing. The maximum number of times sent will not exceed the value set here.
- Send Quick Responses Straightforward: There is a Quick button on top of the input box for quickly sending some template texts. Such as "Translate the above text into Chinese" and so on. When this option is activated, if you click on one of the template texts in Quick and there is no text in the input box, it will directly send the template text to AI; and if there is text in the input box, it will add the template text below the existing text.

Below are some ChatGPT related settings, because SydneyQt supports OpenAI's API:

- OpenAI Key: API key, usually starting with `sk-`, but the program will not validate it.
- OpenAI Endpoint: Custom OpenAI API endpoint, useful when using third-party distributors, such as `openai-sb.com` in China provides API that is much cheaper than official ones. It needs to end with `/v1`.
- Short Model & Long Model & Model Switching Threshold: Now GPT-3.5 supports 4k and 16k two models, and the two models charge differently. How to minimize costs as much as possible? Of course it is long text with long model and short text with short model. Model Switching Threshold is a token count. If the current Chat Context's token count is greater than this value, then use Long Model for the next request sent; otherwise use Short Model.
- Model Temperature: The model's temperature, between 0 and 2. The larger the value, the more random the model's output. Usually keep default.
</details>

## FAQ

If you encounter any of these issues: `Request is throttled`, `Authentication Failed`, `OSError: [WinError 64]`, etc, please try the following steps to fix them:

1. Update SydneyQt to the latest version.
2. Open a private browsing window in Edge, log in to bing.com and send a random message to New Bing.
3. Export the cookies.json file and replace the original one.

If these steps don't work, check your proxy settings as follows:

1. Go to settings in SydneyQt and try different proxy types. For example: http://127.0.0.1:7890, socks5h://127.0.0.1:7890 (the `h` letter means to send the hostname to the proxy)
2. If you use Clash or a similar proxy software, make sure that domains with the suffix `bing.com` are routed through the proxy. Some proxy providers may add `bing.com` to the direct rule, which means it will bypass the proxy.
3. If this doesn't work either, leave the proxy blank in SydneyQt, and try using [Proxifier](https://www.proxifier.com/) or Clash TUN mode.

To avoid the `User needs to solve CAPTCHA to continue` error, please follow these steps:
1. Check the current user with the `Cookie Checker` option on the menu bar. If it shows no user, you need to export a new cookies.json file from your browser.
2. After making sure the cookie is valid, open Bing Web in your browser and sending a random message. You should see a CAPTCHA challenge. If not, verify that the current user matches the cookies.json file. Complete the CAPTCHA and go back to SydneyQt. It should work fine now.

If you experience **infinite CAPTCHA loops**, you can try the following steps:

1. Install Bing for mobile on your phone.

2. Log in with your Microsoft account.

3. Send a message to New Bing.

**Make sure your proxy IP does not change.** If you use Clash, disable load-balancing or round-robin modes and stick to one node only. Otherwise you will need to manually solve the CAPTCHA in your browser frequently.

## Screenshots

![](docs/1.png)

![](docs/2.png)

![](docs/3.png)

![](docs/4.png)

![](docs/5.png)

![](docs/6.png)

![](docs/7.png)

Dark mode is supported now:

![](docs/8.png)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=juzeon/SydneyQt&type=Date)](https://star-history.com/#juzeon/SydneyQt&Date)

## Credits

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>