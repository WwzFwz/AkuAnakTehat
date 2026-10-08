import os
from pathlib import Path
import unittest
from unittest import mock

from scripts.loadtest import run


class RunnerConfigurationTests(unittest.TestCase):
    def test_scenario_overrides_are_allowlisted(self):
        with mock.patch.dict(
            os.environ,
            {
                "VUS": "2",
                "DURATION": "5s",
                "SLEEP": "0",
                "CLIENT_SECRET": "must-not-be-forwarded",
                "K6_RUNNER_TIMEOUT": "2m",
            },
            clear=True,
        ):
            self.assertEqual(
                run.scenario_overrides(),
                {"VUS": "2", "DURATION": "5s", "SLEEP": "0"},
            )

    def test_duration_argument_overrides_environment(self):
        with mock.patch.dict(os.environ, {"DURATION": "60s"}, clear=True):
            self.assertEqual(run.scenario_overrides("5s")["DURATION"], "5s")

    def test_docker_command_forwards_only_scenario_overrides(self):
        with mock.patch.object(run, "RESULTS", Path("C:/repo/docs/evidence/load")):
            command = run.docker_command(
                "outage",
                {"DURATION": "5s", "OUTAGE_SECONDS": "2", "VUS": "1"},
            )

        self.assertIn("DURATION=5s", command)
        self.assertIn("OUTAGE_SECONDS=2", command)
        self.assertIn("VUS=1", command)
        self.assertNotIn("CLIENT_SECRET=must-not-be-forwarded", command)
        self.assertEqual(command[-1], "/scripts/outage.js")

    def test_native_environment_applies_overrides_after_demo_env(self):
        with mock.patch.dict(os.environ, {}, clear=True), mock.patch.object(
            run.Path, "read_text", return_value="DURATION=90s\nVUS=10\n"
        ):
            environment = run.native_environment({"DURATION": "5s", "VUS": "2"})

        self.assertEqual(environment["DURATION"], "5s")
        self.assertEqual(environment["VUS"], "2")
        self.assertEqual(environment["API_URL"], "http://127.0.0.1:8080")


class RunnerTimeoutTests(unittest.TestCase):
    def test_parse_k6_duration(self):
        self.assertEqual(run.parse_duration("1m30s"), 90)
        self.assertEqual(run.parse_duration("500ms"), 0.5)

    def test_default_timeout_preserves_existing_minimum(self):
        with mock.patch.dict(os.environ, {}, clear=True):
            self.assertEqual(run.runner_timeout("sustained"), 240)

    def test_timeout_grows_with_duration_override(self):
        with mock.patch.dict(os.environ, {"DURATION": "5m"}, clear=True):
            self.assertEqual(run.runner_timeout("sustained"), 360)

    def test_timeout_grows_for_long_outage(self):
        with mock.patch.dict(
            os.environ,
            {"OUTAGE_SECONDS": "600", "RECOVERY_SECONDS": "30"},
            clear=True,
        ):
            self.assertEqual(run.runner_timeout("outage"), 692)

    def test_explicit_timeout_override(self):
        with mock.patch.dict(os.environ, {"K6_RUNNER_TIMEOUT": "15m"}, clear=True):
            self.assertEqual(run.runner_timeout("outage"), 900)

    def test_invalid_numeric_outage_override_is_rejected(self):
        with mock.patch.dict(os.environ, {"OUTAGE_SECONDS": "not-a-number"}, clear=True):
            with self.assertRaisesRegex(RuntimeError, "OUTAGE_SECONDS"):
                run.runner_timeout("outage")


if __name__ == "__main__":
    unittest.main()
