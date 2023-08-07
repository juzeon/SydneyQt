from PySide6.QtCore import Slot
from PySide6.QtGui import QCloseEvent, QFont
from PySide6.QtWidgets import QWidget, QFormLayout, QLabel, QLineEdit, QCheckBox, QVBoxLayout, QHBoxLayout, QPushButton, \
    QMessageBox, QComboBox, QFontDialog, QSpinBox, QDoubleSpinBox

from config import Config


class SettingWindow(QWidget):
    def __init__(self, config: Config, on_close: callable = None):
        super().__init__()
        self.on_close = on_close
        self.config = config
        self.font = QFont(config.get('font_family'), config.get('font_size'))

        self.wss_domain = QLineEdit()
        self.wss_domain.setToolTip('Domain for Sydney Websocket, without any prefixes or suffixes.\n'
                                   'Default: sydney.bing.com')
        self.proxy_edit = QLineEdit()
        self.proxy_edit.setToolTip('Example: socks5h://127.0.0.1:7890')
        self.dark_mode = QCheckBox()
        self.dark_mode.setToolTip('Enable dark mode.')
        self.no_suggestion_checkbox = QCheckBox()
        self.no_suggestion_checkbox.setToolTip('Do no show suggestion links if enabled. '
                                               'Note that suggestions will still be generated in the background.')
        self.no_search_checkbox = QCheckBox()
        self.no_search_checkbox.setToolTip('Answer without searching the web under all circumstances.')
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
        self.direct_quick = QCheckBox()
        self.direct_quick.setToolTip('Whether to send quick responses straightforward if the user input is empty.')

        top_form_layout = QFormLayout()
        top_form_layout.addRow(QLabel('Wss Domain:'), self.wss_domain)
        top_form_layout.addRow(QLabel('Proxy:'), self.proxy_edit)
        top_form_layout.addRow(QLabel('(*) Dark Mode:'), self.dark_mode)
        top_form_layout.addRow(QLabel('No Suggestion:'), self.no_suggestion_checkbox)
        top_form_layout.addRow(QLabel('No Search Result:'), self.no_search_checkbox)
        top_form_layout.addRow(QLabel('Font Family and Size:'), self.font_button)
        top_form_layout.addRow(QLabel('(*) Stretch Factor of Chat Context Box: '), self.stretch_factor)
        top_form_layout.addRow(QLabel('Suggestion on Message Revoke: '), self.revoke_text)
        top_form_layout.addRow(QLabel('Revoke Auto Reply Count: '), self.revoke_count)
        top_form_layout.addRow(QLabel('Send Quick Responses Straightforward: '), self.direct_quick)

        self.openai_key = QLineEdit()
        self.openai_key.setToolTip('OpenAI API Key.')
        self.openai_endpoint = QLineEdit()
        self.openai_endpoint.setToolTip('OpenAI API Endpoint. Must be started with https:// and ended with /v1.')
        self.openai_short_model = QLineEdit()
        self.openai_short_model.setToolTip('Model for shorter conversations. See\n'
                                           'https://platform.openai.com/docs/models/model-endpoint-compatibility\n'
                                           'for all models.')
        self.openai_long_model = QLineEdit()
        self.openai_long_model.setToolTip('Model for longer conversations. See\n'
                                          'https://platform.openai.com/docs/models/model-endpoint-compatibility\n'
                                          'for all models.')
        self.openai_threshold = QSpinBox()
        self.openai_threshold.setMinimum(0)
        self.openai_threshold.setMaximum(100000)
        self.openai_threshold.setToolTip('Threshold of token length to switch between '
                                         'the short model and the long model.')
        self.openai_temperature = QDoubleSpinBox()
        self.openai_temperature.setMinimum(0)
        self.openai_temperature.setMaximum(2)
        self.openai_temperature.setSingleStep(0.1)
        self.openai_temperature.setToolTip('Temperature for the model.')

        bottom_form_layout = QFormLayout()
        bottom_form_layout.addRow(QLabel('OpenAI Key:'), self.openai_key)
        bottom_form_layout.addRow(QLabel('OpenAI Endpoint:'), self.openai_endpoint)
        bottom_form_layout.addRow(QLabel('Short Model:'), self.openai_short_model)
        bottom_form_layout.addRow(QLabel('Long Model:'), self.openai_long_model)
        bottom_form_layout.addRow(QLabel('Model Switching Threshold:'), self.openai_threshold)
        bottom_form_layout.addRow(QLabel('Model Temperature:'), self.openai_temperature)

        self.save_button = QPushButton('Save')
        self.save_button.clicked.connect(self.save_config)

        layout = QVBoxLayout()
        layout.addWidget(QLabel('Note: Hover to show tooltips of options. \n'
                                'Settings marked with (*) need a restart to be applied.'))
        layout.addLayout(top_form_layout)
        openai_label = QLabel('ChatGPT specified settings:')
        openai_label.setStyleSheet('padding-top: 5px;')
        layout.addWidget(openai_label)
        layout.addLayout(bottom_form_layout)
        bottom_layout = QHBoxLayout()
        bottom_layout.addStretch()
        bottom_layout.addWidget(self.save_button)
        layout.addLayout(bottom_layout)

        self.setLayout(layout)
        self.setFixedWidth(400)
        self.setWindowTitle('Settings')

        self.render_config()

    def render_config(self):
        self.wss_domain.setText(self.config.get('wss_domain'))
        self.proxy_edit.setText(self.config.cfg['proxy'])
        self.dark_mode.setChecked(self.config.get('dark_mode'))
        self.no_suggestion_checkbox.setChecked(self.config.cfg['no_suggestion'])
        self.no_search_checkbox.setChecked(self.config.cfg['no_search'])
        self.stretch_factor.setValue(self.config.get('stretch_factor'))
        self.revoke_text.setText(self.config.get('revoke_reply_text'))
        self.revoke_count.setValue(self.config.get('revoke_reply_count'))
        self.direct_quick.setChecked(self.config.get('direct_quick'))
        self.openai_key.setText(self.config.get('openai_key'))
        self.openai_endpoint.setText(self.config.get('openai_endpoint'))
        self.openai_short_model.setText(self.config.get('openai_short_model'))
        self.openai_long_model.setText(self.config.get('openai_long_model'))
        self.openai_threshold.setValue(self.config.get('openai_threshold'))
        self.openai_temperature.setValue(self.config.get('openai_temperature'))

    @Slot()
    def open_font_dialog(self):
        font_dialog = QFontDialog()
        font_dialog.setCurrentFont(self.font)
        result = font_dialog.exec()
        if result == 1:
            self.font = font_dialog.currentFont()

    @Slot()
    def save_config(self):
        self.config.cfg['wss_domain'] = self.wss_domain.text()
        self.config.cfg['proxy'] = self.proxy_edit.text()
        self.config.cfg['dark_mode'] = self.dark_mode.isChecked()
        self.config.cfg['no_suggestion'] = self.no_suggestion_checkbox.isChecked()
        self.config.cfg['no_search'] = self.no_search_checkbox.isChecked()
        self.config.cfg['font_family'] = self.font.family()
        self.config.cfg['font_size'] = self.font.pointSize()
        self.config.cfg['stretch_factor'] = self.stretch_factor.value()
        self.config.cfg['revoke_reply_text'] = self.revoke_text.text()
        self.config.cfg['revoke_reply_count'] = self.revoke_count.value()
        self.config.cfg['direct_quick'] = self.direct_quick.isChecked()
        self.config.cfg['openai_key'] = self.openai_key.text()
        self.config.cfg['openai_endpoint'] = self.openai_endpoint.text()
        self.config.cfg['openai_short_model'] = self.openai_short_model.text()
        self.config.cfg['openai_long_model'] = self.openai_long_model.text()
        self.config.cfg['openai_threshold'] = self.openai_threshold.value()
        self.config.cfg['openai_temperature'] = self.openai_temperature.value()
        self.config.save()
        msg = QMessageBox()
        msg.setIcon(QMessageBox.Icon.Information)
        msg.setText('Settings saved.\nSettings marked with (*) need a restart to be applied.')
        msg.exec()

    def closeEvent(self, event: QCloseEvent) -> None:
        if self.on_close is not None:
            self.on_close()
