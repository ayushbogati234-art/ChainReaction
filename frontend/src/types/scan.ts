export type Scan = { id: string; status: string; stage: string; progress: number; incomplete: boolean; limitations: string[] | null; error_message?: string | null };
export type GraphNode = { id: string; name: string; version?: string; confidence?: string; source?: string; marker?: string; groups?: string; limitations?: string[] };
export type GraphData = { nodes: GraphNode[]; edges: { source: string; target: string; marker?: string }[] };
export type PackageDetail = { package: GraphNode; dependencies: GraphNode[]; dependents: GraphNode[] };
export type Paths = { paths: string[][]; truncated: boolean; limit: number };
