from PySide6.QtWidgets import QLineEdit, QDialog, QPushButton, QVBoxLayout


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
