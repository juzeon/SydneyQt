import PySide6
from PySide6.QtCore import Slot
from PySide6.QtGui import QFont
from PySide6.QtWidgets import QWidget, QListWidget, QListWidgetItem, QPlainTextEdit, QPushButton, QVBoxLayout, \
    QHBoxLayout, QDialog, QLineEdit

from config import Config


class PresetWindow(QWidget):
    def __init__(self, config: Config, on_close: callable = None):
        super().__init__()
        self.on_close = on_close
        self.config = config
        self.setWindowTitle('Manage Presets')
        self.resize(550, 400)
        self.list = QListWidget()
        for name in dict(self.config.cfg['presets']).keys():
            QListWidgetItem(name, self.list)

        self.editor = QPlainTextEdit()
        self.editor.setFont(QFont("Microsoft YaHei", 11))
        self.editor.textChanged.connect(self.editor_text_changed)

        self.add_button = QPushButton('Add')
        self.rename_button = QPushButton('Rename')
        self.delete_button = QPushButton('Delete')
        self.add_button.setFixedWidth(50)
        self.rename_button.setFixedWidth(50)
        self.delete_button.setFixedWidth(50)
        self.add_button.clicked.connect(self.add_button_clicked)
        self.rename_button.clicked.connect(self.rename_button_clicked)
        self.delete_button.clicked.connect(self.delete_button_clicked)

        self.save_editor_button = QPushButton('Save')
        self.save_editor_button.setFixedWidth(50)
        self.save_editor_button.clicked.connect(self.save_editor_button_clicked)

        left_layout = QVBoxLayout()
        left_layout.addWidget(self.list)
        left_bottom_layout = QHBoxLayout()
        left_bottom_layout.addWidget(self.add_button)
        left_bottom_layout.addWidget(self.rename_button)
        left_bottom_layout.addWidget(self.delete_button)
        left_layout.addLayout(left_bottom_layout)

        right_layout = QVBoxLayout()
        right_layout.addWidget(self.editor)
        right_bottom_layout = QHBoxLayout()
        right_bottom_layout.addStretch()
        right_bottom_layout.addWidget(self.save_editor_button)
        right_layout.addLayout(right_bottom_layout)

        layout = QHBoxLayout()
        layout.addLayout(left_layout, 1)
        layout.addLayout(right_layout, 4)
        self.setLayout(layout)

        self.list.currentItemChanged.connect(self.list_item_changed)
        self.list.setCurrentRow(0)

    def closeEvent(self, event: PySide6.QtGui.QCloseEvent) -> None:
        if self.on_close is not None:
            self.on_close()

    @Slot()
    def rename_button_clicked(self):
        item = self.list.currentItem()
        dialog = NameDialog(name=item.text())
        result = dialog.exec()
        if result == 0:
            return
        name = dialog.get_name()
        if name == "":
            return
        preset = self.config.cfg['presets'][item.text()]
        self.config.cfg['presets'][name] = preset
        del self.config.cfg['presets'][item.text()]
        if self.config.cfg['last_preset'] == item.text():
            self.config.cfg['last_preset'] = name
        self.config.save()
        item.setText(name)

    @Slot()
    def add_button_clicked(self):
        name = 'new-preset'
        for i in range(2, 10 ** 10):
            if name in self.config.cfg['presets']:
                name = f'new-preset-{i}'
            else:
                break
        self.config.cfg['presets'][name] = '[system](#additional_instructions)\n\n\n'
        self.config.save()
        QListWidgetItem(name, self.list)

    @Slot()
    def delete_button_clicked(self):
        item = self.list.currentItem()
        del self.config.cfg['presets'][item.text()]
        if self.config.cfg['last_preset'] == item.text():
            self.config.cfg['last_preset'] = 'sydney'
        self.config.save()
        self.list.takeItem(self.list.currentRow())

    @Slot()
    def editor_text_changed(self):
        self.save_editor_button.setDisabled(False)

    @Slot()
    def save_editor_button_clicked(self):
        self.config.cfg['presets'][self.list.currentItem().text()] = self.editor.toPlainText().strip() + "\n\n"
        self.config.save()
        self.save_editor_button.setDisabled(True)

    @Slot()
    def list_item_changed(self, item):
        if item.text() == 'sydney':
            self.rename_button.setDisabled(True)
            self.delete_button.setDisabled(True)
        else:
            self.rename_button.setDisabled(False)
            self.delete_button.setDisabled(False)
        self.editor.setPlainText(self.config.cfg['presets'][item.text()])
        self.save_editor_button.setDisabled(True)


class NameDialog(QDialog):
    def __init__(self, parent=None, name=''):
        super().__init__(parent)
        self.name_edit = QLineEdit()
        self.name_edit.setText(name)
        self.ok_button = QPushButton("OK")
        self.ok_button.clicked.connect(self.accept)
        self.layout = QVBoxLayout()
        self.layout.addWidget(self.name_edit)
        self.layout.addWidget(self.ok_button)
        self.setLayout(self.layout)
        self.setWindowTitle("Enter a name")

    def get_name(self):
        return self.name_edit.text()
