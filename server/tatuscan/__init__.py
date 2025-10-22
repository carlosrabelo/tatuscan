"""TatuScan application factory."""
from flask import Flask

from tatuscan.config import Config
from tatuscan.extensions import db
from tatuscan.logging import setup_logging
from tatuscan.errors import register_error_handlers


def create_app() -> Flask:
    """
    Create and configure Flask application.

    Returns:
        Configured Flask application instance
    """
    app = Flask(__name__)
    app.config.from_object(Config)

    # Configure logging
    setup_logging(app)

    # Initialize extensions
    db.init_app(app)

    # Register error handlers
    register_error_handlers(app)

    # Register blueprints
    from tatuscan.blueprint.home import bp as home_bp
    app.register_blueprint(home_bp)

    from tatuscan.blueprint.api import bp as api_bp
    app.register_blueprint(api_bp)

    from tatuscan.blueprint.report import bp as report_bp
    app.register_blueprint(report_bp)

    from tatuscan.blueprint.charts import bp as charts_bp
    app.register_blueprint(charts_bp)

    # Create database tables (development)
    with app.app_context():
        db.create_all()

    app.logger.info("TatuScan application initialized successfully")
    return app
