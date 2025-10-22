"""Charts routes - Data visualization endpoints."""
import logging

from flask import render_template, request

from tatuscan.services import InventoryService
from . import bp

logger = logging.getLogger(__name__)


@bp.route("/")
def charts_index():
    """Display charts with OS, version, and age distribution."""
    try:
        top_n = int(request.args.get("top", 8))

        # Get OS distribution
        os_data = InventoryService.get_os_distribution()
        os_labels = [item["os"] for item in os_data]
        os_values = [item["count"] for item in os_data]

        # Get version distribution
        ver_data = InventoryService.get_version_distribution(top_n=top_n)
        ver_labels = [item["version"] for item in ver_data]
        ver_values = [item["count"] for item in ver_data]

        # Get age distribution
        age_data = InventoryService.get_age_distribution()
        age_labels = [item["range"] for item in age_data]
        age_values = [item["count"] for item in age_data]

        return render_template(
            "charts/index.html",
            os_labels=os_labels,
            os_values=os_values,
            ver_labels=ver_labels,
            ver_values=ver_values,
            age_labels=age_labels,
            age_values=age_values,
            top_n=top_n,
        )
    except Exception as e:
        logger.error(f"Error rendering charts: {str(e)}")
        return render_template(
            "charts/index.html",
            os_labels=[], os_values=[],
            ver_labels=[], ver_values=[],
            age_labels=[], age_values=[],
            top_n=8
        )
