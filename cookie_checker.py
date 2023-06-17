import asyncio
import json
import re
from pathlib import Path

import aiohttp
from PySide6.QtWidgets import QWidget, QLabel, QVBoxLayout, QPushButton, QFileDialog, QErrorMessage

from config import Config


class CookieChecker(QWidget):
    def __init__(self, config: Config):
        super().__init__()
        self.config = config
        self.session = aiohttp.ClientSession(
            headers={
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) '
                              'AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 '
                              'Safari/537.36 Edg/114.0.1788.0'})
        self.user_label = QLabel('')
        self.set_user_label('<Checking...>')
        self.open_cookie_button = QPushButton('Open another cookies.json')
        self.open_cookie_button.clicked.connect(self.open_cookie)
        asyncio.ensure_future(self.check_cookie('cookies.json'))
        layout = QVBoxLayout()
        layout.addWidget(self.user_label)
        layout.addWidget(self.open_cookie_button)
        self.setLayout(layout)
        self.setWindowTitle('Cookie Checker')
        self.setFixedWidth(300)

    def set_user_label(self, value):
        self.user_label.setText('Current User: ' + value)

    def open_cookie(self):
        file_dialog = QFileDialog(self)
        file_dialog.setWindowTitle('Open a JSON file containing cookies')
        file_dialog.setNameFilters(["JSON files (*.json)"])
        file_dialog.setAcceptMode(QFileDialog.AcceptMode.AcceptOpen)
        if not file_dialog.exec():
            return
        filename = file_dialog.selectedFiles()[0]
        self.set_user_label('<Checking...>')
        asyncio.ensure_future(self.check_cookie(filename))

    async def check_cookie(self, path: str):
        try:
            cookie = next(filter(lambda item: item['name'] == '_U',
                                 json.loads(Path(path).read_text(encoding='utf-8'))), None)
            value = cookie['value'] if cookie is not None else None
            if not value:
                self.set_user_label('<Cookie is malformed>')
                return
            resp = await self.session.get('https://www.bing.com/search?q=Bing+AI&showconv=1',
                                          headers={'Cookie': '_U=' + value}, proxy=self.config.get('proxy'))
            resp_text = await resp.text()
            result = re.findall(r'data-clarity-mask="true" title="(.*?)"', resp_text)
            if not result:
                self.set_user_label('<Cannot get current user>')
                return
            self.set_user_label(result[0])
        except Exception as e:
            self.set_user_label('<Error occurred>')
            print(e)
            QErrorMessage(self).showMessage(str(e))
