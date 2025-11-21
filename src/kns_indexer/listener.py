import logging
import re
import time
from datetime import datetime
from enum import IntEnum, unique
from typing import Any, Final, final

from requests import Session
from sqlalchemy import Engine, select, update
from sqlalchemy.dialects.postgresql import insert

from kns_indexer.models import SettingsModel, UsernameModel

LOGGER: Final = logging.getLogger(__name__)

KEETA_BASE_URL: Final = "https://rep1.test.network.api.keeta.com"
KEETOOLS_BASE_URL: Final = "https://api.test.keetools.org"

TRANSACTIONS_PAGE_LIMIT: Final = 100
LAUNCH_DATE: Final = "2025-11-20"

TOKEN_NAME: Final = "KNS"
USERNAME_PATTERN: Final = re.compile(r"^[A-Za-z0-9_]{1,32}$")

FAUCET_ADDRESS: Final = (  # since Keeta Network does not have an official burn address, we will use testnet faucet address as the burn address
    "keeta_aabszsbrqppriqddrkptq5awubshpq3cgsoi4rc624xm6phdt74vo5w7wipwtmiw"
)

SET_CID_PATTERN: Final = re.compile(r"^set_cid (\w+) (\w+)$")


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
                            username = db_session.scalar(
                                insert(UsernameModel)
                                .values(
                                    username=operation["description"].lower(),
                                    address=block["account"],
                                    owner=block["signer"],
                                    timestamp=block_timestamp,
                                )
                                .on_conflict_do_nothing()
                                .returning(UsernameModel),
                            )
                            if isinstance(username, UsernameModel):
                                post_commit_logs.append(
                                    f"{block['signer']} inscribed username {operation['description'].lower()}",
                                )
                        elif _is_set_primary_name_or_cid_instruction(
                            operation,
                        ):
                            if match := SET_CID_PATTERN.match(
                                operation["extra"],
                            ):
                                token_address, cid = match.groups()
                                username = db_session.scalar(
                                    update(UsernameModel)
                                    .values(cid=cid)
                                    .where(
                                        UsernameModel.address == token_address,
                                        UsernameModel.owner
                                        == block["account"],
                                    )
                                    .returning(UsernameModel),
                                )
                                if isinstance(username, UsernameModel):
                                    post_commit_logs.append(
                                        f"{block['account']} set CID {cid} to {username.username}",
                                    )
                        elif _is_transfer_instruction(operation):
                            username = db_session.scalar(
                                update(UsernameModel)
                                .values(owner=operation["to"])
                                .where(
                                    UsernameModel.address
                                    == operation["token"],
                                    UsernameModel.owner == block["account"],
                                )
                                .returning(UsernameModel),
                            )
                            if isinstance(username, UsernameModel):
                                post_commit_logs.append(
                                    f"{block['account']} transferred username {username.username} to {operation['to']}",
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
        return response.json()


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
        return response.json()


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
        and USERNAME_PATTERN.match(operation["description"].lower())
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


def _is_set_primary_name_or_cid_instruction(
    operation: dict[str, Any],
    /,
) -> bool:
    return (
        operation["type"] == OperationType.SEND
        and operation["to"] == FAUCET_ADDRESS
        and operation.get("extra") is not None
    )
