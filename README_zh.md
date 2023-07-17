# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

一个基于Python和Qt的新必应AI（Sydney版）的桌面客户端。

## 特点

- 用提示注入破解新必应。
- 随心所欲地编辑聊天内容。
- 阻止消息撤回。
- 撤回并编辑你的最后一条消息。
- 选择并发送自定义的快速回复到聊天中。
- 在截取的内容中使用富文本和纯文本，支持 LaTeX。
- 和你浏览的网页聊天。
- 和你打开的文档聊天（包括 pdf，docx 和 pptx）。
- 发送图片并让AI对它们做些什么。
- 使用 OpenAI ChatGPT API。
- 在自定义的提示预设之间切换。
- 夜间模式。
- 根据你的喜好定制设置。

## 用法

1. 根据[EdgeGPT](https://github.com/acheong08/EdgeGPT)的README文件中的说明，将你的`cookies.json`放在与`main.py`相同的文件夹中：
   - 为[Chrome](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)或[Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)安装Cookie-Editor扩展
   - 前往`bing.com`
   - 打开扩展
   - 点击右下角的`Export`，然后选择`Export as JSON`（这会将你的cookies保存到剪贴板）
   - 将你的cookies粘贴到一个名为`cookies.json`的文件中，请在与`main.py`同级的文件夹下创建。
2. 安装依赖：

```bash
pip install -r requirements.txt
```

3. 运行程序：

```bash
python main.py
```

4. 如果你看到一个错误信息，像这样 `200, message='Invalid response status', url=URL('wss://sydney.bing.com/sydney/ChatHub')`，你需要用 Cloudflare Workers 建立一个代理服务。以下是操作步骤：

<details>
<summary>点击我</summary>

1. 点击[这个链接](https://dash.cloudflare.com/)，登录或注册一个 Cloudflare 账号。
2. 在侧边栏，选择 `Workers & Pages`。
3. 在打开的页面，点击 `Create application`。
4. 选择 `Create Worker`。
5. 给你的 worker 起一个名字，然后点击 `Deploy`。
6. 在 worker 详情页面，点击 `Quick edit`。
7. 从[这里](https://github.com/adams549659584/go-proxy-bingai/blob/master/cloudflare/worker.js)复制所有的代码，然后粘贴到 `worker.js` 中，覆盖原有的代码。然后点击 `Save and deploy`。
8. 复制 worker 域名，看起来像 `xxxx-xxxx-xxxx.xxxx.workers.dev`（请填入域名，而不是URL：不是`https://xxxx-xxxx-xxxx.xxxx.workers.dev/`，请移除前后缀），然后粘贴到 SydneyQt 的设置对话框中的 `Wss Domain`。然后点击 `Save`。
</details>

## 常见问题

如果你遇到以下问题：`Request is throttled`, `Authentication Failed`, `OSError: [WinError 64]` 等，请尝试以下步骤来解决：

1. 更新 SydneyQt 到最新版本。
2. 在 Edge 浏览器中打开一个隐私窗口，登录 bing.com 并向 New Bing 发送一条随机消息。
3. 导出 cookies.json 文件并替换原来的文件。

如果这些步骤仍然无效，请检查你的代理设置，如下：

1. 在 SydneyQt 的设置中尝试不同的代理类型。例如：http://127.0.0.1:7890, socks5h://127.0.0.1:7890 (这里的 `h` 字母表示将主机名发送给代理)
2. 如果这也不行，就在 SydneyQt 中留空代理设置，并尝试使用 [Proxifier](https://www.proxifier.com/) 或 Clash TUN 模式。

要避免出现`User needs to solve CAPTCHA to continue`的提示，请确保你的代理IP不变。如果你使用Clash，关闭负载均衡或轮询模式，只选择一个节点。

## 截图

![](docs/1.png)

![](docs/2.png)

![](docs/3.png)

![](docs/4.png)

![](docs/5.png)

![](docs/6.png)

![](docs/7.png)

现已支持夜间模式：

![](docs/8.png)

## Star 记录

[![Star History Chart](https://api.star-history.com/svg?repos=juzeon/SydneyQt&type=Date)](https://star-history.com/#juzeon/SydneyQt&Date)

## 致谢

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>