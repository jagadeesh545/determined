import logging
from typing import Any, Dict, List, Optional, Union

from determined import core
from determined.common import api
from determined.common.api import bindings

logger = logging.getLogger("determined.core")


class ProfilingContext:
    def __init__(
        self,
        session: api.Session,
        trial_id: int,
        run_id: int,
        distributed: core.DistributedContext,
    ) -> None:
        self._session = session
        self._trial_id = trial_id
        self._run_id = run_id
        self._distributed = distributed
        self._on = False

    def on(self, interval: int = 1) -> None:
        pass

    def off(self) -> None:
        pass
