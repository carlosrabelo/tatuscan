"""Error handlers for Flask application."""
import logging
from flask import Flask, jsonify
from werkzeug.exceptions import HTTPException

from tatuscan.services.exceptions import (
    ServiceException,
    ValidationError,
    NotFoundError,
    DatabaseError,
)

logger = logging.getLogger(__name__)


def register_error_handlers(app: Flask) -> None:
    """
    Register error handlers for the Flask application.

    Args:
        app: Flask application instance
    """

    @app.errorhandler(ValidationError)
    def handle_validation_error(error: ValidationError):
        """Handle validation errors."""
        logger.warning(f"Validation error: {error.message}")
        return jsonify({
            "error": error.message,
            "missing_fields": error.missing_fields if hasattr(error, "missing_fields") else []
        }), error.status_code

    @app.errorhandler(NotFoundError)
    def handle_not_found_error(error: NotFoundError):
        """Handle not found errors."""
        logger.info(f"Not found: {error.message}")
        return jsonify({"error": error.message}), error.status_code

    @app.errorhandler(DatabaseError)
    def handle_database_error(error: DatabaseError):
        """Handle database errors."""
        logger.error(f"Database error: {error.message}", exc_info=error.original_error)
        return jsonify({"error": error.message}), error.status_code

    @app.errorhandler(ServiceException)
    def handle_service_exception(error: ServiceException):
        """Handle generic service exceptions."""
        logger.error(f"Service error: {error.message}")
        return jsonify({"error": error.message}), error.status_code

    @app.errorhandler(HTTPException)
    def handle_http_exception(error: HTTPException):
        """Handle HTTP exceptions."""
        logger.warning(f"HTTP error {error.code}: {error.description}")
        return jsonify({
            "error": error.description,
            "code": error.code
        }), error.code

    @app.errorhandler(Exception)
    def handle_generic_exception(error: Exception):
        """Handle unexpected exceptions."""
        logger.exception("Unexpected error occurred")
        return jsonify({
            "error": "An unexpected error occurred",
            "details": str(error) if app.debug else None
        }), 500
