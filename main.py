import asyncio
import json
import pathlib
import re
import signal
import traceback

import openai
import tiktoken
import sydney
from PySide6.QtCore import QEvent
from PySide6.QtGui import QTextCursor, Qt, QFont, QIcon
from PySide6.QtWidgets import (
    QApplication,
    QLabel,
    QPushButton,
    QWidget, QPlainTextEdit, QErrorMessage, QHBoxLayout, QFileDialog, QToolButton, QMenu, QSizePolicy, QVBoxLayout,
    QSplitter, QComboBox, QProgressBar, QSpacerItem, QLayout, QStatusBar, QListView, QListWidget, QMessageBox, QMenuBar,
    QGridLayout
)
from qasync import QEventLoop, asyncSlot
from browse_window import BrowseWindow
from cookie_checker import CookieChecker
from document import read_pptx_text, read_pdf_text, read_docx_text
from hyperlink_widget import HyperlinkWidget
from name_dialog import NameDialog
from preset_window import PresetWindow
from quick_template_window import QuickTemplateWindow
from setting_window import SettingWindow
from snap_window import SnapWindow
from user_input import UserInput
from config import Config
from visual_search_window import VisualSearchWindow

signal.signal(signal.SIGINT, signal.SIG_DFL)


class SydneyWindow(QWidget):
    def __init__(self, config: Config, parent=None):
        super().__init__(parent)
        self.visual_search_url = ""
        self.visual_search_window = None
        self.deleted_workspace = False
        self.updating_workspace_list = False
        self.current_responding_task = None
        self.snap_window = None
        self.preset_window = None
        self.setting_window = None
        self.browse_window = None
        self.quick_template_window = None
        self.cookie_checker = None
        self.updating_presets = False
        self.config = config
        self.responding = False
        self.enter_mode = "Enter"

        openai.api_key = self.config.get('openai_key')
        openai.api_base = self.config.get('openai_endpoint')
        openai.proxy = self.config.get('proxy')

        self.status_label = QLabel('Ready.')
        self.token_count_label = QLabel('Token Count')
        self.chat_history = QPlainTextEdit()
        self.chat_history.textChanged.connect(self.update_token_count)
        self.chat_history.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.user_input = UserInput(self, config=self.config)
        self.user_input.textChanged.connect(self.update_token_count)
        self.snap_button = QPushButton("Snap")
        self.snap_button.setToolTip('Take a snapshot of the current chat context '
                                    'and open in a new window with rich text support.')
        self.snap_button.clicked.connect(self.snap_context)
        self.reset_button = QPushButton("Reset")
        self.reset_button.setToolTip('Reset the current chat context using the selected preset.')
        self.reset_button.clicked.connect(self.clear_context)
        self.load_button = QPushButton("Load")
        self.load_button.setToolTip('Load a text file into the chat context.')
        self.load_button.clicked.connect(self.load_file)
        self.save_button = QPushButton("Save")
        self.save_button.setToolTip('Save the current chat context into a text file.')
        self.save_button.clicked.connect(self.save_file)
        self.snap_button.setFixedWidth(50)
        self.reset_button.setFixedWidth(50)
        self.load_button.setFixedWidth(50)
        self.save_button.setFixedWidth(50)
        self.send_button = QToolButton()
        self.send_button.setText("Send")
        self.send_button.setPopupMode(QToolButton.ToolButtonPopupMode.MenuButtonPopup)
        menu = QMenu(self)
        self.enter_action = menu.addAction("Press Enter to send", lambda: self.set_enter_mode("Enter"))
        self.enter_action.setCheckable(True)
        self.ctrl_enter_action = menu.addAction("Press Ctrl+Enter to send",
                                                lambda: self.set_enter_mode("Ctrl+Enter"))
        self.ctrl_enter_action.setCheckable(True)
        self.set_enter_mode(self.config.cfg.get('enter_mode', 'Enter'))
        self.send_button.setMenu(menu)

        self.presets = QComboBox()
        self.presets.currentTextChanged.connect(self.presets_changed)
        self.update_presets()

        self.locales = QComboBox()
        self.locales.addItems(['zh-CN', 'en-US', 'en-IE', 'en-GB'])
        self.locales.setCurrentText(self.config.get('locale'))
        self.locales.setToolTip('Locale hint for the conversation.')

        def change_locale():
            self.config.cfg['locale'] = self.locales.currentText()
            self.config.save()

        self.locales.currentTextChanged.connect(change_locale)

        self.backend = QComboBox()
        self.backend.addItems(["Sydney", "ChatGPT"])
        self.backend.setCurrentText(self.config.get('backend'))
        self.backend.setToolTip('Switch this from sydney to openai will disable Sydney and enable ChatGPT.')

        def change_backend():
            self.config.cfg['backend'] = self.backend.currentText()
            self.config.save()

        self.backend.currentTextChanged.connect(change_backend)

        upper_half = QWidget()
        upper_half_layout = QVBoxLayout()
        upper_half.setLayout(upper_half_layout)
        upper_half_buttons = QHBoxLayout()
        upper_half_layout.addLayout(upper_half_buttons)
        upper_half_buttons.addWidget(QLabel("Chat Context:"))
        upper_half_buttons.addStretch()
        backend_label = QLabel('Backend:')
        backend_label.setStyleSheet("margin-left: 3px")
        preset_label = QLabel('Preset:')
        preset_label.setStyleSheet("margin-left: 3px")
        locale_label = QLabel('Locale:')
        locale_label.setStyleSheet("margin-left: 3px")
        upper_half_buttons.addWidget(backend_label)
        upper_half_buttons.addWidget(self.backend)
        upper_half_buttons.addWidget(locale_label)
        upper_half_buttons.addWidget(self.locales)
        upper_half_buttons.addWidget(preset_label)
        upper_half_buttons.addWidget(self.presets)
        action_label = QLabel('Actions:')
        action_label.setStyleSheet("margin-left: 3px")
        upper_half_buttons.addWidget(action_label)
        upper_half_buttons.addWidget(self.reset_button)
        upper_half_buttons.addWidget(self.snap_button)
        upper_half_buttons.addWidget(self.load_button)
        upper_half_buttons.addWidget(self.save_button)
        upper_half_layout.addWidget(self.chat_history)

        bottom_half = QWidget()
        bottom_half_layout = QVBoxLayout()
        bottom_half.setLayout(bottom_half_layout)
        bottom_half_buttons = QHBoxLayout()

        self.suggestion_layout = QHBoxLayout()
        self.suggestion_layout.setContentsMargins(0, 0, 0, 0)
        self.suggestion_widget = QWidget()
        self.suggestion_widget.setLayout(self.suggestion_layout)
        self.suggestion_widget.setVisible(not self.config.cfg.get('no_suggestion'))

        self.visual_search_button = QPushButton('Image')
        self.visual_search_button.setToolTip('Perform a visual search in the current chat context.')
        self.visual_search_button.clicked.connect(self.visual_search)

        self.document_button = QPushButton('Document')
        self.document_button.clicked.connect(self.open_document)
        self.document_button.setToolTip('Import a document file into the chat context.')

        self.browse_button = QPushButton('Browse')
        self.browse_button.clicked.connect(self.open_browse_window)
        self.browse_button.setToolTip('Browse a webpage and insert its content into the chat context.')

        self.stop_button = QPushButton('Stop')
        self.stop_button.clicked.connect(self.stop_responding_task)
        self.stop_button.setDisabled(True)

        self.revoke_button = QPushButton('Revoke')
        self.revoke_button.clicked.connect(self.revoke_edit)
        self.revoke_button.setToolTip('Revoke the latest message sent by the user.')

        self.quick_button = QToolButton()
        self.quick_button.setPopupMode(QToolButton.ToolButtonPopupMode.InstantPopup)
        self.quick_button.setText('Quick')
        self.quick_menu = QMenu(self)

        def use_quick(txt):
            def quick_func():
                if self.user_input.toPlainText() == "" and self.config.get('direct_quick'):
                    self.current_responding_task = asyncio.ensure_future(self.send_message(text_to_send=txt))
                    return
                self.user_input.moveCursor(QTextCursor.MoveOperation.End)
                if not self.user_input.toPlainText().endswith('\n') and self.user_input.toPlainText() != "":
                    self.user_input.insertPlainText('\n')
                self.user_input.insertPlainText(txt)
                self.user_input.moveCursor(QTextCursor.MoveOperation.End)

            return quick_func

        def reload_quick_menu():
            self.quick_menu.clear()
            for text in self.config.get('quick'):
                self.quick_menu.addAction(text, use_quick(str(text).replace('\\n', '\n')))
            self.quick_menu.addAction('<Edit>', open_quick_template_window)

        def open_quick_template_window():
            def on_save():
                reload_quick_menu()

            self.quick_template_window = QuickTemplateWindow(config=self.config, on_save=on_save)
            self.quick_template_window.show()

        reload_quick_menu()
        self.quick_button.setMenu(self.quick_menu)

        bottom_half_layout.addWidget(self.suggestion_widget)
        bottom_half_layout.addLayout(bottom_half_buttons)
        bottom_half_buttons.addWidget(QLabel("Follow-up User Input:"))
        bottom_half_buttons.addStretch()
        bottom_half_buttons.addWidget(self.visual_search_button)
        bottom_half_buttons.addWidget(self.document_button)
        bottom_half_buttons.addWidget(self.browse_button)
        bottom_half_buttons.addWidget(self.stop_button)
        bottom_half_buttons.addWidget(self.revoke_button)
        bottom_half_buttons.addWidget(self.quick_button)
        bottom_half_buttons.addWidget(self.send_button)
        bottom_half_layout.addWidget(self.user_input)

        push_button_size = self.revoke_button.sizeHint()
        self.send_button.setFixedSize(push_button_size)
        self.quick_button.setFixedSize(push_button_size)

        self.status_bar = QStatusBar()
        self.status_bar.addWidget(self.status_label)
        self.status_bar.addPermanentWidget(self.token_count_label)

        bottom_half_layout.addWidget(self.status_bar)

        self.splitter = QSplitter(Qt.Orientation.Vertical)
        self.splitter.addWidget(upper_half)
        self.splitter.addWidget(bottom_half)
        self.splitter.setStretchFactor(0, self.config.get('stretch_factor'))
        self.splitter.setStretchFactor(1, 1)

        left_layout = QVBoxLayout()
        left_layout.setContentsMargins(8, 8, 8, 8)
        self.workspace_list_widget = QListWidget()
        self.workspace_dict = dict(self.config.get('workspaces'))
        workspace_names = list[str](self.workspace_dict.keys())
        self.workspace_list_widget.addItems(workspace_names)
        self.workspace_ix = self.config.get('workspace_ix')
        self.current_workspace_name = self.config.get('last_workspace')
        self.workspace_list_widget.setCurrentRow(workspace_names.index(self.current_workspace_name))
        self.workspace_list_widget.currentRowChanged.connect(self.switch_workspace)
        workspace_label = QLabel('Workspace: ')
        left_layout.addWidget(workspace_label)
        left_layout.addWidget(self.workspace_list_widget)
        left_buttons_layout = QGridLayout()
        self.add_workspace_button = QPushButton('New')
        self.add_workspace_button.clicked.connect(self.add_workspace)
        self.del_workspace_button = QPushButton('Delete')
        self.del_workspace_button.clicked.connect(self.del_workspace)
        self.rename_workspace_button = QPushButton('Rename')
        self.rename_workspace_button.clicked.connect(self.rename_workspace)
        self.clear_workspace_button = QPushButton('Clear')
        self.clear_workspace_button.clicked.connect(self.clear_workspace)
        left_buttons_layout.addWidget(self.add_workspace_button, 0, 0)
        left_buttons_layout.addWidget(self.del_workspace_button, 0, 1)
        left_buttons_layout.addWidget(self.rename_workspace_button, 1, 0)
        left_buttons_layout.addWidget(self.clear_workspace_button, 1, 1)
        left_layout.addLayout(left_buttons_layout)

        self.left_layout_widget = QWidget()
        self.left_layout_widget.setContentsMargins(0, 0, 0, 0)
        self.left_layout_widget.setLayout(left_layout)
        if not self.config.get('workspace_toggle'):
            self.left_layout_widget.hide()
        menu_bar = QMenuBar()
        menu_bar.setNativeMenuBar(False)

        def toggle_workspace():
            if self.left_layout_widget.isHidden():
                self.left_layout_widget.show()
                self.config.cfg['workspace_toggle'] = True
                self.config.save()
            else:
                self.left_layout_widget.hide()
                self.config.cfg['workspace_toggle'] = False
                self.config.save()

        def open_cookie_checker():
            self.cookie_checker = CookieChecker(config=self.config)
            self.cookie_checker.show()

        menu_bar.addAction('Show/Hide Workspace', toggle_workspace)
        menu_bar.addAction('Cookie Checker', open_cookie_checker)
        menu_bar.addAction('Settings', self.open_setting_window)
        main_layout = QHBoxLayout()
        main_layout.addWidget(self.left_layout_widget, 1)
        main_layout.addWidget(self.splitter, 6)
        main_layout.setContentsMargins(0, 0, 0, 0)

        frame_layout = QVBoxLayout()
        frame_layout.setContentsMargins(0, 0, 0, 0)
        frame_layout.addWidget(menu_bar)
        frame_layout.addLayout(main_layout)

        self.setLayout(frame_layout)

        self.resize(900, 600)
        self.setWindowTitle('SydneyQt')
        icon = QIcon('binglogo.png')
        self.setWindowIcon(icon)

        self.send_button.clicked.connect(self.send_clicked)

        if self.workspace_dict[self.current_workspace_name]['context'] == "":
            self.clear_context()
        else:
            self.restore_workspace()
            self.set_suggestion_line()

    def stop_responding_task(self):
        if self.current_responding_task is not None:
            self.current_responding_task.cancel()
            self.update_status_text('Stopped current responding task.')
            self.set_responding(False)

    def send_clicked(self):
        self.current_responding_task = asyncio.ensure_future(self.send_message())

    async def send_message(self, text_to_send: str = None):
        if str(self.config.get('backend')).lower() == 'sydney':
            await self.send_sydney(text_to_send)
        else:
            await self.send_openai(text_to_send)

    async def send_openai(self, text_to_send: str = None):
        openai.api_key = self.config.get('openai_key')
        openai.api_base = self.config.get('openai_endpoint')
        if self.responding:
            return
        self.set_responding(True)
        self.update_status_text('Creating conversation...')
        user_input = self.user_input.toPlainText()
        if text_to_send is not None:
            user_input = text_to_send
        context_arr = self.get_chat_context_array(self.chat_history.toPlainText())
        context_arr.append({'role': 'user', 'type': 'message', 'message': user_input})
        current_model = str(self.config.get('openai_short_model'))
        current_encoder_name = 'gpt-3.5-turbo' if current_model.startswith('gpt-3.5') else 'gpt-4'
        try:
            messages = []
            for item in context_arr:
                content = item['message']
                if item['type'] != 'message':
                    content = '(#' + item['type'] + ') ' + content
                messages.append({'role': item['role'], 'content': content})
            all_context = ' '.join([message['content'] for message in messages])
            current_length = len(tiktoken.encoding_for_model(current_encoder_name).encode(all_context))
            # print(current_length)
            if current_length > self.config.get('openai_threshold'):
                current_model = self.config.get('openai_long_model')
                current_encoder_name = 'gpt-3.5-turbo' if current_model.startswith('gpt-3.5') else 'gpt-4'
            response = await openai.ChatCompletion.acreate(model=current_model, messages=messages, stream=True,
                                                           temperature=self.config.get('openai_temperature'))
            if text_to_send is None:
                self.user_input.clear()
        except Exception as e:
            print(e)
            err_text = 'Cannot create conversation'
            err = err_text + ': ' + str(e)
            QErrorMessage(self).showMessage(err)
            self.update_status_text(err_text + '.')
            self.set_responding(False)
            return
        self.update_status_text(f'Fetching response using {current_model}...')
        self.append_chat_context(f"[user](#message)\n{user_input}\n\n", new_block=True)
        first_chunk = True
        token_wrote = 0
        try:
            async for chunk in response:
                if 'content' not in chunk['choices'][0]['delta']:
                    continue
                if first_chunk:
                    self.append_chat_context(f"[assistant](#message)\n", new_block=True)
                    first_chunk = False
                content = chunk['choices'][0]['delta']['content']
                token_wrote += len(tiktoken.encoding_for_model(current_encoder_name).encode(content))
                self.append_chat_context(content)
                self.update_status_text(f'Fetching response using {current_model}, '
                                        f'{token_wrote} tokens received currently.')
        except Exception as e:
            print(e)
            err_text = 'Error fetching response'
            err = err_text + ': ' + str(e)
            QErrorMessage(self).showMessage(err)
            self.update_status_text(err_text + '.')
            self.set_responding(False)
            return
        self.update_status_text('Ready.')
        self.set_responding(False)

    async def send_sydney(self, text_to_send: str = None, reply_deep=0):
        if self.responding:
            return
        self.set_responding(True)
        self.update_status_text('Creating conversation...')
        user_input = self.user_input.toPlainText()
        if text_to_send is not None:
            user_input = text_to_send
        proxy = self.config.get('proxy')
        try:
            cookie_path = pathlib.Path('cookies.json')
            cookies = None
            if cookie_path.exists():
                cookies = json.loads(cookie_path.read_text(encoding='utf-8'))
            conversation = await sydney.create_conversation(cookies=cookies, proxy=proxy if proxy != "" else None)
        except Exception as e:
            traceback.print_exc()
            QErrorMessage(self).showMessage(str(e))
            self.update_status_text('Error: ' + str(e))
            self.set_responding(False)
            return
        if text_to_send is None:
            self.user_input.clear()
        self.update_status_text('Fetching response...')
        message_revoked = False
        revoke_reply_text = self.config.get('revoke_reply_text')
        revoke_reply_count = self.config.get('revoke_reply_count')

        async def stream_output():
            nonlocal message_revoked
            nonlocal revoke_reply_count
            nonlocal revoke_reply_text
            self.append_chat_context(f"[user](#message)\n{user_input}\n\n", new_block=True)
            wrote = 0
            replied = False
            async for response in sydney.ask_stream(
                    conversation=conversation,
                    prompt=user_input + (" #no_search" if self.config.cfg['no_search'] else ""),
                    context=self.chat_history.toPlainText(),
                    conversation_style=self.config.cfg['conversation_style'],
                    locale=self.config.get('locale'),
                    proxy=proxy if proxy != "" else None,
                    image_url=self.visual_search_url,
            ):
                # print(response)
                if response["type"] == 1 and "messages" in response["arguments"][0]:
                    self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
                    message = response["arguments"][0]["messages"][0]
                    msg_type = message.get("messageType")
                    if msg_type == "InternalSearchQuery":
                        self.append_chat_context(
                            f"[assistant](#search_query)\n{message['hiddenText']}\n\n")
                    elif msg_type == "InternalSearchResult":
                        try:
                            links = []
                            if 'Web search returned no relevant result' in message['hiddenText']:
                                self.append_chat_context(
                                    f"[assistant](#search_results)\n{message['hiddenText']}\n\n")
                            else:
                                for group in json.loads(message['hiddenText'][8:-4]).values():
                                    for sub_group in group:
                                        links.append(
                                            f'[^{sub_group["index"]}^][{sub_group["title"]}]({sub_group["url"]})')
                                self.append_chat_context(
                                    "[assistant](#search_results)\n" + '\n\n'.join(links) + "\n\n")
                        except Exception as err:
                            print('Error when parsing InternalSearchResult: ' + str(err))
                            traceback.print_exc()
                    elif msg_type == "InternalLoaderMessage":
                        if 'hiddenText' in message:
                            self.append_chat_context(
                                f"[assistant](#loading)\n{message['hiddenText']}\n\n")
                        elif 'text' in message:
                            self.append_chat_context(
                                f"[assistant](#loading)\n{message['text']}\n\n")
                        else:
                            self.append_chat_context(
                                f"[assistant](#loading)\n{json.dumps(message)}\n\n")
                    elif msg_type is None:
                        if "cursor" in response["arguments"][0]:
                            self.append_chat_context("[assistant](#message)\n")
                            wrote = 0
                        if message.get("contentOrigin") == "Apology":
                            message_revoked = True
                            if replied and (revoke_reply_text == '' or reply_deep >= revoke_reply_count):
                                QErrorMessage(self).showMessage("Message revoke detected")
                            else:
                                raise Exception("Looks like the user message has triggered the Bing filter")
                            break
                        else:
                            replied = True
                            self.append_chat_context(message["text"][wrote:])
                            wrote = len(message["text"])
                            token_wrote = len(tiktoken.encoding_for_model('gpt-4').encode(message["text"]))
                            self.update_status_text(f'Fetching response, {token_wrote} tokens received currently.')
                            if "suggestedResponses" in message:
                                suggested_responses = list(
                                    map(lambda x: x["text"], message["suggestedResponses"]))
                                self.set_suggestion_line(suggested_responses)
                                break
                    else:
                        print(f'Unsupported message type: {msg_type}')
                        print(f'Triggered by {user_input}, response: {message}')
                if response["type"] == 2 and "item" in response and "messages" in response["item"]:
                    message = response["item"]["messages"][-1]
                    if "suggestedResponses" in message:
                        suggested_responses = list(
                            map(lambda x: x["text"], message["suggestedResponses"]))
                        self.set_suggestion_line(suggested_responses)
                        break

        try:
            await stream_output()
        except Exception as e:
            traceback.print_exc()
            QErrorMessage(self).showMessage(str(e))
            self.update_status_text('Error: ' + str(e))
        else:
            self.update_status_text('Ready.')
        self.set_responding(False)
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
        if revoke_reply_text != '' and message_revoked:
            if reply_deep < revoke_reply_count:
                await self.send_sydney(revoke_reply_text, reply_deep + 1)
            else:
                self.set_suggestion_line([revoke_reply_text])

    def revoke_edit(self):
        arr = self.get_chat_context_array()
        users_arr = [obj for obj in arr if obj['role'] == 'user' and obj['type'] == 'message']
        if len(users_arr) < 1:
            QMessageBox(self).information(self, 'Message', 'Nothing to revoke.')
            return
        self.user_input.setPlainText(users_arr[-1]['message'])
        self.apply_chat_context_array(arr[:arr.index(users_arr[-1])])

    def get_chat_context_array(self, chat_context: str = None):
        ctx = self.chat_history.toPlainText() if chat_context is None else chat_context
        ctx += '\n\n[system](#sydney__placeholder)'
        result = re.compile(
            r"\[(system|user|assistant)]\(#(.*?)\)([\s\S]*?)(?=\n.*?(^\[(system|user|assistant)]\(#.*?\)))", re.M) \
            .findall(ctx)
        arr = []
        for match in result:
            if match[1] == 'sydney__placeholder':
                continue
            arr.append({
                'role': match[0],
                'type': match[1],
                'message': str(match[2]).strip()
            })
        return arr

    def apply_chat_context_array(self, arr):
        self.chat_history.setPlainText('')
        self.append_chat_context(
            '\n\n'.join([f"[{obj['role']}](#{obj['type']})\n{obj['message']}" for obj in arr]) + '\n\n')

    def add_workspace(self):
        self.workspace_ix = self.workspace_ix + 1
        self.workspace_dict[f'Workspace {self.workspace_ix}'] = {
            'context': self.config.get_last_preset_text(),
            'input': '',
            'backend': self.backend.currentText(),
            'locale': self.locales.currentText(),
            'preset': self.presets.currentText()
        }
        self.workspace_list_widget.addItem(f'Workspace {self.workspace_ix}')
        self.workspace_list_widget.setCurrentRow(self.workspace_list_widget.count() - 1)

    def del_workspace(self):
        if self.workspace_list_widget.count() <= 1:
            QMessageBox(self).information(self, 'Message', 'You cannot delete the only workspace.')
            return
        name = self.workspace_list_widget.currentItem().text()
        del self.workspace_dict[name]
        self.deleted_workspace = True
        self.workspace_list_widget.takeItem(self.workspace_list_widget.currentRow())

    def clear_workspace(self):
        msg_box = QMessageBox()
        msg_box.setIcon(QMessageBox.Icon.Warning)
        msg_box.setText("This will clear all your workspaces and cannot be undone!")
        msg_box.setStandardButtons(QMessageBox.StandardButton.Yes | QMessageBox.StandardButton.No)
        resp = msg_box.exec()
        if resp != QMessageBox.StandardButton.Yes:
            return
        self.config.cfg['workspaces'] = dict(self.config.defaultCfg['workspaces'])
        self.config.cfg['workspace_ix'] = self.config.defaultCfg['workspace_ix']
        self.config.cfg['last_workspace'] = self.config.defaultCfg['last_workspace']
        self.config.save()
        self.updating_workspace_list = True
        for row in range(self.workspace_list_widget.count()):
            self.workspace_list_widget.takeItem(0)
        self.current_workspace_name = self.config.cfg['last_workspace']
        self.workspace_ix = 1
        self.workspace_dict = self.config.cfg['workspaces']
        self.workspace_list_widget.addItem(self.current_workspace_name)
        self.workspace_list_widget.setCurrentRow(0)
        self.updating_workspace_list = False

    def rename_workspace(self):
        dialog = NameDialog()
        if not dialog.exec():
            return
        name = dialog.get_name()
        if name == "":
            return
        if name in self.workspace_dict:
            QMessageBox(self).information(self, 'Message', f'Workspace named {name} has already exists.')
            return
        old_name = self.workspace_list_widget.currentItem().text()
        self.workspace_list_widget.currentItem().setText(name)
        workspace = self.workspace_dict[old_name]
        del self.workspace_dict[old_name]
        self.workspace_dict[name] = workspace
        self.current_workspace_name = name

    def flush_workspace(self):
        self.workspace_dict[self.current_workspace_name] = {
            'context': self.chat_history.toPlainText(),
            'input': self.user_input.toPlainText(),
            'backend': self.backend.currentText(),
            'locale': self.locales.currentText(),
            'preset': self.presets.currentText()
        }

    def _restore_optional_workspace_value(self, workspace_name):
        if 'backend' in self.workspace_dict[workspace_name]:
            self.backend.setCurrentText(self.workspace_dict[workspace_name]['backend'])
        if 'locale' in self.workspace_dict[workspace_name]:
            self.locales.setCurrentText(self.workspace_dict[workspace_name]['locale'])
        if 'preset' in self.workspace_dict[workspace_name]:
            self.presets.setCurrentText(self.workspace_dict[workspace_name]['preset'])

    def restore_workspace(self):
        self.chat_history.setPlainText(self.workspace_dict[self.current_workspace_name]['context'])
        self.user_input.setPlainText(self.workspace_dict[self.current_workspace_name]['input'])
        self._restore_optional_workspace_value(self.current_workspace_name)
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)

    def switch_workspace(self):
        if self.updating_workspace_list:
            return
        new_workspace_name = self.workspace_list_widget.currentItem().text()
        if self.deleted_workspace:
            self.deleted_workspace = False
        else:
            self.flush_workspace()
        self.chat_history.setPlainText(self.workspace_dict[new_workspace_name]['context'])
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
        self.user_input.setPlainText(self.workspace_dict[new_workspace_name]['input'])
        self.user_input.moveCursor(QTextCursor.MoveOperation.End)
        self._restore_optional_workspace_value(new_workspace_name)
        self.current_workspace_name = new_workspace_name
        self.set_suggestion_line()

    def append_chat_context(self, text, new_block=False):
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
        if new_block:
            history = self.chat_history.toPlainText()
            if not history.endswith("\n\n"):
                if history.endswith("\n"):
                    self.chat_history.insertPlainText("\n")
                else:
                    self.chat_history.insertPlainText("\n\n")
        self.chat_history.insertPlainText(text)
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)

    def open_browse_window(self):
        self.browse_window = BrowseWindow(
            self.config,
            on_insert=lambda context: self.append_chat_context(f"[user](#webpage_context)\n```\n{context}\n```\n\n"))
        self.browse_window.show()

    def update_token_count(self):
        count_chat_ctx = len(tiktoken.encoding_for_model('gpt-4').encode(self.chat_history.toPlainText()))
        count_user_input = len(tiktoken.encoding_for_model('gpt-4').encode(self.user_input.toPlainText()))
        self.token_count_label.setText(f'Chat Context: {count_chat_ctx} tokens; User Input: {count_user_input} tokens.')

    def clear_layout(self, layout: QLayout):
        while layout.count() > 0:
            item = layout.takeAt(0)
            if item.widget():
                item.widget().deleteLater()
            elif item.layout():
                self.clear_layout(item.layout())
            del item

    def set_suggestion_line(self, suggestions=None):
        self.clear_layout(self.suggestion_layout)
        self.suggestion_layout.addWidget(QLabel('Suggestions: '))
        if suggestions:
            for suggestion in suggestions:
                def make_hyperlink_clicked(s):
                    def hyperlink_clicked():
                        self.set_suggestion_line()
                        self.current_responding_task = asyncio.ensure_future(self.send_message(s))

                    return hyperlink_clicked

                self.suggestion_layout.addWidget(
                    HyperlinkWidget(suggestion, on_clicked=make_hyperlink_clicked(suggestion)))
        else:
            self.suggestion_layout.addWidget(QLabel('<Not available>'))
        self.suggestion_layout.addItem(QSpacerItem(0, 0, QSizePolicy.Expanding, QSizePolicy.Minimum))

    def update_status_text(self, text: str):
        self.status_label.setText(text)

    def update_settings(self):
        self.chat_history.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.user_input.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.suggestion_widget.setVisible(not self.config.cfg.get('no_suggestion'))
        self.update_status_text('Settings updated successfully.')

    def open_setting_window(self):
        self.setting_window = SettingWindow(self.config, on_close=self.update_settings)
        self.setting_window.show()

    def snap_context(self):
        self.snap_window = SnapWindow(self.config, self.chat_history.toPlainText())
        self.snap_window.show()

    def clear_context(self):
        self.chat_history.setPlainText(self.config.get_last_preset_text())
        self.set_suggestion_line()
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)

    @asyncSlot()
    async def open_document(self):
        file_dialog = QFileDialog(self)
        file_dialog.setWindowTitle('Open a document to chat with it')
        file_dialog.setNameFilters(["Document files (*.pptx *.pdf *.docx)"])
        file_dialog.setAcceptMode(QFileDialog.AcceptMode.AcceptOpen)
        if file_dialog.exec():
            self.set_responding(True)
            self.update_status_text('Loading document...')
            loop_local = asyncio.get_event_loop()
            file_name = file_dialog.selectedFiles()[0]
            ext = pathlib.Path(file_name).suffix
            try:
                if ext == ".pptx":
                    text = await loop_local.run_in_executor(None, read_pptx_text, file_name)
                    self.append_chat_context(f'[user](#pptx_slide_context)\n```\n{text}\n```\n\n')
                elif ext == ".pdf":
                    text = await loop_local.run_in_executor(None, read_pdf_text, file_name)
                    self.append_chat_context(f'[user](#pdf_document_context)\n```\n{text}\n```\n\n')
                elif ext == ".docx":
                    text = await loop_local.run_in_executor(None, read_docx_text, file_name)
                    self.append_chat_context(f'[user](#docx_document_context)\n```\n{text}\n```\n\n')
                else:
                    QErrorMessage(self).showMessage('Unsupported file type')
            except Exception as e:
                QErrorMessage(self).showMessage(str(e))
            else:
                self.update_status_text(f'Loaded a {ext} document successfully.')
            self.set_responding(False)

    def load_file(self):
        file_dialog = QFileDialog(self)
        file_dialog.setNameFilters(["Text files (*.txt)", "All files (*)"])
        file_dialog.setAcceptMode(QFileDialog.AcceptMode.AcceptOpen)
        if file_dialog.exec():
            file_name = file_dialog.selectedFiles()[0]
            with open(file_name, "r", encoding='utf-8') as f:
                file_content = f.read()
            self.chat_history.setPlainText(file_content)
            self.chat_history.moveCursor(QTextCursor.MoveOperation.End)

    def save_file(self):
        file_dialog = QFileDialog(self)
        file_dialog.setNameFilters(["Text files (*.txt)", "All files (*)"])
        file_dialog.setAcceptMode(QFileDialog.AcceptMode.AcceptSave)
        if file_dialog.exec():
            file_name = file_dialog.selectedFiles()[0]
            with open(file_name, "w", encoding='utf-8') as f:
                f.write(self.chat_history.toPlainText())

    def update_presets(self):
        self.updating_presets = True
        self.presets.clear()
        self.presets.addItems(list(dict(self.config.cfg['presets']).keys()))
        self.presets.addItems(['----', '<Edit>'])
        self.updating_presets = False
        self.presets.setCurrentText(self.config.cfg.get('last_preset', 'sydney'))
        self.update_status_text('Preset list updated successfully.')

    def set_enter_mode(self, key):
        if key == "Enter":
            self.enter_mode = "Enter"
            self.enter_action.setChecked(True)
            self.ctrl_enter_action.setChecked(False)
        else:
            self.enter_mode = "Ctrl+Enter"
            self.enter_action.setChecked(False)
            self.ctrl_enter_action.setChecked(True)
        self.config.cfg['enter_mode'] = key
        self.config.save()

    def set_responding(self, responding):
        self.responding = responding
        self.send_button.setEnabled(not responding)
        self.load_button.setEnabled(not responding)
        self.chat_history.setReadOnly(responding)
        self.visual_search_button.setDisabled(responding)
        self.browse_button.setDisabled(responding)
        self.revoke_button.setDisabled(responding)
        self.document_button.setDisabled(responding)
        self.reset_button.setDisabled(responding)
        self.workspace_list_widget.setDisabled(responding)
        self.add_workspace_button.setDisabled(responding)
        self.del_workspace_button.setDisabled(responding)
        self.rename_workspace_button.setDisabled(responding)
        self.clear_workspace_button.setDisabled(responding)
        self.stop_button.setDisabled(not responding)
        self.quick_button.setDisabled(responding)

    def presets_changed(self, new_value: str):
        if self.updating_presets:
            return
        last_preset = self.config.cfg.get('last_preset', 'sydney')
        if new_value in ['----', '<Edit>']:
            self.presets.setCurrentText(last_preset)
        else:
            last_preset = new_value
        if new_value == '<Edit>':
            self.preset_window = PresetWindow(self.config, on_close=self.update_presets)
            self.preset_window.show()
            return
        self.config.cfg['last_preset'] = last_preset
        self.update_status_text('Preset changed. Click `Reset` to reset chat context.')
        self.config.save()

    def visual_search(self):
        def update_image_url(url: str):
            self.visual_search_url = url
            if url == "":
                self.visual_search_button.setText("Image")
            else:
                self.visual_search_button.setText("Image(Set)")

        self.visual_search_window = VisualSearchWindow(config=self.config, current_image_url=self.visual_search_url,
                                                       update_image_url=update_image_url)
        self.visual_search_window.show()

    def eventFilter(self, watched, event) -> bool:
        if event.type() == QEvent.WindowDeactivate or event.type() == QEvent.Close:
            self.flush_workspace()
            self.config.cfg['workspaces'] = self.workspace_dict
            self.config.cfg['last_workspace'] = self.current_workspace_name
            self.config.cfg['workspace_ix'] = self.workspace_ix
            self.config.save()
        if event.type() == QEvent.Close:
            app.exit()
        return super().eventFilter(watched, event)


if __name__ == "__main__":
    app = QApplication()
    loop = QEventLoop(app)
    asyncio.set_event_loop(loop)
    gui = SydneyWindow(Config())
    gui.installEventFilter(gui)
    gui.show()
    with loop:
        loop.run_forever()
