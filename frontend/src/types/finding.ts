export type Advisory = {
 id: string; aliases: string[] | null; source: string; summary: string; details: string;
 severity: {type:string; score:string; source?:string}[] | null;
 references: {type:string; url:string}[] | null;
 affected: unknown; fixed_versions: string[]; match_confidence:string; match_evidence:string;
};
export type Finding = {
 id:string; scan_id:string; package_id:string; name:string; version:string; ecosystem:string; purl:string;
 direct:boolean; advisory_id:string; aliases:string[]; dependency_confidence:string;
 vulnerability_match_confidence:string; confidence:string; explanation:string[]; advisories:Advisory[];
};
export type Enrichment = {
 status:string; provider:string; graph_status:string; error?:string;
 packages:{package_id:string; purl:string; status:string; reason?:string; checked_at?:string}[];
 findings:Finding[];
};
