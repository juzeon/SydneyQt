from PySide6.QtWidgets import QDialog, QVBoxLayout, QPushButton, QFileDialog, QLabel, QLineEdit, QHBoxLayout, \
    QErrorMessage
from qasync import asyncSlot

import sydney
from config import Config


class VisualSearchWindow(QDialog):
    def __init__(self, config: Config, current_image_url: str, update_image_url: callable):
        super().__init__()
        self.config = config
        self.update_image_url = update_image_url

        self.layout = QVBoxLayout()

        header_layout = QHBoxLayout()
        header_layout.addWidget(QLabel('Current Image URL:'))
        self.url_input = QLineEdit(text=current_image_url)
        self.layout.addLayout(header_layout)
        self.layout.addWidget(self.url_input)

        self.file_button = QPushButton('Choose')
        self.file_button.clicked.connect(self.file_button_clicked)
        self.clear_button = QPushButton('Clear')
        self.clear_button.clicked.connect(self.clear_action)
        self.ok_button = QPushButton('OK')
        self.ok_button.clicked.connect(self.ok_action)
        self.cancel_button = QPushButton('Cancel')
        self.cancel_button.clicked.connect(self.cancel_action)
        actions_layout = QHBoxLayout()
        actions_layout.addStretch()
        actions_layout.addWidget(self.file_button)
        actions_layout.addWidget(self.clear_button)
        actions_layout.addWidget(self.ok_button)
        actions_layout.addWidget(self.cancel_button)
        self.layout.addLayout(actions_layout)

        self.setLayout(self.layout)
        self.setFixedWidth(400)

    def clear_action(self):
        self.url_input.setText('')

    def ok_action(self):
        self.update_image_url(self.url_input.text())
        self.close()

    def cancel_action(self):
        self.close()

    def url_input_changed(self):
        self.update_image_url(self.url_input.text())

    @asyncSlot()
    async def file_button_clicked(self):
        self.file_button.setDisabled(True)
        file_dialog = QFileDialog()
        file_dialog.setFileMode(QFileDialog.FileMode.ExistingFile)
        file_dialog.setNameFilters(["Image files (*.gif *.jfif *.pjpeg *.jpeg *.pjp *.jpg *.png *.webp)"])
        try:
            if file_dialog.exec_():
                file_path = file_dialog.selectedFiles()[0]
                self.url_input.setText("https://www.bing.com/images/blob?bcid=" +
                                       await sydney.upload_image(file_path, proxy=self.config.get('proxy')))
        except Exception as e:
            print(e)
            QErrorMessage(self).showMessage(str(e))
        finally:
            self.file_button.setDisabled(False)
