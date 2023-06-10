import json
import aiohttp
from PySide6.QtWidgets import QWidget, QVBoxLayout, QHBoxLayout, QLineEdit, QPushButton, QLabel, QPlainTextEdit, \
    QErrorMessage
from bs4 import BeautifulSoup
from qasync import asyncSlot

from config import Config


class BrowseWindow(QWidget):
    def __init__(self, config: Config, on_insert: callable([str]) = None):
        super().__init__()
        self.config = config
        self.on_insert = on_insert

        self.session = aiohttp.ClientSession(
            headers={'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) '
                                   'Gecko/20100101 Firefox/113.0'})

        self.url_edit = QLineEdit()
        self.url_edit.returnPressed.connect(self.fetch_button_clicked)
        self.fetch_button = QPushButton('Fetch')
        self.fetch_button.clicked.connect(self.fetch_button_clicked)

        url_layout = QHBoxLayout()
        url_layout.addWidget(QLabel('URL:'))
        url_layout.addWidget(self.url_edit)
        url_layout.addWidget(self.fetch_button)

        self.webpage_context_edit = QPlainTextEdit()

        self.insert_button = QPushButton('Insert')
        self.insert_button.clicked.connect(self.insert_button_clicked)

        bottom_layout = QHBoxLayout()
        bottom_layout.addStretch()
        bottom_layout.addWidget(self.insert_button)

        layout = QVBoxLayout()
        layout.addLayout(url_layout)
        layout.addWidget(self.webpage_context_edit)
        layout.addLayout(bottom_layout)
        self.setLayout(layout)

        self.setWindowTitle('Browse URL')
        self.resize(850, 400)

    @asyncSlot()
    async def fetch_button_clicked(self):
        self.set_responding(True)
        try:
            context = await self.fetch_webpage(self.url_edit.text())
            self.webpage_context_edit.setPlainText(context)
        except Exception as e:
            QErrorMessage(self).showMessage(str(e))
        self.set_responding(False)

    async def fetch_webpage(self, url: str) -> str:
        html = await self.session.get(url, proxy=self.config.get('proxy') if self.config.get('proxy') != '' else None)
        soup = BeautifulSoup(await html.text(), features="html.parser")
        for script in soup(["script", "style"]):
            script.extract()
        text = soup.get_text()
        lines = (line.strip() for line in text.splitlines())
        chunks = (phrase.strip() for line in lines for phrase in line.split("  "))
        text = '\n'.join(chunk for chunk in chunks if chunk)
        return json.dumps(text, ensure_ascii=False)

    def set_responding(self, responding: bool):
        self.fetch_button.setDisabled(responding)
        self.insert_button.setDisabled(responding)

    def insert_button_clicked(self):
        if self.on_insert is not None:
            self.on_insert(self.webpage_context_edit.toPlainText())
        self.close()
