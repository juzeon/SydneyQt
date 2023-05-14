import asyncio
import pathlib
import signal
from typing import List

import tiktoken
from PySide6.QtGui import QTextCursor, Qt, QFont, QIcon
from PySide6.QtWidgets import (
    QApplication,
    QLabel,
    QPushButton,
    QWidget, QPlainTextEdit, QErrorMessage, QHBoxLayout, QFileDialog, QToolButton, QMenu, QSizePolicy, QVBoxLayout,
    QSplitter, QComboBox, QProgressBar, QSpacerItem, QLayout, QStatusBar
)
from qasync import QEventLoop, asyncSlot
from EdgeGPT import Chatbot

from browse_window import BrowseWindow
from document import read_pptx_text, read_pdf_text
from hyperlink_widget import HyperlinkWidget
from preset_window import PresetWindow
from setting_window import SettingWindow
from snap_window import SnapWindow
from user_input import UserInput
from config import Config

signal.signal(signal.SIGINT, signal.SIG_DFL)


class SydneyWindow(QWidget):
    def __init__(self, config: Config, parent=None):
        super().__init__(parent)
        self.snap_window = None
        self.preset_window = None
        self.setting_window = None
        self.browse_window = None
        self.updating_presets = False
        self.config = config
        self.responding = False
        self.enter_mode = "Enter"
        self.status_label = QLabel('Ready.')
        self.token_count_label = QLabel('Token Count')
        self.chat_history = QPlainTextEdit()
        self.chat_history.textChanged.connect(self.update_token_count)
        self.chat_history.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.user_input = UserInput(self, config=self.config)
        self.user_input.textChanged.connect(self.update_token_count)
        self.snap_button = QPushButton("Snap")
        self.snap_button.clicked.connect(self.snap_context)
        self.reset_button = QPushButton("Reset")
        self.reset_button.clicked.connect(self.clear_context)
        self.load_button = QPushButton("Load")
        self.load_button.clicked.connect(self.load_file)
        self.save_button = QPushButton("Save")
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

        self.setting_button = QPushButton('Settings')
        self.setting_button.clicked.connect(self.open_setting_window)

        upper_half = QWidget()
        upper_half_layout = QVBoxLayout()
        upper_half.setLayout(upper_half_layout)
        upper_half_buttons = QHBoxLayout()
        upper_half_layout.addLayout(upper_half_buttons)
        upper_half_buttons.addWidget(QLabel("Chat Context:"))
        upper_half_buttons.addStretch()
        upper_half_buttons.addWidget(self.setting_button)
        preset_label = QLabel('Preset:')
        preset_label.setStyleSheet("padding-left: 10px")
        upper_half_buttons.addWidget(preset_label)
        upper_half_buttons.addWidget(self.presets)
        action_label = QLabel('Actions:')
        action_label.setStyleSheet("padding-left: 10px")
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

        self.document_button = QPushButton('Document')
        self.document_button.clicked.connect(self.open_document)

        self.browse_button = QPushButton('Browse')
        self.browse_button.clicked.connect(self.open_browse_window)

        bottom_half_layout.addWidget(self.suggestion_widget)
        bottom_half_layout.addLayout(bottom_half_buttons)
        bottom_half_buttons.addWidget(QLabel("Follow-up User Input:"))
        bottom_half_buttons.addStretch()
        bottom_half_buttons.addWidget(self.document_button)
        bottom_half_buttons.addWidget(self.browse_button)
        bottom_half_buttons.addWidget(self.send_button)
        bottom_half_layout.addWidget(self.user_input)

        self.status_bar = QStatusBar()
        self.status_bar.addWidget(self.status_label)
        self.status_bar.addPermanentWidget(self.token_count_label)

        bottom_half_layout.addWidget(self.status_bar)

        self.splitter = QSplitter(Qt.Orientation.Vertical)
        self.splitter.addWidget(upper_half)
        self.splitter.addWidget(bottom_half)
        self.splitter.setStretchFactor(0, self.config.get('stretch_factor'))
        self.splitter.setStretchFactor(1, 1)
        layout = QVBoxLayout()
        layout.addWidget(self.splitter)
        layout.setContentsMargins(0, 0, 0, 0)
        self.setLayout(layout)

        self.resize(900, 600)
        self.setWindowTitle('SydneyQt')
        icon = QIcon('binglogo.png')
        self.setWindowIcon(icon)

        self.send_button.clicked.connect(self.send_message)
        self.clear_context()

    @asyncSlot()
    async def send_message(self):
        if self.responding:
            return
        self.set_responding(True)
        self.update_status_text('Creating conversation...')
        user_input = self.user_input.toPlainText()
        proxy = self.config.get('proxy')
        try:
            chatbot = await Chatbot.create(cookie_path="cookies.json", proxy=proxy if proxy != "" else None)
        except Exception as e:
            QErrorMessage(self).showMessage(str(e))
            self.update_status_text('Error: ' + str(e))
            self.set_responding(False)
            return
        self.user_input.clear()
        self.update_status_text('Fetching response...')

        async def stream_output():
            self.append_chat_context(f"[user](#message)\n{user_input}\n\n", new_block=True)
            wrote = 0
            async for final, response in chatbot.ask_stream(
                    prompt=user_input,
                    raw=True,
                    webpage_context=self.chat_history.toPlainText(),
                    conversation_style=self.config.cfg['conversation_style'],
                    search_result=not self.config.cfg['no_search']
            ):
                if not final and response["type"] == 1 and "messages" in response["arguments"][0]:
                    self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
                    message = response["arguments"][0]["messages"][0]
                    match message.get("messageType"):
                        case "InternalSearchQuery":
                            self.chat_history.insertPlainText(
                                f"[assistant](#search_query)\n{message['hiddenText']}\n\n")
                        case "InternalSearchResult":
                            self.chat_history.insertPlainText(
                                f"[assistant](#search_results)\n{message['hiddenText']}\n\n")
                        case None:
                            if "cursor" in response["arguments"][0]:
                                self.chat_history.insertPlainText("[assistant](#message)\n")
                                wrote = 0
                            if message.get("contentOrigin") == "Apology":
                                QErrorMessage(self).showMessage("Message revoke detected")
                                break
                            else:
                                self.chat_history.insertPlainText(message["text"][wrote:])
                                wrote = len(message["text"])
                                if "suggestedResponses" in message:
                                    suggested_responses = list(
                                        map(lambda x: x["text"], message["suggestedResponses"]))
                                    self.set_suggestion_line(suggested_responses)
                                    break
                if final and not response["item"]["messages"][-1].get("text"):
                    raise Exception("Looks like the user message has triggered the Bing filter")

        try:
            await stream_output()
        except Exception as e:
            QErrorMessage(self).showMessage(str(e))
            self.update_status_text('Error: ' + str(e))
        else:
            self.update_status_text('Ready.')
        self.set_responding(False)
        self.chat_history.moveCursor(QTextCursor.MoveOperation.End)
        await chatbot.close()

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
            on_insert=lambda context: self.append_chat_context(f"[user](#webpage_context)\n{context}\n\n"))
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
                        self.user_input.setPlainText(s)
                        self.set_suggestion_line()
                        self.send_message()

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
        self.chat_history.setPlainText(self.config.get_last_preset())
        self.set_suggestion_line()

    @asyncSlot()
    async def open_document(self):
        file_dialog = QFileDialog(self)
        file_dialog.setWindowTitle('Open a document to chat with it')
        file_dialog.setNameFilters(["Document files (*.pptx *.pdf)"])
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
                    self.append_chat_context(f'[user](#ppt_slide_context)\n{text}\n\n')
                elif ext == ".pdf":
                    text = await loop_local.run_in_executor(None, read_pdf_text, file_name)
                    self.append_chat_context(f'[user](#pdf_document_context)\n{text}\n\n')
                else:
                    QErrorMessage(self).showMessage('Unsupported file type')
            except Exception as e:
                QErrorMessage(self).showMessage(str(e))
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
        self.browse_button.setDisabled(responding)
        self.document_button.setDisabled(responding)
        self.reset_button.setDisabled(responding)

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


if __name__ == "__main__":
    app = QApplication()
    loop = QEventLoop(app)
    asyncio.set_event_loop(loop)
    gui = SydneyWindow(Config())
    gui.show()
    with loop:
        loop.run_forever()
