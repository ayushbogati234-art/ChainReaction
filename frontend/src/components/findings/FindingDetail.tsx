import type { Finding } from "../../types/finding";
import type { Impact } from "../../types/scan";

function safeURL(raw:string) {try {const url=new URL(raw);return url.protocol==="https:" || url.protocol==="http:" ? url.href:undefined;}catch{return undefined;}}
export function FindingDetail({finding,onPaths,impact}:{finding:Finding;onPaths:()=>void;impact?:Impact}) {
 return <section aria-label="Finding details">
  <h3>{finding.advisory_id}</h3>
  <p>Potential dependency exposure: {finding.name}@{finding.version} ({finding.ecosystem})</p>
  <p>Aliases: {finding.aliases.join(", ")}</p>
  <p>Dependency evidence confidence: {finding.dependency_confidence}. Vulnerability-match confidence: {finding.vulnerability_match_confidence}. Combined evidence: {finding.confidence}.</p>
  <p>HIGH describes evidence quality, not package safety, confirmed installation or exploitability.</p>
  {impact && <p>Graph impact evidence: {impact.applications_exposed} application root(s) potentially exposed, {impact.downstream_packages} downstream package(s), and {impact.paths.count_capped?impact.paths.total_paths+"+":impact.paths.total_paths} application path(s). See the package impact explorer for bounded-result details.</p>}
  {finding.explanation.map((text,index)=><p key={index}>{text}</p>)}
  {finding.advisories.map((advisory,index)=><article key={advisory.id+":"+index}>
   <h4>{advisory.id} — {advisory.source}</h4>
   <p>{advisory.summary || "Provider supplied no summary."}</p>
   <p>{advisory.match_evidence}</p>
   <p>Supplied severity / CVSS: {advisory.severity?.length ? advisory.severity.map(value=>value.type+": "+value.score+(value.source?" ("+value.source+")":"")).join("; ") : "Not supplied"}</p>
   <p>Applicable fixed-version evidence: {advisory.fixed_versions.length ? advisory.fixed_versions.join(", ") : "Not supplied or not established for this version"}. Provider evidence is not a verified compatible upgrade.</p>
   <details><summary>Affected-version evidence</summary><pre>{JSON.stringify(advisory.affected,null,2)}</pre></details>
   {advisory.details && <details><summary>Provider description</summary><p>{advisory.details}</p></details>}
   <ul>{advisory.references?.map((ref,index)=>{const href=safeURL(ref.url);return <li key={index}>{href?<a href={href} target="_blank" rel="noopener noreferrer">{ref.type}: {ref.url}</a>:"Unsupported reference URL omitted"}</li>;})}</ul>
  </article>)}
  <button type="button" onClick={onPaths}>Show dependency paths</button>
  <p>See the package impact explorer for Blast Radius and its availability evidence. Overall Risk Score, simulations and mitigation ranking are not available in this phase.</p>
 </section>;
}
