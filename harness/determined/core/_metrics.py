import logging
from typing import Any, Dict, List, Optional, Union

from determined.common import api
from determined.common.api import bindings

logger = logging.getLogger("determined.core")


class MetricsContext:
    def __init__(
        self,
        session: api.Session,
        trial_id: int,
        run_id: int,
    ) -> None:
        self._session = session
        self._trial_id = trial_id
        self._run_id = run_id

    def report_metrics(
        self,
        group: str,
        steps_completed: int,
        metrics: Dict[str, Any],
    ) -> None:
        """
        Report metrics data to the master.

        Arguments:
            group (string): metrics group name. Can be used to partition metrics
                into different logical groups or time series.
                "training" and "validation" group names map to built-in training
                and validation time series. Note: Group cannot contain ``.`` character.
            steps_completed (int): global step number, e.g. the number of batches processed.
            metrics (Dict[str, Any]): metrics data dictionary. Must be JSON-serializable.
                When reporting metrics with the same ``group`` and ``steps_completed`` values,
                the dictionary keys must not overlap.
        """
        logger.info(
            f"report_metrics(group={group}, steps_completed={steps_completed}, metrics={metrics})"
        )
        self._report_trial_metrics(group, steps_completed, metrics)

    def _report_trial_metrics(
        self,
        group: str,
        steps_completed: int,
        metrics: Dict[str, Any],
        batch_metrics: Optional[List[Dict[str, Any]]] = None,
    ) -> None:
        """
        Report trial metrics to the master.

        You can include a list of ``batch_metrics``.  Batch metrics are not be shown in the WebUI
        but may be accessed from the master using the CLI for post-processing.
        """

        # reportable_metrics = metrics
        # if group == util._LEGACY_VALIDATION:
        #     # keep the old behavior of filtering out some metrics for validations.
        #     serializable_metrics = self._get_serializable_metrics(metrics)
        #     reportable_metrics = {k: metrics[k] for k in serializable_metrics}

        v1metrics = bindings.v1Metrics(avgMetrics=metrics, batchMetrics=batch_metrics)
        v1TrialMetrics = bindings.v1TrialMetrics(
            metrics=v1metrics,
            stepsCompleted=steps_completed,
            trialId=self._trial_id,
            trialRunId=self._run_id,
        )
        body = bindings.v1ReportTrialMetricsRequest(metrics=v1TrialMetrics, group=group)
        bindings.post_ReportTrialMetrics(self._session, body=body, metrics_trialId=self._trial_id)
