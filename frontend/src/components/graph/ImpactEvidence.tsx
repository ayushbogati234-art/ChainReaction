import type { Impact } from "../../types/scan";
import styles from "./GraphWorkspace.module.css";

export function ImpactEvidence({impact,applicationID,onApplication,onPath,activePath}:{impact:Impact;applicationID:string;onApplication:(id:string)=>void;onPath:(path:string[])=>void;activePath?:string[]}) {
 const paths=impact.paths;
 const score=impact.blast_radius;
 const count=paths.count_capped ? paths.total_paths+"+" : String(paths.total_paths);
 return <section aria-label="Graph impact evidence">
  <h3>Potential impact</h3>
  {score && <section aria-label="Blast Radius score">
   <h3>Blast Radius — scan-relative structural impact score</h3>
   <p>{score.score===null?"Score unavailable":score.score+"/100 — "+score.classification}</p>
   <p>Full scan: {score.total_applications} application roots · {score.total_packages} package versions.</p>
   {score.reason_codes.map(reason=><p key={reason}>{reason}: {reason==="INCOMPLETE_DEPENDENCY_EVIDENCE"?"Dependency evidence is incomplete.":reason==="SCAN_PATH_COUNT_CAPPED"?"A package count is capped, so the scan-wide normalization maximum cannot be established exactly.":"Exact normalization could not be established within analysis limits."}</p>)}
   {score.blocking_package && <p>Normalization blocked by package: {score.blocking_package}. This can prevent scoring other packages in the scan.</p>}
   {score.components && Object.entries(score.components).map(([name,c])=><p key={name}>{name.replaceAll("_"," ")}: {c.numerator.toFixed(4)} / {c.denominator.toFixed(4)} = {c.value.toFixed(4)} · weight {c.weight*100}% · contribution {c.contribution.toFixed(2)} points{name==="path_multiplicity"?" (numerator and denominator use ln(1 + path count))":""}</p>)}
   <p>HIGH / VERY HIGH does not establish vulnerability severity, exploitability, or absolute business impact. Bands are prototype product classifications. Compare scores only within this scan.</p>
  </section>}
  <p><strong>{impact.applications_exposed}</strong> application root{impact.applications_exposed===1?"":"s"} potentially exposed · <strong>{impact.downstream_packages}</strong> downstream package{impact.downstream_packages===1?"":"s"}.</p>
  <p>{impact.immediate_dependent_packages.length} immediate · {impact.transitive_dependent_packages.length} transitive dependents. Counts are {impact.reachability_exact?"exact within this isolated scan snapshot":"bounded"}.</p>
  <p>Impact evidence confidence: {impact.evidence_confidence}. This describes dependency evidence, not exploitability.</p>
  {!impact.evidence_complete && <p className="failure">Dependency evidence is incomplete; potentially exposed applications or paths may be missing.</p>}
  {impact.application_ids.length>0 && <label>Filter paths by application<select value={applicationID} onChange={event=>onApplication(event.target.value)}><option value="">All exposed applications</option>{impact.application_ids.map(id=><option key={id} value={id}>{id.replace(/^app:/,"")}</option>)}</select></label>}
  <h3>Representative paths</h3>
  <p>{count} {paths.count_capped?"or more ":""}simple path{paths.total_paths===1&&!paths.count_capped?"":"s"}; showing {paths.paths.length}. {paths.count_exact?"Aggregate count is exact.":"Aggregate count is bounded, not exact."}</p>
  <p>Maximum depth: {paths.maximum_depth} edge{paths.maximum_depth===1?"":"s"} — {paths.maximum_depth_exact?"exact":"maximum observed within bounded results"}.</p>
  {paths.paths.map((path,index)=><button type="button" key={path.join("\u0000")} className={"path "+styles.path} onClick={()=>onPath(path)} aria-pressed={activePath===path}>Length {paths.path_lengths?.[index] ?? path.length-1}: {path.join(" -> ")}</button>)}
  {paths.paths.length===0 && <p>No application path was established for this filter.</p>}
  {paths.cycles_detected && <p>Cycles were detected; only cycle-safe simple paths were counted.</p>}
  {(impact.limitations.length+(paths.limitations?.length||0)>0) && <details><summary>Impact evidence limitations</summary>{[...impact.limitations,...(paths.limitations||[])].map((text,index)=><p key={index}>{text}</p>)}</details>}
 </section>;
}
