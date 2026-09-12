"use client";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { DependencyGraph } from "../../components/graph/DependencyGraph";
import graphStyles from "../../components/graph/GraphWorkspace.module.css";
import { request } from "../../lib/api/client";
import type { Scan, GraphData, PackageDetail, Impact } from "../../types/scan";
import { InputEvidence } from "./InputEvidence";
import { FindingsPanel } from "../findings/FindingsPanel";
import { ImpactEvidence } from "../../components/graph/ImpactEvidence";

export default function ScanWorkspace() {
 const [url,setURL]=useState("");
 const [scan,setScan]=useState<Scan>();
 const [graph,setGraph]=useState<GraphData>();
 const [selected,setSelected]=useState<PackageDetail>();
 const [impact,setImpact]=useState<Impact>();
 const [applicationID,setApplicationID]=useState("");
 const [selectedID,setSelectedID]=useState<string>();
 const [activePath,setActivePath]=useState<string[]>();
 const [error,setError]=useState("");
 const [submitting,setSubmitting]=useState(false);
 const selection=useRef(0);
 const running=!!scan && !["COMPLETED","FAILED"].includes(scan.status);
 async function start(event:FormEvent) {
  event.preventDefault(); selection.current++; setSubmitting(true); setError(""); setScan(undefined); setGraph(undefined); setSelected(undefined); setImpact(undefined); setApplicationID(""); setSelectedID(undefined); setActivePath(undefined);
  try {
   const project=await request<{id:string}>("/api/v1/projects",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({repository_url:url})});
   setScan(await request<Scan>("/api/v1/projects/"+project.id+"/scans",{method:"POST"}));
  } catch (e) { setError(e instanceof Error ? e.message : "Unable to start scan."); }
  finally { setSubmitting(false); }
 }
 useEffect(() => {
  if (!scan || !running) return;
  const abort=new AbortController();
  const timer=setTimeout(async () => {
   try { setScan(await request<Scan>("/api/v1/scans/"+scan.id+"/status",{signal:abort.signal})); }
   catch (e) { if (!abort.signal.aborted) setError(e instanceof Error ? e.message : "Unable to read scan status."); }
  },1000);
  return () => { clearTimeout(timer); abort.abort(); };
 },[scan,running]);
 useEffect(() => {
  if (scan?.status!=="COMPLETED") return;
  const abort=new AbortController();
  request<GraphData>("/api/v1/scans/"+scan.id+"/graph",{signal:abort.signal}).then(setGraph).catch(e => { if (!abort.signal.aborted) setError(e.message); });
  return () => abort.abort();
 },[scan?.id,scan?.status]);
 const select=useCallback(async (id:string) => {
  if (!scan) return;
  if (selected?.package.id===id) { setActivePath(undefined); return; }
  const token=++selection.current;
  setSelectedID(id); setApplicationID(""); setActivePath(undefined); setSelected(undefined); setImpact(undefined); setError("");
  try {
   const route="/api/v1/scans/"+scan.id+"/packages/"+encodeURIComponent(id);
   const detail=await request<PackageDetail>(route);
   if (token===selection.current) setSelected(detail);
  } catch (e) { if (token===selection.current) { setSelectedID(undefined); setError(e instanceof Error ? e.message : "Package request failed."); } }
 },[scan?.id,selected?.package.id]);
 useEffect(()=>{
  if(!scan || !selectedID)return;
  const abort=new AbortController();setImpact(undefined);setActivePath(undefined);
  const query=new URLSearchParams({max_paths:"20",max_depth:"100"});
  if(applicationID)query.set("application_id",applicationID);
  request<Impact>("/api/v1/scans/"+scan.id+"/packages/"+encodeURIComponent(selectedID)+"/impact?"+query,{signal:abort.signal}).then(setImpact).catch(e=>{if(!abort.signal.aborted)setError(e instanceof Error?e.message:"Impact analysis failed.")});
  return()=>abort.abort();
 },[scan?.id,selectedID,applicationID]);
 return <main>
  <header><strong>ChainReaction</strong><span>Python dependency analysis</span><span className="phase">Phase 3 impact evidence</span></header>
  <section className="input">
   <form onSubmit={start}><label htmlFor="repo">Public GitHub repository<input id="repo" required value={url} onChange={e=>setURL(e.target.value)} placeholder="https://github.com/owner/repository"/></label><button disabled={submitting || running}>Start scan</button></form>
   {error && <p role="alert" className="failure">{error}</p>}
   {scan && <div role="status" className={scan.status==="FAILED"?"failure":"progress"}><b>{scan.stage}</b><span>{scan.progress}%</span>{scan.error_message && <p>{scan.error_message}</p>}</div>}
   <InputEvidence scan={scan} graph={graph}/>
   {scan?.status==="COMPLETED" && <p>Dependency metadata: {scan.incomplete?"incomplete — see limitations":"processed"} · Vulnerability enrichment: {scan.enrichment_status || "NOT_RUN"}</p>}
  </section>
  <section className={"workspace " + graphStyles.workspace}>
   <aside><h2>Scan evidence</h2><p>Root Poetry project only.</p><p>Edges point from dependent to dependency.</p>
    {scan?.incomplete && <><h3>Incomplete result</h3>{(scan.limitations || []).map((x,i)=><p key={i}>{x}</p>)}</>}
    {graph && <><h3>Packages</h3><p>{graph.nodes.length} nodes · {graph.edges.length} dependency edges</p>{graph.nodes.filter(n=>n.version).map(n=><button className="package-choice" key={n.id} onClick={()=>select(n.id)}>{n.name}@{n.version}</button>)}</>}
   </aside>
   <DependencyGraph graph={graph} selectedID={selectedID} activePath={activePath} affectedIDs={impact?.affected_node_ids} immediateDependentIDs={impact?.immediate_dependent_packages} onSelect={select}/>
   <aside className={"inspector " + graphStyles.inspector}><h2>Package inspector</h2>
    {scan?.status==="COMPLETED" && <FindingsPanel key={scan.id} scanID={scan.id} packageID={selectedID} impact={impact} onSelect={select}/>}
    {selected ? <><h3>{selected.package.name}@{selected.package.version}</h3><p>Confidence: {selected.package.confidence} (metadata evidence)</p><p>Package-version evidence: {selected.package.source}</p>
     {graph?.edges.filter(edge=>edge.direct && edge.target===selected.package.id).map((edge,index)=><p key={index}>Direct declaration: pyproject.toml · group {edge.group || "main"} · constraint <code>{edge.constraint || "unspecified"}</code>. Locked version: poetry.lock.</p>)}
     {!graph?.edges.some(edge=>edge.direct && edge.target===selected.package.id) && <p>No supported direct application relationship was established. This package is present in the lockfile; inspect paths and scan limitations for context.</p>}
     {selected.package.marker && <p>Condition: {selected.package.marker}</p>}<p>{selected.dependencies.length} dependencies · {selected.dependents.length} dependents</p>
     {impact ? <ImpactEvidence impact={impact} applicationID={applicationID} onApplication={setApplicationID} activePath={activePath} onPath={path=>setActivePath(activePath===path?undefined:path)}/> : <p>Loading graph impact evidence…</p>}
    </> : <p>Select a package node to see its evidence and paths.</p>}
   </aside>
  </section>
 </main>;
}
