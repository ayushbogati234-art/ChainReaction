"use client";

import { useEffect, useState } from "react";
import { request } from "../../lib/api/client";
import type { SystemicReport } from "../../types/scan";

export function SystemicPanel({ scanID, onSelect }: { scanID: string; onSelect: (id: string) => void }) {
  const [report, setReport] = useState<SystemicReport>();
  const [error, setError] = useState("");
  useEffect(() => {
    const abort = new AbortController();
    setReport(undefined);
    setError("");
    request<SystemicReport>(`/api/v1/scans/${scanID}/systemic-dependencies`, { signal: abort.signal })
      .then(value => { if (!abort.signal.aborted) setReport(value); })
      .catch(cause => { if (!abort.signal.aborted) setError(cause instanceof Error ? cause.message : "Unable to load systemic findings."); });
    return () => abort.abort();
  }, [scanID]);
  return <section aria-label="Systemic Dependencies">
    <h3>Systemic Dependencies</h3>
    <p>High structural criticality with no known advisory at scan time.</p>
    {error ? <p role="alert">{error}</p> : !report ? <p>Checking systemic evidence…</p> : <>
      <p>{report.evaluated_packages}/{report.total_packages} packages assessed.</p>
      {report.status === "PARTIAL" && <p>Assessment incomplete: {report.reason_codes.join(", ")}. An empty list does not establish that none qualify.</p>}
      {report.status === "COMPLETE" && report.findings.length === 0 && <p>No packages meet the systemic rule in this scan.</p>}
      {report.findings.map(finding => <button className="package-choice" type="button" key={finding.package_id} onClick={() => onSelect(finding.package_id)}>{finding.package_id} — criticality {finding.criticality_score}/100</button>)}
    </>}
  </section>;
}
