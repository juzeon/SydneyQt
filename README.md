# SydneyQt

A desktop client for New Bing AI (Sydney ver.) based on Python and Qt.

[简体中文](README_zh.md)

## Screenshots

![](docs/1.png)

![](docs/2.png)

![](docs/3.png)

![](docs/4.png)

## Features

- Jailbroken version of New Bing using prompt injection.
- Ability to edit chat context freely.
- Anti message revoke.
- Rich and plain text support in snapped context.
- Multiple custom prompt presets.
- Highly customizable settings.

## Usage

1. Put your `cookies.json` in the same folder as `main.py` according to the instructions in the README file of [EdgeGPT](https://github.com/acheong08/EdgeGPT):
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

## Credits

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>