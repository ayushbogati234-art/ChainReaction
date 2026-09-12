import type { StylesheetJson, Core } from "cytoscape";

export const graphStyle: StylesheetJson = [
  { selector: "node", style: {
    width: 32, height: 32, "background-color": "#23845b", label: "data(displayLabel)",
    color: "#e1eaf3", "font-size": 13, "font-weight": 500,
    "text-valign": "bottom", "text-margin-y": 10,
    "text-wrap": "wrap", "text-max-width": "180px", "text-overflow-wrap": "anywhere",
    "text-outline-color": "#101820", "text-outline-width": 2,
    "min-zoomed-font-size": 9,
  } },
  { selector: "edge", style: {
    width: 1.6, "line-color": "#526a7d", "target-arrow-color": "#718ba0",
    "target-arrow-shape": "triangle", "arrow-scale": 1.25, "curve-style": "bezier",
    "control-point-step-size": 45, opacity: 0.75,
  } },
  { selector: ".is-dimmed", style: { opacity: 0.14, "text-opacity": 0.12 } },
  { selector: "node.is-related", style: {
    "background-color": "#3d9c96", "border-width": 2, "border-color": "#b2e8e1",
    opacity: 1, "text-opacity": 1,
  } },
  { selector: "edge.is-related", style: {
    width: 3, "line-color": "#62c2bc", "target-arrow-color": "#62c2bc", opacity: 1, "z-index": 5,
  } },
  { selector: "node.is-affected", style: {
    "background-color": "#9a6a28", "border-width": 2, "border-color": "#f1c879", opacity: 1, "text-opacity": 1,
  } },
  { selector: "edge.is-affected", style: {
    width: 3, "line-color": "#d59b48", "target-arrow-color": "#d59b48", opacity: 1, "z-index": 6,
  } },
  { selector: "node.is-immediate-impact", style: {
    "background-color": "#c17d28", "border-width": 3, "border-color": "#ffe0a3", opacity: 1, "text-opacity": 1,
  } },
  { selector: ".is-path", style: { opacity: 1, "text-opacity": 1, "z-index": 10 } },
  { selector: "node.is-path", style: {
    "background-color": "#8b72c4", "border-width": 2, "border-color": "#d9ccfa",
  } },
  { selector: "edge.is-path", style: {
    width: 4, "line-color": "#b69cf3", "target-arrow-color": "#b69cf3", "arrow-scale": 1.4,
  } },
  { selector: "node.application-root", style: {
    shape: "round-rectangle", width: 60, height: 38, "background-color": "#1d6286",
    "border-width": 3, "border-color": "#8acbf0", "font-weight": 700,
  } },
  { selector: "node.is-focused", style: {
    "background-color": "#38bdf8", "border-width": 4, "border-color": "#e0f2fe",
    opacity: 1, "text-opacity": 1, "min-zoomed-font-size": 0, "z-index": 20,
    "text-background-color": "#101820", "text-background-opacity": 0.95, "text-background-padding": "4px",
  } },
];

export function refreshLabels(cy: Core, selectedID?: string) {
  const zoom = cy.zoom();
  cy.batch(() => cy.nodes().forEach(node => {
    const selected = node.id() === selectedID;
    node.data("displayLabel", selected || zoom >= 0.8 ? node.data("fullLabel") : node.data("shortLabel"));
    // The selected package's full label remains legible at overview zoom.
    if (selected) node.style("font-size", Math.max(13, 12 / zoom));
    else node.removeStyle("font-size");
  }));
}
