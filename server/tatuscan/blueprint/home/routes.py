"""Home routes - Dashboard and main page."""
import logging

from flask import render_template
from sqlalchemy import func

from tatuscan.services import InventoryService
from tatuscan.models.inventory import Inventory
from tatuscan.extensions import db
from . import bp

logger = logging.getLogger(__name__)


@bp.route("/")
def index():
    """Display dashboard with OS and age distribution."""
    try:
        # Query OS counts
        os_rows = db.session.query(
            Inventory.os.label("os"),
            func.count(Inventory.os).label("count")
        ).group_by(Inventory.os).order_by(func.count(Inventory.os).desc()).all()

        # Query version counts
        version_rows = db.session.query(
            Inventory.os_version.label("os_version"),
            func.count(Inventory.os_version).label("count")
        ).group_by(Inventory.os_version).order_by(func.count(Inventory.os_version).desc()).all()

        # Get age distribution from service
        age_data = InventoryService.get_age_distribution()

        # Convert service response to template format
        age_ranges = []
        for item in age_data:
            range_label = item["range"]
            if range_label.startswith(">"):
                min_val = int(range_label[1:])
                max_val = 9999
            else:
                parts = range_label.split("–")
                min_val = int(parts[0])
                max_val = int(parts[1])

            age_ranges.append({
                "min": min_val,
                "max": max_val,
                "count": item["count"]
            })

        # Calculate age statistics
        total_count = sum(r["count"] for r in age_ranges)
        if total_count > 0:
            total_weighted = sum(r["count"] * ((r["min"] + min(r["max"], 120)) / 2) for r in age_ranges)
            average = round(total_weighted / total_count, 1)

            ranges_with_data = [r for r in age_ranges if r["count"] > 0]
            min_age = min(r["min"] for r in ranges_with_data) if ranges_with_data else 0
            max_age = max(r["max"] if r["max"] != 9999 else 120 for r in ranges_with_data) if ranges_with_data else 0

            age_stats = {
                "count": total_count,
                "average": average,
                "min": round(min_age, 1),
                "max": round(max_age, 1)
            }
        else:
            age_stats = {
                "count": 0,
                "average": 0,
                "min": 0,
                "max": 0
            }

        return render_template(
            "home/index.html",
            os_list=os_rows,
            version_list=version_rows,
            age_stats=age_stats,
            age_ranges=age_ranges,
        )
    except Exception as e:
        logger.error(f"Error rendering dashboard: {str(e)}")
        return render_template(
            "home/index.html",
            os_list=[],
            version_list=[],
            age_stats={"count": 0, "average": 0, "min": 0, "max": 0},
            age_ranges=[]
        )
