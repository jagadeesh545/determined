import tempfile
import threading
import time
from http.server import HTTPServer, SimpleHTTPRequestHandler
from typing import Dict, Tuple, Type

import pytest
import yaml

from determined.common.api import bindings
from tests import api_utils
from tests import config as conf
from tests import experiment as exp


class HTTPServerWithRequest(HTTPServer):
    def __init__(
        self, server_address: Tuple[str, int], RequestHandlerClass: Type[SimpleHTTPRequestHandler]
    ):
        super().__init__(server_address, RequestHandlerClass)
        self.url_to_request_body: Dict[str, str] = {}
        self.url_to_request_body_lock = threading.Lock()


class WebhookServer:
    def __init__(self, port: int):
        class WebhookRequestHandler(SimpleHTTPRequestHandler):
            def do_POST(self) -> None:
                assert isinstance(self.server, HTTPServerWithRequest)
                with self.server.url_to_request_body_lock:
                    if self.path in self.server.url_to_request_body:
                        pytest.fail(f"got two webhooks sent to path {self.path}")

                    content_length = int(self.headers.get("content-length"))
                    request_body = self.rfile.read(content_length)
                    self.server.url_to_request_body[self.path] = request_body.decode("utf-8")

                    self.send_response(200, "Success")
                    self.end_headers()
                    self.wfile.write("".encode("utf-8"))

        self.server = HTTPServerWithRequest(("", port), WebhookRequestHandler)

        self.server_thread = threading.Thread(target=self.server.serve_forever)
        self.server_thread.start()

    def close_and_return_responses(self) -> Dict[str, str]:
        self.server.shutdown()
        self.server.server_close()
        self.server_thread.join()
        return self.server.url_to_request_body


@pytest.mark.e2e_cpu
@pytest.mark.parametrize("should_match", [True, False])
def test_log_pattern_policy_dont_retry(should_match: bool) -> None:
    # TODO we have this problem where a regex will always match itself.
    # Like can we add a validation here? IDK what the best academic way to solve this is.
    # This does this since the first line logs expconf which has the pattern.
    # Maybe we can censor pattern? I'm not sure. Maybe this isn't an issue.
    regex = r"assert 0 <= self\.metrics_sigma"
    if not should_match:
        regex = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"

    config = conf.load_config(conf.fixtures_path("no_op/single-medium-train-step.yaml"))
    config["log_pattern_policies"] = [
        {
            "pattern": regex,
            "policy": {
                "type": "on_failure_dont_retry",
            },
        },
    ]
    config["hyperparameters"]["metrics_sigma"] = -1
    config["max_restarts"] = 1

    with tempfile.NamedTemporaryFile() as tf:
        with open(tf.name, "w") as f:
            yaml.dump(config, f)
        exp_id = exp.create_experiment(tf.name, conf.fixtures_path("no_op"))

    exp.wait_for_experiment_state(exp_id, bindings.experimentv1State.ERROR)

    experiment_trials = exp.experiment_trials(exp_id)
    assert len(experiment_trials) == 1
    trial_logs = "\n".join(exp.trial_logs(experiment_trials[0].trial.id))

    if should_match:
        assert experiment_trials[0].trial.restarts == 0
        assert "trial failed and matched logs to a don't retry policy" in trial_logs
    else:
        assert experiment_trials[0].trial.restarts == 1
        assert "trial failed and matched logs to a don't retry policy" not in trial_logs


# TODO(DET-9872) slurm test mark.
@pytest.mark.e2e_cpu
@pytest.mark.e2e_k8s
@pytest.mark.parametrize("should_match", [True, False])
def test_log_pattern_retry_different_node(should_match: bool) -> None:
    regex = r"assert 0 <= self\.metrics_sigma"
    if not should_match:
        regex = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"

    config = conf.load_config(conf.fixtures_path("no_op/single-medium-train-step.yaml"))
    config["log_pattern_policies"] = [
        {
            "pattern": regex,
            "policy": {
                "type": "on_failure_exclude_node",
            },
        },
    ]
    config["hyperparameters"]["metrics_sigma"] = -1
    config["max_restarts"] = 1

    agents = bindings.get_GetAgents(api_utils.determined_test_session()).agents
    assert len(agents) == 1
    assert agents[0].slots is not None
    config["resources"] = {"slots_per_trial": len(agents[0].slots)}

    with tempfile.NamedTemporaryFile() as tf:
        with open(tf.name, "w") as f:
            yaml.dump(config, f)
        exp_id = exp.create_experiment(tf.name, conf.fixtures_path("no_op"))

    exp.wait_for_experiment_state(exp_id, bindings.experimentv1State.RUNNING)

    if should_match:
        # We can run another job to completion since our original should be stuck in queued.
        second_exp_id = exp.create_experiment(
            conf.fixtures_path("no_op/single-one-short-step.yaml"), conf.fixtures_path("no_op")
        )
        exp.wait_for_experiment_state(second_exp_id, bindings.experimentv1State.COMPLETED)

        exp.wait_for_experiment_state(exp_id, bindings.experimentv1State.QUEUED)

        experiment_trials = exp.experiment_trials(exp_id)
        assert len(experiment_trials) == 1
        assert experiment_trials[0].trial.restarts == 1
        trial_logs = "\n".join(exp.trial_logs(experiment_trials[0].trial.id))
        assert "therefore will not schedule on" in trial_logs

        exp.kill_experiments([exp_id])
    else:
        exp.wait_for_experiment_state(exp_id, bindings.experimentv1State.ERROR)

        experiment_trials = exp.experiment_trials(exp_id)
        assert len(experiment_trials) == 1
        assert experiment_trials[0].trial.restarts == 1
        trial_logs = "\n".join(exp.trial_logs(experiment_trials[0].trial.id))
        assert "therefore will not schedule on" not in trial_logs


@pytest.mark.e2e_cpu
@pytest.mark.parametrize("should_match", [True, False])
def test_log_pattern_send_webhook(should_match: bool) -> None:
    port = 5006
    server = WebhookServer(port)

    # bleh. Make a new interface.
    regex = r"assert 0 <= self\.metrics_sigma"
    if not should_match:
        regex = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"

    default_path = "/test/path/here"
    slack_path = "/test/slack/path/here"
    config = conf.load_config(conf.fixtures_path("no_op/single-medium-train-step.yaml"))
    config["log_pattern_policies"] = [
        {
            "pattern": regex,
            "policy": {
                "type": "send_webhook",
                "webhook_url": f"http://127.0.0.1:{port}{default_path}",
                "webhook_type": "default",
            },
        },
        {
            "pattern": regex,
            "policy": {
                "type": "send_webhook",
                "webhook_url": f"http://127.0.0.1:{port}{slack_path}",
                "webhook_type": "slack",
            },
        },
    ]
    config["hyperparameters"]["metrics_sigma"] = -1
    config["max_restarts"] = 0

    with tempfile.NamedTemporaryFile() as tf:
        with open(tf.name, "w") as f:
            yaml.dump(config, f)
        exp_id = exp.create_experiment(tf.name, conf.fixtures_path("no_op"))

    exp.wait_for_experiment_state(exp_id, bindings.experimentv1State.ERROR)

    experiment_trials = exp.experiment_trials(exp_id)
    assert len(experiment_trials) == 1
    trial_logs = "\n".join(exp.trial_logs(experiment_trials[0].trial.id))

    time.sleep(10)  # Not ideal but give us a buffer for webhooks arriving.

    responses = server.close_and_return_responses()
    if should_match:
        assert "therefore sent webhook" in trial_logs
        assert len(responses) == 2

        # Only need a spot check we get the default / slack responses.
        # Further tested in integrations.
        assert "LOG_PATTERN_POLICY" in responses[default_path]
        assert "This log matched the regex" in responses[slack_path]
    else:
        assert len(responses) == 0
        assert "therefore sent webhook" not in trial_logs
