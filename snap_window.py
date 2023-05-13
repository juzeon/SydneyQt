from PySide6.QtGui import QPageLayout, QTextCursor, QFont
from PySide6.QtWidgets import QWidget, QBoxLayout, QVBoxLayout, QPlainTextEdit

from config import Config


class SnapWindow(QWidget):
    def __init__(self, config: Config, text: str):
        super().__init__()
        self.config=config
        layout = QVBoxLayout()
        self.editor = QPlainTextEdit()
        self.editor.setPlainText(text)
        self.editor.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.editor.moveCursor(QTextCursor.MoveOperation.End)
        self.editor.verticalScrollBar().setValue(self.editor.verticalScrollBar().maximum())
        layout.addWidget(self.editor)
        self.setWindowTitle('Snapped Context')
        self.setLayout(layout)
        self.resize(850, 400)
