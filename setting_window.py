from PySide6.QtCore import Slot
from PySide6.QtGui import QCloseEvent, QFont
from PySide6.QtWidgets import QWidget, QFormLayout, QLabel, QLineEdit, QCheckBox, QVBoxLayout, QHBoxLayout, QPushButton, \
    QMessageBox, QComboBox, QFontDialog, QSpinBox

from config import Config


class SettingWindow(QWidget):
    def __init__(self, config: Config, on_close: callable = None):
        super().__init__()
        self.on_close = on_close
        self.config = config
        self.font = QFont(config.get('font_family'), config.get('font_size'))

        self.proxy_edit = QLineEdit()
        self.proxy_edit.setToolTip('Example: socks5h://127.0.0.1:7890')
        self.conversation_style = QComboBox()
        self.conversation_style.addItems(["creative", "balanced", "precise"])
        self.conversation_style.setToolTip('Balanced mode uses GPT-3.5, '
                                           'while creative mode and precise mode use GPT-4.')
        self.no_suggestion_checkbox = QCheckBox()
        self.no_suggestion_checkbox.setToolTip('Do no show suggestion links if enabled. '
                                               'Note that suggestions will still be generated in the background.')
        self.no_search_checkbox = QCheckBox()
        self.no_search_checkbox.setToolTip('Do no fill search result into the context if enabled. '
                                           'Note that web search will still be performed in the background.')
        self.font_button = QPushButton('Select')
        self.font_button.setToolTip('Open a dialog to adjust font settings.')
        self.font_button.clicked.connect(self.open_font_dialog)
        self.stretch_factor = QSpinBox()
        self.stretch_factor.setToolTip('Adjust the height of the chat context box.')
        self.stretch_factor.setMinimum(1)
        self.stretch_factor.setMaximum(10)
        self.revoke_text = QLineEdit()
        self.revoke_text.setToolTip('Show this text as a clickable suggestion on message revoke.')
        self.revoke_count = QSpinBox()
        self.revoke_count.setMinimum(0)
        self.revoke_count.setToolTip('Maximum count for auto-reply on message revoke. \n'
                                     'If `Suggestion on Message Revoke` is available, send it automatically. \n'
                                     'The error dialog will not show if an auto-reply can be applied. \n'
                                     'Set this to 0 to disable and show a suggestion only.')

        form_layout = QFormLayout()
        form_layout.addRow(QLabel('Proxy:'), self.proxy_edit)
        form_layout.addRow(QLabel('Conversation Style:'), self.conversation_style)
        form_layout.addRow(QLabel('No Suggestion:'), self.no_suggestion_checkbox)
        form_layout.addRow(QLabel('No Search Result:'), self.no_search_checkbox)
        form_layout.addRow(QLabel('Font Family and Size:'), self.font_button)
        form_layout.addRow(QLabel('(*) Stretch Factor of Chat Context Box: '), self.stretch_factor)
        form_layout.addRow(QLabel('Suggestion on Message Revoke: '), self.revoke_text)
        form_layout.addRow(QLabel('Revoke Auto Reply Count: '), self.revoke_count)

        self.save_button = QPushButton('Save')
        self.save_button.clicked.connect(self.save_config)

        layout = QVBoxLayout()
        layout.addWidget(QLabel('Note: Hover to show tooltips of options. \n'
                                'Settings marked with (*) need a restart to be applied.'))
        layout.addLayout(form_layout)
        bottom_layout = QHBoxLayout()
        bottom_layout.addStretch()
        bottom_layout.addWidget(self.save_button)
        layout.addLayout(bottom_layout)

        self.setLayout(layout)
        self.setFixedWidth(400)
        self.setWindowTitle('Settings')

        self.render_config()

    def render_config(self):
        self.proxy_edit.setText(self.config.cfg['proxy'])
        self.conversation_style.setCurrentText(self.config.cfg['conversation_style'])
        self.no_suggestion_checkbox.setChecked(self.config.cfg['no_suggestion'])
        self.no_search_checkbox.setChecked(self.config.cfg['no_search'])
        self.stretch_factor.setValue(self.config.get('stretch_factor'))
        self.revoke_text.setText(self.config.get('revoke_reply_text'))
        self.revoke_count.setValue(self.config.get('revoke_reply_count'))

    @Slot()
    def open_font_dialog(self):
        font_dialog = QFontDialog()
        font_dialog.setCurrentFont(self.font)
        result = font_dialog.exec()
        if result == 1:
            self.font = font_dialog.currentFont()

    @Slot()
    def save_config(self):
        self.config.cfg['proxy'] = self.proxy_edit.text()
        self.config.cfg['conversation_style'] = self.conversation_style.currentText()
        self.config.cfg['no_suggestion'] = self.no_suggestion_checkbox.isChecked()
        self.config.cfg['no_search'] = self.no_search_checkbox.isChecked()
        self.config.cfg['font_family'] = self.font.family()
        self.config.cfg['font_size'] = self.font.pointSize()
        self.config.cfg['stretch_factor'] = self.stretch_factor.value()
        self.config.cfg['revoke_reply_text'] = self.revoke_text.text()
        self.config.cfg['revoke_reply_count'] = self.revoke_count.value()
        self.config.save()
        msg = QMessageBox()
        msg.setIcon(QMessageBox.Icon.Information)
        msg.setText('Settings saved.\nSettings marked with (*) need a restart to be applied.')
        msg.exec()

    def closeEvent(self, event: QCloseEvent) -> None:
        if self.on_close is not None:
            self.on_close()
