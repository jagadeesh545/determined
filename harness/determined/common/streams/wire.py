# Code generated by stream-gen. DO NOT EDIT.

"""Wire formats for the determined streaming updates subsystem"""

import typing


class MsgBase:
    @classmethod
    def from_json(cls, obj: typing.Any):
        return cls(**obj)

    def __repr__(self) -> str:
        body = ", ".join(f"{k}={v}" for k, v in vars(self).items())
        return f"{type(self).__name__}({body})"


class TrialMsg(MsgBase):
    def __init__(
        self,
        id: "int",
        task_id: "str",
        experiment_id: "int",
        request_id: "typing.Optional[int]",
        seed: "int",
        hparams: "typing.Any",
        state: "str",
        start_time: "float",
        end_time: "typing.Optional[float]",
        runner_state: "str",
        restarts: "int",
        tags: "typing.Any",
        seq: "int",
    ) -> None:
        self.id = id
        self.task_id = task_id
        self.experiment_id = experiment_id
        self.request_id = request_id
        self.seed = seed
        self.hparams = hparams
        self.state = state
        self.start_time = start_time
        self.end_time = end_time
        self.runner_state = runner_state
        self.restarts = restarts
        self.tags = tags
        self.seq = seq


class TrialSubscriptionSpec(MsgBase):
    def __init__(
        self,
        trial_ids: "typing.List[int]",
        experiment_ids: "typing.List[int]",
        since: "int",
    ) -> None:
        self.trial_ids = trial_ids
        self.experiment_ids = experiment_ids
        self.since = since