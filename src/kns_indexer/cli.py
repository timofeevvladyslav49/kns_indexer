import logging
import os
from typing import Final

from sqlalchemy import URL, create_engine
from sqlalchemy.dialects.postgresql import insert

from kns_indexer.listener import listen_instructions
from kns_indexer.models import SettingsModel
from kns_indexer.models.base import Base

LOGGER: Final = logging.getLogger(__name__)


def main() -> None:
    engine = create_engine(
        URL.create(
            drivername="postgresql+psycopg",
            username=os.environ["POSTGRES_USER"],
            password=os.environ["POSTGRES_PASSWORD"],
            host="database",
            database=os.environ["POSTGRES_DB"],
        ),
    )

    Base.metadata.create_all(engine)

    with engine.connect() as connection:
        connection.execute(
            insert(SettingsModel)
            .values(id=1, page=1)
            .on_conflict_do_nothing(),
        )
        connection.commit()

    listen_instructions(engine)


def cli() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(name)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    LOGGER.info("Starting KNS Indexer")

    try:
        main()
    except KeyboardInterrupt:
        LOGGER.info("KNS Indexer stopped!")


if __name__ == "__main__":
    cli()
