from PySide6.QtWidgets import QDialog, QVBoxLayout, QPushButton, QFileDialog, QLabel, QLineEdit
from qasync import asyncSlot

import EdgeGPT


class VisualSearchWindow(QDialog):
    def __init__(self, parent=None, *args, **kwargs):
        super(VisualSearchWindow, self).__init__(parent, *args, **kwargs)
        self.parent = parent

        self.layout = QVBoxLayout()

        self.url_input = QLineEdit(text=self.parent.visual_search_url)
        self.layout.addWidget(self.url_input)
        self.url_input.textChanged.connect(self.url_input_changed)

        self.file_button = QPushButton('Choose and Upload Image')
        self.file_button.clicked.connect(self.file_button_clicked)
        self.layout.addWidget(self.file_button)

        self.setLayout(self.layout)

    def url_input_changed(self):
        self.parent.visual_search_url = self.url_input.text()

    @asyncSlot()
    async def file_button_clicked(self):
        file_dialog = QFileDialog()
        file_dialog.setFileMode(QFileDialog.FileMode.ExistingFile)
        file_dialog.setNameFilters(["Image files (*.gif *.jfif *.pjpeg *.jpeg *.pjp *.jpg *.png *.webp)"])

        if file_dialog.exec_():
            file_path = file_dialog.selectedFiles()[0]
            self.url_input.setText("https://www.bing.com/images/blob?bcid=" +
                                   await EdgeGPT.upload_image(file_path, proxy=self.parent.config.get('proxy')))
