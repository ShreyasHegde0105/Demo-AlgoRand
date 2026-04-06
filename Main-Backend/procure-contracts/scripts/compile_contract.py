from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CONTRACT_PATH = ROOT / "smart_contracts" / "escrow_contract" / "contract.py"
VENV_PUYAPY = ROOT / ".venv" / "Scripts" / "puyapy.exe"
ARTIFACTS_DIR = ROOT / "artifacts"
GENERATED_DIR = ROOT / "smart_contracts" / "escrow_contract" / "artifacts"


def compile_contract() -> None:
    ARTIFACTS_DIR.mkdir(parents=True, exist_ok=True)

    subprocess.run(
        [str(VENV_PUYAPY), str(CONTRACT_PATH.relative_to(ROOT)), "--out-dir", "artifacts"],
        cwd=ROOT,
        check=True,
    )

    copy_required_output("ProcureEscrowContract.arc56.json", "procure_escrow.arc56.json")
    copy_required_output("ProcureEscrowContract.approval.teal", "procure_escrow.approval.teal")
    copy_required_output("ProcureEscrowContract.clear.teal", "procure_escrow.clear.teal")


def copy_required_output(source_name: str, target_name: str) -> None:
    source_path = GENERATED_DIR / source_name
    if not source_path.exists():
        raise FileNotFoundError(f"expected compiler output at {source_path}")
    shutil.copy2(source_path, ARTIFACTS_DIR / target_name)


def main() -> int:
    try:
        compile_contract()
    except Exception as exc:
        print(f"compile failed: {exc}", file=sys.stderr)
        return 1

    print(f"wrote {ARTIFACTS_DIR / 'procure_escrow.arc56.json'}")
    print(f"wrote {ARTIFACTS_DIR / 'procure_escrow.approval.teal'}")
    print(f"wrote {ARTIFACTS_DIR / 'procure_escrow.clear.teal'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
