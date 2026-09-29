import type { RiskScore, SystemicDependency } from "../../types/scan";

const labels: Record<string, string> = {
  vulnerability_severity: "Vulnerability severity",
  blast_radius: "Blast Radius",
  structural_criticality: "Structural criticality",
  exploitability_evidence: "Exploitability evidence",
};

export function RiskEvidence({ risk }: { risk: RiskScore }) {
  return <section aria-label="Overall risk evidence">
    <h3>Overall package risk</h3>
    <p>{risk.score === null ? "Overall risk unavailable" : `${risk.score}/100 — prototype prioritization score`}</p>
    {risk.score === null && <p>Missing evidence stays unknown. Available components remain visible below.</p>}
    <dl>{Object.entries(risk.components).map(([name, component]) => <div key={name}>
      <dt>{labels[name] || name}</dt>
      <dd>{component.value === null ? "Unknown" : `${component.value.toFixed(1)}/100`} · weight {Math.round(component.weight * 100)}% · {component.contribution === null ? "contribution unavailable" : `${component.contribution.toFixed(2)} points`}</dd>
    </div>)}</dl>
    <details><summary>Risk calculation and evidence</summary>
      {risk.explanation.map((text, index) => <p key={index}>{text}</p>)}
      {risk.reason_codes.length > 0 && <p>Availability reasons: {risk.reason_codes.join(", ")}</p>}
      {Object.entries(risk.components).map(([name, component]) => <div key={name}>
        <h4>{labels[name] || name}</h4>
        {component.evidence.map((text, index) => <p key={index}>{text}</p>)}
        {component.reason_codes.map(reason => <p key={reason}>{reason}</p>)}
      </div>)}
    </details>
    <p>This score describes package prioritization within this scan. It is not an exploit probability or proof that your application is exploitable.</p>
  </section>;
}

export function SystemicEvidence({ finding }: { finding: SystemicDependency }) {
  return <section aria-label="Systemic Dependency evidence">
    <h3>{finding.status === "MATCH" ? "SYSTEMIC DEPENDENCY" : "Systemic Dependency assessment"}</h3>
    <p>{finding.status === "MATCH" ? "No known advisory returned at scan time." : finding.status === "UNAVAILABLE" ? "Classification unavailable — evidence is incomplete." : "Does not meet the systemic rule."}</p>
    <p>Rule: criticality ≥{finding.threshold}/100, complete scan enrichment, and no known advisory for this package.</p>
    {finding.checked_at && <p>{finding.provider} check: <time dateTime={finding.checked_at}>{finding.checked_at}</time></p>}
    <details><summary>Systemic classification evidence</summary>
      {finding.explanation.map((text, index) => <p key={index}>{text}</p>)}
      {finding.reason_codes.map(reason => <p key={reason}>{reason}</p>)}
    </details>
  </section>;
}
