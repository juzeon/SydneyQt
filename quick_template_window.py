from PySide6.QtGui import QFont
from PySide6.QtWidgets import QWidget, QLabel, QPushButton, QVBoxLayout, QPlainTextEdit, QHBoxLayout

from config import Config


class QuickTemplateWindow(QWidget):
    def __init__(self, config: Config, on_save: callable = None):
        super().__init__()
        self.config = config
        self.on_save = on_save
        header_label = QLabel(
            'Note: One template a line. \nUse `\\n` if the template content itself contains linebreaks.')
        self.save_button = QPushButton('Save')
        self.save_button.setDisabled(True)
        self.save_button.clicked.connect(self.save)
        self.template_editor = QPlainTextEdit()
        self.template_editor.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.template_editor.setPlainText('\n'.join(self.config.get('quick')))
        self.template_editor.textChanged.connect(lambda: self.save_button.setDisabled(False))
        bottom_layout = QHBoxLayout()
        bottom_layout.addStretch()
        bottom_layout.addWidget(self.save_button)
        main_layout = QVBoxLayout()
        main_layout.addWidget(header_label)
        main_layout.addWidget(self.template_editor)
        main_layout.addLayout(bottom_layout)
        self.setLayout(main_layout)

        self.setWindowTitle('Templates for Quick Send')
        self.resize(850, 400)

    def save(self):
        self.config.cfg['quick'] = [text for text in self.template_editor.toPlainText().split('\n') if text != '']
        self.config.save()
        self.save_button.setDisabled(True)
        if self.on_save is not None:
            self.on_save()
