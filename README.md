# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

A desktop client for New Bing AI (Sydney ver.) based on Python and Qt.

[简体中文](README_zh.md)

## Features

- Jailbroken version of New Bing using prompt injection.
- Ability to edit chat context freely.
- Anti message revoke.
- Revoke and edit the last user message sent.
- Select and send custom quick response messages to the chat.
- Rich and plain text support in snapped context, with LaTeX support.
- Browse webpages and chat with them.
- Open documents (including pdf, docx and pptx) and chat with them.
- Multiple custom prompt presets.
- Highly customizable settings.

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

## FAQ

If you encounter any of these issues: `Request is throttled`, `Authentication Failed`, `OSError: [WinError 64]`, etc, please try the following steps to fix them:

1. Update SydneyQt to the latest version.
2. Open a private browsing window in Edge, log in to bing.com and send a random message to New Bing.
3. Export the cookies.json file and replace the original one.

If these steps don't work, check your proxy settings as follows:

1. Go to settings in SydneyQt and try different proxy types. For example: http://127.0.0.1:7890, socks5h://127.0.0.1:7890 (the `h` letter means to send the hostname to the proxy)
2. If this doesn't work either, leave the proxy blank in SydneyQt, and try using [Proxifier](https://www.proxifier.com/) or Clash TUN mode.

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