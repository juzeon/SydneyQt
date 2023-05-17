import base64
import os
import pathlib

from PySide6.QtGui import QPageLayout, QTextCursor, QFont, QDesktopServices
from PySide6.QtWebEngineCore import QWebEnginePage
from PySide6.QtWebEngineWidgets import QWebEngineView
from PySide6.QtWidgets import QWidget, QBoxLayout, QVBoxLayout, QPlainTextEdit, QTabWidget

from config import Config


class SnapWindow(QWidget):
    def __init__(self, config: Config, text: str):
        super().__init__()
        self.config = config
        layout = QVBoxLayout()
        tab_widget = QTabWidget()

        self.editor = QPlainTextEdit()
        self.editor.setPlainText(text)
        self.editor.setFont(QFont(self.config.get('font_family'), self.config.get('font_size')))
        self.editor.moveCursor(QTextCursor.MoveOperation.End)
        self.editor.verticalScrollBar().setValue(self.editor.verticalScrollBar().maximum())

        self.webview = QWebEngineView()
        html_text = pathlib.Path('assets/snap_template.html').read_text(encoding='utf-8')
        html_text = html_text.replace('##CONTENT_HERE##',
                                      base64.b64encode(text.encode('utf-8')).decode(encoding='utf-8'))
        html_text = html_text.replace('##FONT_SIZE_HERE##', str(config.get('font_size')))
        html_text = html_text.replace('##FONT_FAMILY_HERE##', config.get('font_family'))
        web_page = CustomWebEnginePage(self.webview)
        self.webview.setPage(web_page)
        self.webview.setHtml(html_text, 'file:///assets/snap_template.html')

        tab_widget.addTab(self.webview, 'Rich Text')
        tab_widget.addTab(self.editor, 'Plain Text')
        layout.addWidget(tab_widget)
        self.setWindowTitle('Snapped Context')
        self.setLayout(layout)
        self.resize(850, 700)


class CustomWebEnginePage(QWebEnginePage):
    def acceptNavigationRequest(self, url, typ, is_main_frame):
        if typ == QWebEnginePage.NavigationType.NavigationTypeLinkClicked:
            QDesktopServices.openUrl(url)
            return False
        else:
            return super().acceptNavigationRequest(url, typ, is_main_frame)
