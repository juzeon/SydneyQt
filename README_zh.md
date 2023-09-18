# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

一个基于Python和Qt的越狱版新必应AI（Sydney版）的桌面客户端。

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

## 环境

- Python 3.11+ 和 pip。
- Windows 10+，macOS 或 Linux。

## 用法

1. 将你的`cookies.json`放在与`main.py`相同的文件夹中：
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

## Settings

下面是SydneyQt支持的设置项的详细介绍：

<details>
<summary>Click me</summary>

- Wss Domain：用于代理websocket接口，破解地域限制。
- Proxy：访问New Bing使用的代理，建议为http代理，例如Clash的7890端口。如果使用了Cloudflare反代的Wss域名，可能不需要梯子就能连接，但由于创建会话的HTTP GET接口依旧被墙，所以还是需要代理。
- Dark Mode：导入了Python Qt的一个自定义css实现暗黑模式效果，部分UI上可能会出现小小的渲染问题，例如文字溢出按钮等。
- Conversation Style：New Bing提供三种聊天模式，即Creative、Balanced、Precise。其中Creative和Precise模式后台是GPT-4，Balanced模式后台是GPT-3.5。建议使用Creative模式。
- No Suggestion：New Bing会根据AI的输出结果，生成三个建议的用户回复。勾选之后不显示建议栏，但实际上AI仍然会生成建议，也就是在每轮消息发送结束后要等待一段时间，这个就算通过修改optionsSets也没法关闭。
- No Search Result：目前禁用搜索的方式有在越狱prompt中指示、在每个用户发送的消息后面自动加上「#no_search」关键词这两种。这个选项使用的是第二种。
- Font Family and Size：上下文框和输入框字体字号设置。
- Stretch Factor：用来调节Chat Context和User Input输入框的占位比例，是一个整数。这个值越大，代表Chat Context越高，相应的，User Input高度就越小。
- Suggestion on Message Revoke：由于微软的限制，AI可能在输出一段内容后突然意识到不对，然后把消息撤回并道歉。当然在第三方客户端里撤回是无效的，顶多就是后续内容无法输出了。但与此同时也不会生成回复建议了。因此这个地方的文本是在这种时候用来替代建议栏显示的文本的。默认是`Continue from where you stopped`，指示AI继续输出。由于新发送的消息是将聊天记录上下文附带在webpage_context中的，不会经过外置审查，因此AI可以就刚刚中断的内容续写，除非在续写的内容中又一次出现了敏感输出。
- Revoke Auto Reply Count：如果值不为0，则当检测到消息被撤回时自动发送「消息撤回建议」的文本，以让AI继续写。最大发送次数不会超过这个地方设置的数值。
- Send Quick Responses Straightforward：输入框顶上有个Quick的按钮，用于快速发送一些模板文本。例如「翻译上面的文字为中文」之类的。这个选项在激活状态时，如果点击了Quick中的某一个模板文本，输入框里又没有文字时，就直接把模板文本发送给AI；而如果输入框中有文字时，就把模板文本加在已有文本的下面。

下面是一些ChatGPT相关的设置，因为SydneyQt是支持OpenAI的API的：

- OpenAI Key：API密钥，通常以`sk-`开头，但程序不会进行检测。
- OpenAI Endpoint：自定义OpenAI API的端点，在使用第三方分销商时有用，例如国内的`openai-sb.com`提供的API就比官方便宜不少。需要以`/v1`结尾。
- Short Model & Long Model & Model Switching Threshold：现在GPT-3.5支持4k和16k两种模型了，两种模型收费不一样。如何尽可能地减少开销？那当然是长文本用长模型，短文本用短模型了。Model Switching Threshold是一个token计数，如果当前Chat Context的token计数大于这个值，那下一次发送请求时就用Long Model，反之则用Short Model。
- Model Temperature：模型的temperature，在0到2之间，数值越大模型的输出越随机。通常保持默认即可。
</details>

## 常见问题

如果你遇到以下问题：`Request is throttled`, `Authentication Failed`, `OSError: [WinError 64]` 等，请尝试以下步骤来解决：

1. 更新 SydneyQt 到最新版本。
2. 在 Edge 浏览器中打开一个隐私窗口，登录 bing.com 并向 New Bing 发送一条随机消息。
3. 导出 cookies.json 文件并替换原来的文件。

如果这些步骤仍然无效，请检查你的代理设置，如下：

1. 在 SydneyQt 的设置中尝试不同的代理类型。例如：http://127.0.0.1:7890, socks5h://127.0.0.1:7890 (这里的 `h` 字母表示将主机名发送给代理)
2. 如果你使用Clash或类似的代理软件，请确保以`bing.com`为后缀的域名都通过代理转发。有些代理提供商可能会把`bing.com`加入直连规则，这意味着它会绕过代理。
3. 如果这也不行，就在 SydneyQt 中留空代理设置，并尝试使用 [Proxifier](https://www.proxifier.com/) 或 Clash TUN 模式。

为了避免出现`User needs to solve CAPTCHA to continue`的错误，请按照以下步骤操作：
1. 在菜单栏上选择`Cookie Checker`选项，检查当前的用户。如果显示没有用户，你需要从你的浏览器中导出一个新的cookies.json文件。
2. 确认cookie是有效的后，打开浏览器中的Bing Web，发送一条随机的消息。你应该看到一个CAPTCHA验证。如果没有，验证当前的用户是否和cookies.json文件匹配。完成CAPTCHA验证后，返回SydneyQt。它应该可以正常工作了。

您可以尝试以下步骤，以解决**无限验证码循环的问题**：

1. 在您的手机上安装移动版必应。
2. 使用您的微软账号登录。
3. 向新版必应发送一条消息。

**确保你的代理IP不要变化。**如果你使用Clash，禁用负载均衡或轮询模式，只使用一个节点。否则你会需要经常在浏览器中手动解决CAPTCHA验证。


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