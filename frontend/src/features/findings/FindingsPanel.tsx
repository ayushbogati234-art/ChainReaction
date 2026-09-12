"use client";
import {useEffect,useMemo,useState} from "react";
import {request} from "../../lib/api/client";
import {FindingDetail} from "../../components/findings/FindingDetail";
import type {Enrichment,Finding} from "../../types/finding";
import type {Impact} from "../../types/scan";
import styles from "./FindingsPanel.module.css";

export function FindingsPanel({scanID,packageID,onSelect,impact}:{scanID:string;packageID?:string;onSelect:(id:string)=>void;impact?:Impact}) {
 const [data,setData]=useState<Enrichment>();const [error,setError]=useState("");
 const [chosen,setChosen]=useState<string>();const [detail,setDetail]=useState<Finding>();
 useEffect(()=>{const abort=new AbortController();setData(undefined);setDetail(undefined);setChosen(undefined);setError("");request<Enrichment>("/api/v1/scans/"+scanID+"/findings",{signal:abort.signal}).then(setData).catch(e=>{if(!abort.signal.aborted)setError(e.message)});return()=>abort.abort();},[scanID]);
 const chosenFinding=useMemo(()=>data?.findings.find(f=>f.id===chosen),[data,chosen]);
 // A package selection owns the inspector. Keep its finding only when both
 // selections refer to the same package; this also cancels a stale detail read.
 useEffect(()=>{
  if(chosenFinding && chosenFinding.package_id===packageID)return;
  setChosen(undefined);setDetail(undefined);
 },[packageID,chosenFinding]);
 useEffect(()=>{
  setDetail(undefined);
  if(!chosen || chosenFinding?.package_id!==packageID)return;
  const abort=new AbortController();
  request<Finding>("/api/v1/scans/"+scanID+"/findings/"+chosen,{signal:abort.signal}).then(result=>{
   if(!abort.signal.aborted && result.id===chosen && result.package_id===packageID)setDetail(result);
  }).catch(e=>{if(!abort.signal.aborted)setError(e.message)});
  return()=>abort.abort();
 },[scanID,packageID,chosen,chosenFinding]);
 const coverage=data?.packages.find(p=>p.package_id===packageID);
 const checked=data?.packages.filter(p=>p.status==="CHECKED").length || 0;
 return <section className={styles.panel} aria-label="Vulnerability findings">
  <h2>Vulnerability findings</h2>
  {error && <p role="alert">Enrichment results unavailable: {error}. Coverage is unknown.</p>}
  {!data && !error && <p>Loading enrichment evidence…</p>}
  {data && <>
   <p>Vulnerability enrichment: <strong>{data.status}</strong> · {data.provider}</p>
   <p>{checked} / {data.packages.length} packages fully checked. Vulnerability graph: {data.graph_status}.</p>
   {data.error && <p>{data.error}</p>}
   {data.status!=="COMPLETE" && <p>Coverage is incomplete or unavailable. An empty findings list does not establish that packages are free of known vulnerabilities.</p>}
   {coverage && <p>Selected package coverage: <strong>{coverage.status}</strong>{coverage.reason?" — "+coverage.reason:""}{coverage.checked_at?" · checked "+coverage.checked_at:""}</p>}
   {coverage?.status==="CHECKED" && !data.findings.some(f=>f.package_id===packageID) && <p>No matching advisories returned for this checked package at scan time. This is not a safety guarantee.</p>}
   <details><summary>Package coverage and provider limitations</summary>{data.packages.map(p=><p key={p.package_id}>{p.package_id}: {p.status}{p.reason?" — "+p.reason:""}</p>)}</details>
   <p>{data.findings.length} evidence-backed findings. Potential dependency exposure, not confirmed exploitation.</p>
   <div className={styles.list}>{data.findings.map(f=><button type="button" key={f.id} aria-pressed={chosen===f.id} onClick={()=>{setError("");setDetail(undefined);setChosen(f.id);onSelect(f.package_id)}}>{f.name}@{f.version} · {f.advisory_id}</button>)}</div>
   {chosen && !detail && !error && <p>Loading finding details…</p>}
   {detail && <FindingDetail finding={detail} impact={impact?.package_id===detail.package_id?impact:undefined} onPaths={()=>onSelect(detail.package_id)}/>}
  </>}
 </section>;
}
