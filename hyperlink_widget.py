import html

from PySide6.QtCore import Qt, Slot
from PySide6.QtWidgets import QWidget, QLabel, QSizePolicy


class HyperlinkWidget(QWidget):
    def __init__(self, text: str, on_clicked: callable = None):
        super().__init__()
        self.text = text
        self.on_clicked = on_clicked
        self.label = QLabel(self)
        self.label.setTextFormat(Qt.RichText)
        self.label.setText('<a href="#">' + html.escape(text) + '</a>')
        self.label.linkActivated.connect(self.on_link_clicked)
        self.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Preferred)

    @Slot(str)
    def on_link_clicked(self):
        if self.on_clicked is not None:
            self.on_clicked()
