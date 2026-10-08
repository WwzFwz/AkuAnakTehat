"""Regression checks for safe presentation of authenticated demo requests."""
import contextlib
import http.server
import importlib.util
import io
from pathlib import Path
import threading
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("demo", Path(__file__).with_name("demo.py"))
demo = importlib.util.module_from_spec(spec)
spec.loader.exec_module(demo)


class PresentationSafety(unittest.TestCase):
    def test_redirect_does_not_forward_login_body(self):
        destinations = []

        class Handler(http.server.BaseHTTPRequestHandler):
            def do_POST(self):
                self.rfile.read(int(self.headers.get("Content-Length", 0)))
                destinations.append(self.path)
                self.send_response(307)
                self.send_header("Location", "/unexpected-target")
                self.end_headers()

            def do_GET(self):
                destinations.append(self.path)
                self.send_response(200)
                self.end_headers()

            def log_message(self, *args):
                pass

        with http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler) as server:
            thread = threading.Thread(target=server.serve_forever, daemon=True)
            thread.start()
            try:
                status, _, _ = demo.request(server.server_port, "/oauth/token",
                                             {"client_secret": "test-only-secret"}, form=True)
                self.assertEqual(status, 307)
                self.assertEqual(destinations, ["/oauth/token"])
            finally:
                server.shutdown()
                thread.join(timeout=2)

    def test_display_masks_known_credentials_in_nested_values(self):
        credential = "synthetic-demo-credential-for-test"
        output = io.StringIO()
        with patch.object(demo, "local_env", return_value={"KEY": credential}), \
                patch.object(demo, "clients", return_value=[]), contextlib.redirect_stdout(output):
            demo.emit({"body": {"attributes": [credential]}, "http_status": 200})
        self.assertNotIn(credential, output.getvalue())
        self.assertIn("[REDACTED]", output.getvalue())
        self.assertIn("200", output.getvalue())


if __name__ == "__main__":
    unittest.main()
