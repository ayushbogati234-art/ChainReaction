"use client";
import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { DependencyGraph } from "../../components/graph/DependencyGraph";
import graphStyles from "../../components/graph/GraphWorkspace.module.css";
import { request } from "../../lib/api/client";
import type { Scan, GraphData, PackageDetail, Paths } from "../../types/scan";
import { InputEvidence } from "./InputEvidence";

export default function ScanWorkspace() {
 const [url,setURL]=useState("");
 const [scan,setScan]=useState<Scan>();
 const [graph,setGraph]=useState<GraphData>();
 const [selected,setSelected]=useState<PackageDetail>();
 const [paths,setPaths]=useState<Paths>();
 const [selectedID,setSelectedID]=useState<string>();
 const [activePath,setActivePath]=useState<string[]>();
 const [error,setError]=useState("");
 const [submitting,setSubmitting]=useState(false);
 const selection=useRef(0);
 const running=!!scan && !["COMPLETED","FAILED"].includes(scan.status);
 async function start(event:FormEvent) {
  event.preventDefault(); selection.current++; setSubmitting(true); setError(""); setScan(undefined); setGraph(undefined); setSelected(undefined); setPaths(undefined); setSelectedID(undefined); setActivePath(undefined);
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
  const token=++selection.current;
  setSelectedID(id); setActivePath(undefined); setSelected(undefined); setPaths(undefined); setError("");
  try {
   const route="/api/v1/scans/"+scan.id+"/packages/"+encodeURIComponent(id);
   const [detail,found]=await Promise.all([request<PackageDetail>(route),request<Paths>(route+"/paths?limit=20")]);
   if (token===selection.current) { setSelected(detail); setPaths(found); }
  } catch (e) { if (token===selection.current) { setSelectedID(undefined); setError(e instanceof Error ? e.message : "Package request failed."); } }
 },[scan?.id]);
 return <main>
  <header><strong>ChainReaction</strong><span>Python dependency analysis</span><span className="phase">Phase 1</span></header>
  <section className="input">
   <form onSubmit={start}><label htmlFor="repo">Public GitHub repository<input id="repo" required value={url} onChange={e=>setURL(e.target.value)} placeholder="https://github.com/owner/repository"/></label><button disabled={submitting || running}>Start scan</button></form>
   {error && <p role="alert" className="failure">{error}</p>}
   {scan && <div role="status" className={scan.status==="FAILED"?"failure":"progress"}><b>{scan.stage}</b><span>{scan.progress}%</span>{scan.error_message && <p>{scan.error_message}</p>}</div>}
   <InputEvidence scan={scan} graph={graph}/>
  </section>
  <section className={"workspace " + graphStyles.workspace}>
   <aside><h2>Scan evidence</h2><p>Root Poetry project only.</p><p>Edges point from dependent to dependency.</p>
    {scan?.incomplete && <><h3>Incomplete result</h3>{(scan.limitations || []).map((x,i)=><p key={i}>{x}</p>)}</>}
    {graph && <><h3>Packages</h3><p>{graph.nodes.length} nodes · {graph.edges.length} dependency edges</p>{graph.nodes.filter(n=>n.version).map(n=><button className="package-choice" key={n.id} onClick={()=>select(n.id)}>{n.name}@{n.version}</button>)}</>}
   </aside>
   <DependencyGraph graph={graph} selectedID={selectedID} activePath={activePath} onSelect={select}/>
   <aside className={"inspector " + graphStyles.inspector}><h2>Package inspector</h2>
    {selected ? <><h3>{selected.package.name}@{selected.package.version}</h3><p>Confidence: {selected.package.confidence} (metadata evidence)</p><p>Package-version evidence: {selected.package.source}</p>
     {graph?.edges.filter(edge=>edge.direct && edge.target===selected.package.id).map((edge,index)=><p key={index}>Direct declaration: pyproject.toml · group {edge.group || "main"} · constraint <code>{edge.constraint || "unspecified"}</code>. Locked version: poetry.lock.</p>)}
     {!graph?.edges.some(edge=>edge.direct && edge.target===selected.package.id) && <p>No supported direct application relationship was established. This package is present in the lockfile; inspect paths and scan limitations for context.</p>}
     {selected.package.marker && <p>Condition: {selected.package.marker}</p>}<p>{selected.dependencies.length} dependencies · {selected.dependents.length} dependents</p>
     <h3>Representative paths</h3>{paths?.paths.map((path,i)=><button type="button" key={i} className={"path " + graphStyles.path} onClick={()=>setActivePath(activePath===path?undefined:path)} aria-pressed={activePath===path}>{path.join(" -> ")}</button>)}{paths?.paths.length===0 && <p>No application path was established.</p>}{paths?.truncated && <p>Path results capped at {paths.limit}; more paths exist.</p>}
    </> : <p>Select a package node to see its evidence and paths.</p>}
   </aside>
  </section>
 </main>;
}
