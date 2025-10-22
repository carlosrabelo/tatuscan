from flask import Blueprint

# Blueprint dedicado ao relatório
bp = Blueprint("report", __name__, url_prefix="/report")

from . import routes  # noqa: F401
