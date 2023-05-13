from PySide6.QtCore import Qt
from PySide6.QtGui import QFont
from PySide6.QtWidgets import QPlainTextEdit

from config import Config


class UserInput(QPlainTextEdit):
    def __init__(self, parent, config: Config):
        super().__init__(parent)
        self.parent = parent
        self.config = config
        self.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))

    def keyPressEvent(self, event):
        key = event.key()
        modifiers = event.modifiers()
        if key == Qt.Key.Key_Enter or key == Qt.Key.Key_Return:
            match self.parent.enter_mode:
                case "Enter":
                    if modifiers == Qt.KeyboardModifier.NoModifier:
                        self.parent.send_message()
                    else:
                        super().keyPressEvent(event)
                        self.insertPlainText("\n")
                case "Ctrl+Enter":
                    if modifiers == Qt.KeyboardModifier.ControlModifier:
                        self.parent.send_message()
                    else:
                        super().keyPressEvent(event)
        else:
            super().keyPressEvent(event)
