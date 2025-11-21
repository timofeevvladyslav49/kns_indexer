import logging
import time
from datetime import datetime
from enum import IntEnum, unique
from typing import Any, Final, final

import orjson
import tldextract
from requests import Session
from sqlalchemy import Engine, select, update
from sqlalchemy.dialects.postgresql import insert

from kns_indexer.models import DomainModel, SettingsModel

LOGGER: Final = logging.getLogger(__name__)

KEETA_BASE_URL: Final = "https://rep1.test.network.api.keeta.com"
KEETOOLS_BASE_URL: Final = "https://api.test.keetools.org"

TRANSACTIONS_PAGE_LIMIT: Final = 100
LAUNCH_DATE: Final = "2025-11-20"

TOKEN_NAME: Final = "KNS"
DOMAIN: Final = "keeta"


@final
@unique
class OperationType(IntEnum):
    SEND = 0
    SET_INFO = 2
    CREATE_IDENTIFIER = 4


def listen_instructions(engine: Engine, /) -> None:
    with engine.connect() as db_session:
        settings = db_session.scalar(
            select(SettingsModel).where(SettingsModel.id == 1),
        )
        current_page = settings.page
        last_block_timestamp = settings.last_block_timestamp
        last_block_hash = settings.last_block_hash

    with Session() as session:
        while True:
            with engine.connect() as db_session:
                LOGGER.debug(
                    "Loaded settings: page=%s last_block_timestamp=%s last_block_hash=%s",
                    current_page,
                    last_block_timestamp,
                    last_block_hash,
                )

                pagination = _get_pagination(session, page=current_page)

                LOGGER.debug("Pagination: %s", pagination)

                history = _get_history(session, pagination=pagination)

                last_block_operations = None

                post_commit_logs: list[str] = []

                for block in _get_sorted_blocks(history):
                    block_timestamp = datetime.fromisoformat(block["date"])

                    if (
                        last_block_timestamp is not None
                        and last_block_hash is not None
                        and (
                            block_timestamp < last_block_timestamp
                            or block["$hash"] == last_block_hash
                        )
                    ):  # skip already processed blocks
                        LOGGER.debug(
                            "Skipping block %s: older or equal to last processed",
                            block["$hash"],
                        )
                        continue

                    LOGGER.debug(
                        "Processing block %s at %s",
                        block["$hash"],
                        block_timestamp,
                    )

                    for operation in block["operations"]:
                        if _is_inscribe_instruction(
                            operation,
                            token_account=block["account"],
                            last_block_operations=last_block_operations,
                        ):
                            domain = db_session.scalar(
                                insert(DomainModel)
                                .values(
                                    domain=operation["description"],
                                    address=block["account"],
                                    owner=block["signer"],
                                    timestamp=block_timestamp,
                                )
                                .on_conflict_do_nothing()
                                .returning(DomainModel),
                            )
                            if isinstance(domain, DomainModel):
                                post_commit_logs.append(
                                    f"{block['signer']} inscribed domain {operation['description']}",
                                )
                        elif _is_transfer_instruction(operation):
                            domain = db_session.scalar(
                                update(DomainModel)
                                .values(owner=operation["to"])
                                .where(
                                    DomainModel.address == operation["token"],
                                    DomainModel.owner == block["account"],
                                )
                                .returning(DomainModel),
                            )
                            if isinstance(domain, DomainModel):
                                post_commit_logs.append(
                                    f"{block['account']} transferred domain {domain.domain} to {operation['to']}",
                                )

                    last_block_timestamp = block_timestamp
                    last_block_hash = block["$hash"]
                    last_block_operations = block["operations"]

                if current_page != pagination["totalPages"]:
                    current_page += 1

                db_session.execute(
                    update(SettingsModel).values(
                        page=current_page,
                        last_block_timestamp=last_block_timestamp,
                        last_block_hash=last_block_hash,
                    ),
                )
                db_session.commit()

                for post_commit_log in post_commit_logs:
                    LOGGER.debug(post_commit_log)

                LOGGER.debug(
                    "Committed settings: page=%s last_block_hash=%s",
                    current_page,
                    last_block_hash,
                )

            time.sleep(1)


def _get_pagination(session: Session, /, *, page: int) -> dict[str, str | int]:
    with session.get(
        f"{KEETOOLS_BASE_URL}/api/staples/metadata",
        params={
            "limit": TRANSACTIONS_PAGE_LIMIT,
            "page": page,
            "sortOrder": "asc",
            "dateFrom": LAUNCH_DATE,
        },
    ) as response:
        return orjson.loads(response.content)


def _get_history(
    session: Session,
    /,
    *,
    pagination: dict[str, str | int],
) -> dict[str, Any]:
    params = {"limit": TRANSACTIONS_PAGE_LIMIT}
    if start := pagination.get("startBlocksHash"):
        params["start"] = start
    with session.get(
        f"{KEETA_BASE_URL}/api/node/ledger/history",
        params=params,
    ) as response:
        return orjson.loads(response.content)


def _get_sorted_blocks(history: dict[str, Any], /) -> list[Any]:
    return sorted(
        [
            block
            for inner_history in history["history"]
            for block in inner_history["voteStaple"]["blocks"]
        ],
        key=lambda block: datetime.fromisoformat(block["date"]),
    )


def _is_inscribe_instruction(
    operation: dict[str, Any],
    /,
    *,
    token_account: str,
    last_block_operations: list[dict[str, Any]],
) -> bool:
    return (
        operation["type"] == OperationType.SET_INFO
        and operation["name"] == TOKEN_NAME
        and tldextract.extract(operation["description"]).domain == DOMAIN
        and any(
            token_account == operation["identifier"]
            for operation in last_block_operations
            if operation["type"] == OperationType.CREATE_IDENTIFIER
        )
    )


def _is_transfer_instruction(operation: dict[str, Any], /) -> bool:
    return (
        operation["type"] == OperationType.SEND
        and operation["amount"] == "0x1"
    )
