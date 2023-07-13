# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

A desktop client for New Bing AI (Sydney ver.) based on Python and Qt.

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
- Customize settings to your liking.

## Usage

1. (Optional) Put your `cookies.json` in the same folder as `main.py` according to the instructions in the README file of [EdgeGPT](https://github.com/acheong08/EdgeGPT):
   - Install the Cookie-Editor extension for [Chrome](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm) or [Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)
   - Go to `bing.com`
   - Open the extension
   - Click `Export` on the bottom right, then `Export as JSON` (This saves your cookies to clipboard)
   - Paste your cookies into a file `cookies.json`
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
8. Copy the worker domain that looks like `xxxx-xxxx-xxxx.xxxx.workers.dev` and paste it as `Wss Domain` in the settings dialog of SydneyQt. Then click `Save`.
</details>

## FAQ

If you encounter any of these issues: `Request is throttled`, `Authentication Failed`, `OSError: [WinError 64]`, etc, please try the following steps to fix them:

1. Update SydneyQt to the latest version.
2. Open a private browsing window in Edge, log in to bing.com and send a random message to New Bing.
3. Export the cookies.json file and replace the original one.

If these steps don't work, check your proxy settings as follows:

1. Go to settings in SydneyQt and try different proxy types. For example: http://127.0.0.1:7890, socks5h://127.0.0.1:7890 (the `h` letter means to send the hostname to the proxy)
2. If this doesn't work either, leave the proxy blank in SydneyQt, and try using [Proxifier](https://www.proxifier.com/) or Clash TUN mode.

To avoid the `User needs to solve CAPTCHA to continue` message, make sure your proxy IP does not change. If you use Clash, disable load-balancing or round-robin modes and stick to one node only.

## Screenshots

![](docs/1.png)

![](docs/2.png)

![](docs/3.png)

![](docs/4.png)

![](docs/5.png)

![](docs/6.png)

![](docs/7.png)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=juzeon/SydneyQt&type=Date)](https://star-history.com/#juzeon/SydneyQt&Date)

## Credits

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>