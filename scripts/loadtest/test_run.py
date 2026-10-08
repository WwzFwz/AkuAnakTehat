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
    def test_short_explicit_timeout_is_rejected(self):
        for timeout in ("1s", "5m", "359s"):
            with self.subTest(timeout=timeout), mock.patch.dict(
                os.environ, {"DURATION": "5m", "K6_RUNNER_TIMEOUT": timeout}, clear=True
            ):
                with self.assertRaisesRegex(RuntimeError, "at least 360s"):
                    run.runner_timeout("sustained")

    def test_explicit_timeout_accepts_duration_plus_margin(self):
        with mock.patch.dict(os.environ, {"DURATION": "5m", "K6_RUNNER_TIMEOUT": "6m"}, clear=True):
            self.assertEqual(run.runner_timeout("sustained"), 360)

    def test_duration_argument_controls_deadline(self):
        with mock.patch.dict(os.environ, {"DURATION": "5s"}, clear=True):
            self.assertEqual(run.runner_timeout("sustained", duration="10m"), 660)

    def test_outage_deadline_includes_recovery_offset_and_margin(self):
        with mock.patch.dict(os.environ, {"K6_RUNNER_TIMEOUT": "111s"}, clear=True):
            with self.assertRaisesRegex(RuntimeError, "at least 112s"):
                run.runner_timeout("outage")

    def test_explicit_timeout_does_not_bypass_invalid_outage_duration(self):
        with mock.patch.dict(
            os.environ, {"K6_RUNNER_TIMEOUT": "15m", "OUTAGE_SECONDS": "bad"}, clear=True
        ):
            with self.assertRaisesRegex(RuntimeError, "OUTAGE_SECONDS"):
                run.runner_timeout("outage")

    def test_non_finite_duration_is_rejected(self):
        with self.assertRaisesRegex(RuntimeError, "invalid duration"):
            run.parse_duration("9" * 400 + "h")

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


class RunnerExecutionTests(unittest.TestCase):
    def test_long_duration_argument_does_not_stop_at_old_deadline(self):
        # Simulate a process still running at 300s, then completing successfully.
        with mock.patch.dict(os.environ, {"DURATION": "5s"}, clear=True), \
                mock.patch.object(run, "NATIVE_K6", None), \
                mock.patch.object(run.subprocess, "Popen") as popen, \
                mock.patch.object(run.time, "monotonic", side_effect=[0, 300]), \
                mock.patch.object(run.time, "sleep"), \
                mock.patch.object(Path, "open", mock.mock_open()), \
                mock.patch.object(Path, "read_text", return_value='{"metrics": {}}'), \
                mock.patch("builtins.print"):
            popen.return_value.poll.side_effect = [None, 0, 0]
            popen.return_value.returncode = 0
            run.run("sustained", duration="10m")
            self.assertIn("DURATION=10m", popen.call_args.args[0])
            popen.return_value.terminate.assert_not_called()

    def test_invalid_deadline_fails_before_starting_process(self):
        with mock.patch.dict(os.environ, {"K6_RUNNER_TIMEOUT": "1s"}, clear=True), \
                mock.patch.object(run.subprocess, "Popen") as popen, \
                mock.patch.object(Path, "open") as open_file:
            with self.assertRaisesRegex(RuntimeError, "K6_RUNNER_TIMEOUT"):
                run.run("sustained", duration="10m")
            popen.assert_not_called()
            open_file.assert_not_called()

    def test_main_validates_all_scenarios_before_side_effects(self):
        # 120s fits seismic but cannot cover sustained's 90s plus 60s margin.
        with mock.patch.dict(os.environ, {"K6_RUNNER_TIMEOUT": "120s"}, clear=True), \
                mock.patch.object(run.sys, "argv", ["run.py"]), \
                mock.patch.object(run, "docker") as docker, \
                mock.patch.object(run.subprocess, "run") as process, \
                mock.patch.object(Path, "mkdir") as mkdir:
            with self.assertRaisesRegex(RuntimeError, "sustained.*at least 150s"):
                run.main()
            docker.assert_not_called()
            process.assert_not_called()
            mkdir.assert_not_called()


if __name__ == "__main__":
    unittest.main()
