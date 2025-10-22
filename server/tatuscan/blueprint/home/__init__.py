from flask import Blueprint

# sem template_folder — usaremos o diretório global: src/tatuscan/templates
bp = Blueprint("home", __name__)

from . import routes  # noqa: F401
