from PySide6.QtCore import Slot
from PySide6.QtGui import QCloseEvent, QFont
from PySide6.QtWidgets import QWidget, QFormLayout, QLabel, QLineEdit, QCheckBox, QVBoxLayout, QHBoxLayout, QPushButton, \
    QMessageBox, QComboBox, QFontDialog

from config import Config


class SettingWindow(QWidget):
    def __init__(self, config: Config, on_close: callable = None):
        super().__init__()
        self.on_close = on_close
        self.config = config
        self.font = QFont(config.get('font_family'), config.get('font_size'))

        self.proxy_edit = QLineEdit()
        self.conversation_style = QComboBox()
        self.conversation_style.addItems(["creative", "balanced", "precise"])
        self.no_suggestion_checkbox = QCheckBox()
        self.no_search_checkbox = QCheckBox()
        self.no_search_checkbox.setToolTip('''Do no fill search result into the context if enabled. 
Note that web search will still be performed in the background.''')
        self.font_button = QPushButton('Select')
        self.font_button.clicked.connect(self.open_font_dialog)

        form_layout = QFormLayout()
        form_layout.addRow(QLabel('Proxy:'), self.proxy_edit)
        form_layout.addRow(QLabel('Conversation Style:'), self.conversation_style)
        form_layout.addRow(QLabel('No Suggestion:'), self.no_suggestion_checkbox)
        form_layout.addRow(QLabel('No Search Result:'), self.no_search_checkbox)
        form_layout.addRow(QLabel('(*) Font Family and Size:'), self.font_button)

        self.save_button = QPushButton('Save')
        self.save_button.clicked.connect(self.save_config)

        layout = QVBoxLayout()
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
        self.config.save()
        msg = QMessageBox()
        msg.setIcon(QMessageBox.Icon.Information)
        msg.setText('Settings saved.\nSettings marked with (*) need a restart to be applied.')
        msg.exec()

    def closeEvent(self, event: QCloseEvent) -> None:
        if self.on_close is not None:
            self.on_close()
