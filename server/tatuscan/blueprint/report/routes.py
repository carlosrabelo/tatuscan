"""Report routes - Inventory table and reporting."""
import logging

from flask import render_template, request

from tatuscan.services import InventoryService
from tatuscan.utils import get_timezone, to_timezone
from . import bp

logger = logging.getLogger(__name__)


@bp.route("/")
def report_index():
    """Display inventory report table with sorting."""
    try:
        sort = request.args.get("sort", "hostname").lower()
        direction = request.args.get("dir", "asc").lower()

        inventories = InventoryService.list_all(order_by=sort, direction=direction)
        tz = get_timezone()

        # Convert timestamps to local timezone
        for inv in inventories:
            inv.created_at = to_timezone(inv.created_at, tz)
            inv.updated_at = to_timezone(inv.updated_at, tz)
            inv.computer_activation = to_timezone(inv.computer_activation, tz)

        return render_template(
            "report/index.html",
            inventories=inventories,
            sort=sort,
            direction=direction,
        )
    except Exception as e:
        logger.error(f"Error rendering report: {str(e)}")
        return render_template("report/index.html", inventories=[], sort="hostname", direction="asc")
