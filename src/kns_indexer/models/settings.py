from datetime import datetime
from typing import final

from sqlalchemy import TIMESTAMP
from sqlalchemy.orm import Mapped, mapped_column

from kns_indexer.models.base import Base


@final
class SettingsModel(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    page: Mapped[int]
    last_block_timestamp: Mapped[datetime | None] = mapped_column(
        TIMESTAMP(timezone=True),
    )
    last_block_hash: Mapped[str | None]
