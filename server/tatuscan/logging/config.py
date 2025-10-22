"""Centralized logging configuration."""
import logging
import sys
from flask import Flask


def setup_logging(app: Flask) -> None:
    """
    Configure logging for the application.

    Args:
        app: Flask application instance
    """
    # Determine log level from config or default to INFO
    log_level_name = app.config.get("LOG_LEVEL", "INFO").upper()
    log_level = getattr(logging, log_level_name, logging.INFO)

    # Configure root logger
    logging.basicConfig(
        level=log_level,
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
        handlers=[
            logging.StreamHandler(sys.stdout)
        ]
    )

    # Set Flask's logger level
    app.logger.setLevel(log_level)

    # Reduce noise from libraries
    logging.getLogger("werkzeug").setLevel(logging.WARNING)
    logging.getLogger("sqlalchemy.engine").setLevel(logging.WARNING)

    app.logger.info(f"Logging configured with level: {log_level_name}")
