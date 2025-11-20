from datetime import datetime
from typing import final

from sqlalchemy import TIMESTAMP
from sqlalchemy.orm import Mapped, mapped_column

from kns_indexer.models.base import Base


@final
class DomainModel(Base):
    domain: Mapped[str] = mapped_column(primary_key=True)
    address: Mapped[str]
    owner: Mapped[str]
    timestamp: Mapped[datetime] = mapped_column(
        TIMESTAMP(timezone=True),
    )
