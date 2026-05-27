import unittest


class Infra001IntentionalFailure(unittest.TestCase):
    def test_intentional_failure_for_readiness_tooling(self) -> None:
        self.assertEqual("expected", "intentionally-wrong")
