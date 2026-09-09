"use client";

import { useEffect, useRef, useState } from "react";
import cytoscape, { type Core } from "cytoscape";
import { focusPackage, focusPath } from "../../lib/graph/focus";
import { runDependencyLayout, INITIAL_GRAPH_ZOOM, MAX_GRAPH_ZOOM, MIN_GRAPH_ZOOM, ZOOM_STEP } from "../../lib/graph/layout";
import { graphStyle, refreshLabels } from "../../lib/graph/style";
import type { GraphData } from "../../types/scan";
import styles from "./GraphWorkspace.module.css";

type Props = {
  graph?: GraphData;
  selectedID?: string;
  activePath?: string[];
  onSelect: (id: string) => void;
};

export function DependencyGraph({ graph, selectedID, activePath, onSelect }: Props) {
  const host = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const selectRef = useRef(onSelect);
  const selectedRef = useRef(selectedID);
  const [zoom, setZoom] = useState(INITIAL_GRAPH_ZOOM);

  useEffect(() => { selectRef.current = onSelect; }, [onSelect]);

  useEffect(() => {
    if (!host.current || !graph) return;
    const cy = cytoscape({
      container: host.current,
      layout: { name: "preset" },
      minZoom: MIN_GRAPH_ZOOM,
      maxZoom: MAX_GRAPH_ZOOM,
      wheelSensitivity: 0.18,
      boxSelectionEnabled: false,
      elements: [
        ...graph.nodes.map(node => ({
          data: {
            id: node.id,
            fullLabel: node.version ? node.name + "\n@" + node.version : "Application\n" + node.name,
            shortLabel: node.name,
            displayLabel: node.version ? node.name + "\n@" + node.version : "Application\n" + node.name,
          },
          classes: node.id.startsWith("app:") ? "application-root" : "package-node",
        })),
        ...graph.edges.map((edge, index) => ({
          data: { id: "edge:" + index, source: edge.source, target: edge.target },
        })),
      ],
      style: graphStyle,
    });
    cyRef.current = cy;
    cy.on("tap", "node", event => {
      const id = event.target.id();
      if (!id.startsWith("app:")) selectRef.current(id);
    });
    cy.on("zoom", () => {
      refreshLabels(cy, selectedRef.current);
      setZoom(cy.zoom());
    });
    runDependencyLayout(cy);
    refreshLabels(cy, selectedRef.current);
    setZoom(cy.zoom());
    const observer = new ResizeObserver(() => cy.resize());
    observer.observe(host.current);
    return () => { observer.disconnect(); cy.destroy(); cyRef.current = null; };
    // Only a different graph replaces the Cytoscape instance or starts layout.
  }, [graph]);

  useEffect(() => {
    selectedRef.current = selectedID;
    const cy = cyRef.current;
    if (!cy) return;
    if (activePath?.length) focusPath(cy, activePath);
    else focusPackage(cy, selectedID);
    refreshLabels(cy, selectedID);
  }, [graph, activePath, selectedID]);

  const withGraph = (action: (cy: Core) => void) => () => {
    if (cyRef.current) action(cyRef.current);
  };
  const resetLayout = withGraph(cy => {
    cy.nodes().removeStyle("font-size");
    cy.nodes().forEach(node => { node.data("displayLabel", node.data("fullLabel")); });
    runDependencyLayout(cy);
    refreshLabels(cy, selectedRef.current);
  });
  const zoomBy = (factor: number) => withGraph(cy => cy.zoom({
    level: Math.max(MIN_GRAPH_ZOOM, Math.min(MAX_GRAPH_ZOOM, cy.zoom() * factor)),
    renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 },
  }));

  return <section className={styles.shell} aria-label="Dependency graph workspace">
    <div className={styles.toolbar} role="toolbar" aria-label="Graph controls">
      <button type="button" disabled={!graph} onClick={withGraph(cy => cy.fit(cy.elements(), 80))}>Fit graph</button>
      <button type="button" disabled={!graph || zoom >= MAX_GRAPH_ZOOM} onClick={zoomBy(ZOOM_STEP)}>Zoom in</button>
      <button type="button" disabled={!graph || zoom <= MIN_GRAPH_ZOOM} onClick={zoomBy(1 / ZOOM_STEP)}>Zoom out</button>
      <button type="button" disabled={!graph} onClick={resetLayout}>Reset layout</button>
      <output className={styles.zoom} aria-label="Graph zoom">{Math.round(zoom * 100)}%</output>
    </div>
    <div className={styles.legend}>
      <span><i className={styles.rootKey} aria-hidden="true" />Application</span>
      <span><i className={styles.packageKey} aria-hidden="true" />Package</span>
      <span>Dependent → dependency</span>
    </div>
    <div className={styles.canvasArea} aria-label="Dependency graph" data-nodes={graph?.nodes.length || 0} data-edges={graph?.edges.length || 0}>
      <div ref={host} className={"graph-canvas " + styles.canvas} />
      {!graph && <p className={styles.empty}>Submit a supported public repository to inspect its dependency graph.</p>}
    </div>
    <div className={styles.hint}>
      <span>Drag background to pan · Scroll to zoom</span>
      <span>{activePath?.length ? "Purple: selected path" : selectedID ? "Teal: immediate dependencies and dependents" : "Explore at readable zoom, or Fit graph for an overview"}</span>
    </div>
  </section>;
}
