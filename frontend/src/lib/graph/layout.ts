import type { Core, CoseLayoutOptions } from "cytoscape";

export const dependencyLayout: CoseLayoutOptions = {
  name: "cose",
  animate: false,
  fit: false,
  padding: 120,
  nodeDimensionsIncludeLabels: true,
  randomize: false,
  nodeRepulsion: () => 160000,
  nodeOverlap: 20,
  idealEdgeLength: () => 190,
  // CoSE uses elasticity as a force divisor, not a spring multiplier.
  edgeElasticity: () => 100,
  nestingFactor: 0.8,
  gravity: 0.22,
  numIter: 1800,
  initialTemp: 260,
  coolingFactor: 0.985,
  minTemp: 1,
  componentSpacing: 180,
};

export const INITIAL_GRAPH_ZOOM = 1;
export const MIN_GRAPH_ZOOM = 0.05;
export const MAX_GRAPH_ZOOM = 2.4;
export const ZOOM_STEP = 1.22;

// Deterministic dependency layers give CoSE a reproducible, directional seed.
// This affects positions only; all elements, including cyclic edges, stay intact.
export function runDependencyLayout(cy: Core) {
  const depth = new Map<string, number>();
  const queue = cy.nodes(".application-root").toArray();
  queue.forEach(node => depth.set(node.id(), 0));
  for (let index = 0; index < queue.length; index++) {
    const node = queue[index];
    node.outgoers("node").forEach(child => {
      if (!depth.has(child.id())) {
        depth.set(child.id(), (depth.get(node.id()) || 0) + 1);
        queue.push(child);
      }
    });
  }
  const fallback = Math.max(0, ...depth.values()) + 1;
  const layers = new Map<number, string[]>();
  cy.nodes().forEach(node => {
    const level = depth.get(node.id()) ?? fallback;
    layers.set(level, [...(layers.get(level) || []), node.id()]);
  });
  const positions = new Map<string, { x: number; y: number }>();
  for (const [level, ids] of [...layers].sort(([a], [b]) => a - b)) {
    const parentX = (id: string) => {
      const xs = cy.getElementById(id).incomers("node").map(n => positions.get(n.id())?.x).filter((x): x is number => x !== undefined);
      return xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : 0;
    };
    ids.sort((a, b) => parentX(a) - parentX(b) || a.localeCompare(b));
    ids.forEach((id, index) => positions.set(id, { x: (index - (ids.length - 1) / 2) * 240, y: level * 240 }));
  }
  cy.nodes().positions(node => positions.get(node.id())!);
  // animate:false is synchronous: set the viewport after run() returns.
  cy.layout(dependencyLayout).run();
  cy.zoom(INITIAL_GRAPH_ZOOM);
  const roots = cy.nodes(".application-root");
  cy.center(roots.empty() ? cy.nodes() : roots);
}
