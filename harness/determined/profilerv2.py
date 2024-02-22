import contextlib
import logging
from datetime import datetime, timezone
import enum
import threading
import time
from typing import Any, Dict, Iterator, List
import queue

import psutil

from determined.common import api
from determined.common.api import bindings


class Profiler:
    def __init__(self):
        pass

    def start(self):
        pass

    def stop(self):
        pass

    def record_metric(self, metric_name: str, value: float) -> None:
        pass

    @contextlib.contextmanager
    def record_timing(
        self, metric_name: str, accumulate: bool = False, requires_sync: bool = True
    ) -> Iterator[None]:
        pass


GIGA = 1_000_000_000


class MetricCollector:
    def __init__(self):
        pass

    def collect(self) -> Dict[str, Any]:
        pass


class NetworkCollector(MetricCollector):
    def __init__(self):
        # Set initial values for throughput calculations.
        self._collect_ts = time.time()
        self._net_io = psutil.net_io_counters()
        super().__init__()

    def collect(self) -> Dict[str, Any]:
        now = time.time()
        net_io = psutil.net_io_counters()

        time_d = now - self._collect_ts

        prev_bytes_sent = self._net_io.bytes_sent
        bytes_sent = net_io.bytes_sent
        sent_thru = (bytes_sent - prev_bytes_sent) / time_d

        prev_bytes_recv = self._net_io.bytes_recv
        bytes_recv = net_io.bytes_recv
        recv_thru = (bytes_recv - prev_bytes_recv) / time_d

        self._collect_ts = now
        self._net_io = net_io

        return {
            "net_throughput_sent": sent_thru,
            "net_throughput_recv": recv_thru,
        }


class DiskCollector(MetricCollector):
    def __init__(self):
        # Set initial values for throughput calculations.
        self._collect_ts = time.time()
        self._disk_io = psutil.disk_io_counters()
        super().__init__()

    def collect(self) -> Dict[str, Any]:
        new_disk_io = psutil.disk_io_counters()
        now = time.time()
        time_d = now - self._collect_ts

        prev_iop = self._disk_io.read_count + self._disk_io.write_count
        new_iop = new_disk_io.read_count + new_disk_io.write_count
        iops = (new_iop - prev_iop) / time_d

        # xxx: can use disk_io.read/write_time but not on NetBSD and OpenBSD
        prev_read_bytes = self._disk_io.read_bytes
        new_read_bytes = new_disk_io.read_bytes
        read_thru = (new_read_bytes - prev_read_bytes) / time_d

        prev_write_bytes = self._disk_io.write_bytes
        new_write_bytes = new_disk_io.write_bytes
        write_thru = (new_write_bytes - prev_write_bytes) / time_d

        self._disk_io = new_disk_io
        self._collect_ts = now

        # xxx: would be nice to have disk usage

        return {
            "disk_iops": iops,
            "disk_throughput_read": read_thru,
            "disk_throughput_write": write_thru,
        }


class MemoryCollector(MetricCollector):
    def __init__(self):
        super().__init__()

    def collect(self) -> Dict[str, Any]:
        free_mem_bytes = psutil.virtual_memory().available
        return {
            "free_memory": free_mem_bytes / GIGA,
        }


class CPUCollector(MetricCollector):
    def __init__(self):
        super().__init__()

    def collect(self) -> Dict[str, Any]:
        cpu_util = psutil.cpu_percent()
        return {
            "cpu_util_simple": cpu_util,
        }


class GPUCollector(MetricCollector):
    try:
        import pynvml
    except ImportError:
        logging.warning(f"pynvml module not found. GPU metrics will not be collected.")
        pynvml = None

    def __init__(self):
        # xxx: map[uuid] -> handle
        self._pynvml_device_handles: Dict[str, Any] = {}
        self._init_pynvml()
        super().__init__()

    def _init_pynvml(self) -> None:
        if not self.pynvml:
            return
        try:
            self.pynvml.nvmlInit()
            # xxx: need to catch exception here?
            num_gpus = self.pynvml.nvmlDeviceGetCount()
            for i in range(num_gpus):
                handle = self.pynvml.nvmlDeviceGetHandleByIndex(i)
                uuid = self.pynvml.nvmlDeviceGetUUID(handle)
                self.pynvml.nvmlDeviceGetMemoryInfo(handle)
                self.pynvml.nvmlDeviceGetUtilizationRates(handle)
                self._pynvml_device_handles[uuid] = handle

        except self.pynvml.NVMLError as ne:
            logging.error(f"Error initializing pynvml: {ne}")

    def collect(self) -> Dict[str, Any]:
        """
        "GPU-UUID-1": {
          "gpu_util": 0.3,
          "gpu_free_memory": 1234,
        },
        "GPU-UUID-2": {
          "gpu_util": 0.3,
          "gpu_free_memory": 1234,
        }
        """
        metrics = {}

        if not self.pynvml:
            return metrics

        for uuid, handle in self._pynvml_device_handles.items():
            metrics[uuid] = {}
            gpu_util = self.pynvml.nvmlDeviceGetUtilizationRates(handle).gpu
            free_memory = self.pynvml.nvmlDeviceGetMemoryInfo(handle).free
            metrics[uuid]["gpu_util"] = gpu_util
            metrics[uuid]["gpu_free_memory"] = free_memory
        return metrics


# class SystemMetric(enum.Enum):
#     GPU_UTIL = "gpu_util"
#     GPU_FREE_MEMORY = "gpu_free_memory"
#     NET_THRU_SENT = "net_throughput_sent"
#     NET_THRU_RECV = "net_throughput_recv"
#     DISK_IOPS = "disk_iops"
#     DISK_THRU_READ = "disk_throughput_read"
#     DISK_THRU_WRITE = "disk_throughput_write"
#     FREE_MEM = "free_memory"
#     CPU_UTIL = "cpu_util_simple"


class SystemMetricGroup(enum.Enum):
    GPU = "gpu"
    NETWORK = "network"
    DISK = "disk"
    MEMORY = "memory"
    CPU = "cpu"


class SystemMetric:
    def __init__(
        self,
        group: str,
        metrics: Dict[str, Any],
        timestamp: datetime,
    ):
        self.group = group
        self.metrics = metrics
        self.timestamp = timestamp


class SystemMetricsCollector(threading.Thread):
    def __init__(
        self,
        metrics_queue: queue.Queue,
        collection_interval: int = 1,
    ):
        self._collection_interval = collection_interval
        self._should_exit = False
        self._last_flush_ts = None
        self._metrics_queue = metrics_queue

        self._metric_collectors = {
            SystemMetricGroup.GPU: GPUCollector(),
            SystemMetricGroup.CPU: CPUCollector(),
            SystemMetricGroup.MEMORY: MemoryCollector(),
            SystemMetricGroup.DISK: DiskCollector(),
            SystemMetricGroup.NETWORK: NetworkCollector(),
        }

        # xxx: does this need to be a daemon?
        super().__init__(daemon=True)

    def run(self) -> None:
        while not self._should_exit:
            time.sleep(self._collection_interval)

    def _collect_system_metrics(self) -> None:
        for group, collector in self._metric_collectors.items():
            timestamp = datetime.now(timezone.utc)
            metrics = collector.collect()
            if not metrics:
                continue
            group_metrics = SystemMetric(group=group.value, metrics=metrics, timestamp=timestamp)
            self._metrics_queue.put(group_metrics)


class MetricsShipper(threading.Thread):
    def __init__(
        self, session: api.Session, trial_id: int, run_id: int, metrics_queue: queue.Queue
    ):
        self._metrics_queue = metrics_queue
        self._should_exit = False
        self._flush_interval = None
        self._session = session
        self._trial_id = trial_id
        self._run_id = run_id

        # xxx: does this need to be a daemon?
        super().__init__(daemon=True)

    def run(self) -> None:
        while not self._should_exit:
            time.sleep(self._flush_interval)

    def ship(self, group: str, metrics: Dict[str, Any]) -> None:
        v1metrics = bindings.v1Metrics(avgMetrics=metrics)
        v1TrialMetrics = bindings.v1TrialMetrics(
            metrics=v1metrics,
            trialId=self._trial_id,
            trialRunId=self._run_id,
        )
        body = bindings.v1ReportTrialMetricsRequest(metrics=v1TrialMetrics, group=group)
        bindings.post_ReportTrialMetrics(self._session, body=body, metrics_trialId=self._trial_id)


if __name__ == "__main__":
    mq = queue.Queue()
    smc = SystemMetricsCollector(mq, 1)
    smc._collect_system_metrics()
    # time.sleep(1)
    # smc._collect_system_metrics()

    while not mq.empty():
        m = mq.get()
        print(f"group: {m.group}")
        print(f"timestamp: {m.timestamp}")
        print(f"metrics: {m.metrics}")
