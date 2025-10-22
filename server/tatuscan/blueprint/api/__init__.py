from flask import Blueprint

# Prefixo /api
bp = Blueprint("api", __name__, url_prefix="/api")

from . import routes  # noqa: F401
