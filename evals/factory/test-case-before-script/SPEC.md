# Test cases before scripts

Fail if `/implement-api`, `/implement-ui`, `/apply-requirements`, `/fix-issue`, `/generate-python-module`, or `/validate-api-endpoints` in any twin can generate or run slice scripts/probes without `docs/test-cases.md` and a **PASS** `docs/test-cases-br-coverage.md` written by br-coverage-validator (not an implementer). Fail if `/expand-test-coverage` generates or runs gap scripts. Fail if `/run-tests` may start without coverage PASS **and** human test-package **Approved**.
