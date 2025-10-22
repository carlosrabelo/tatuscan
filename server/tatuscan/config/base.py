"""Base configuration."""
import os
from pathlib import Path


def _get_database_uri() -> str:
    """
    Get database URI from environment or construct SQLite path.

    Returns:
        Database URI string
    """
    # Se SQLALCHEMY_DATABASE_URI definido, usa direto
    uri = os.getenv("SQLALCHEMY_DATABASE_URI")
    if uri:
        return uri

    # Caso contrário, monta caminho SQLite
    db_dir = os.getenv("TATUSCAN_DB_DIR", "/data")
    db_file = os.getenv("TATUSCAN_DB_FILE", "tatuscan.db")

    # Garante que o diretório existe (ignora erros de permissão)
    try:
        Path(db_dir).mkdir(parents=True, exist_ok=True)
    except (PermissionError, OSError):
        pass

    db_path = Path(db_dir) / db_file
    return f"sqlite:///{db_path}"


class Config:
    """Base configuration class."""

    # Flask
    SECRET_KEY = os.getenv("SECRET_KEY", "dev-secret-key-change-in-production")

    # Database
    SQLALCHEMY_DATABASE_URI = _get_database_uri()
    SQLALCHEMY_TRACK_MODIFICATIONS = False

    # Timezone
    TIMEZONE = os.getenv("TIMEZONE", "America/Cuiaba")

    # Logging
    LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
