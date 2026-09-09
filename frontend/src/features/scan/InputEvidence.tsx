import type { GraphData, Scan } from "../../types/scan";
import styles from "./InputEvidence.module.css";

export function InputEvidence({scan, graph}: {scan?: Scan; graph?: GraphData}) {
 const application=graph?.nodes.find(node=>node.id.startsWith("app:"));
 return <details className={styles.evidence}>
  <summary>Supported inputs and scan provenance</summary>
  <p>Phase 1 requires paired root <code>pyproject.toml</code> + <code>poetry.lock</code>, using Poetry declarations under <code>[tool.poetry]</code>. Nested projects, standalone requirements files, and other dependency formats are unsupported.</p>
  <p><code>tox.ini</code> is not parsed. Tox test-environment dependencies are outside this scan; Poetry development groups are declaration evidence, not a resolved tox environment.</p>
  {application && <>
   <h3>Repository files read</h3>
   {application.input_files?.length ? <ul>{application.input_files.map(file=><li key={file}><code>{file}</code> (repository root)</li>)}</ul> : <p>This saved scan predates recorded file provenance. Run a new scan to record the files read.</p>}
   {scan?.commit_sha && <p>Repository commit: <code>{scan.commit_sha}</code></p>}
   <p>Application identity: <strong>{application.name}</strong> from <code>{application.source || "pyproject.toml (Phase 1 input contract)"}</code>.</p>
   <p>Direct dependency declarations and constraints come from <code>pyproject.toml</code>. Exact package versions and package-to-package dependency evidence come from <code>poetry.lock</code>. A locked package may remain in the graph without a direct application declaration.</p>
  </>}
  <p>Metadata evidence does not prove installation or runtime use. Lockfile freshness (content hash) is not verified. Version ranges are preserved without validating compatibility; new scans report this as incomplete. Simple exact release mismatches and unsupported declarations omit the affected relationship with a reason.</p>
  <p>Incomplete results retain supported evidence. Missing or ambiguous lock matches, unsupported sources or declarations, and unselected Python/platform/extras/group conditions can prevent a complete dependency picture. Conditions are not claimed active.</p>
  {!!scan?.limitations?.length && <details><summary>Scan limitations ({scan.limitations.length})</summary><ul>{scan.limitations.map((message,index)=><li key={index}>{message}</li>)}</ul></details>}
 </details>;
}
