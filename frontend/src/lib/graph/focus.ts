import type { Core } from "cytoscape";

const clearClasses = (cy: Core) => cy.elements().removeClass("is-focused is-related is-affected is-immediate-impact is-path is-dimmed");

export function focusPackage(cy: Core, packageID?: string) {
  clearClasses(cy);
  if (!packageID) return;
  const selected = cy.getElementById(packageID);
  if (selected.empty()) return;
  const related = selected.closedNeighborhood();
  cy.elements().addClass("is-dimmed");
  related.removeClass("is-dimmed").addClass("is-related");
  selected.removeClass("is-dimmed").addClass("is-focused");
}

export function focusImpact(cy: Core, packageID?: string, affectedIDs: string[] = [], immediateIDs: string[] = []) {
  focusPackage(cy, packageID);
  if (!packageID || !affectedIDs.length) return;
  const affected = new Set(affectedIDs);
  cy.nodes().filter(node => affected.has(node.id())).removeClass("is-dimmed").addClass("is-affected");
  cy.edges().filter(edge => affected.has(edge.source().id()) && affected.has(edge.target().id())).removeClass("is-dimmed").addClass("is-affected");
  for (const id of immediateIDs) cy.getElementById(id).removeClass("is-dimmed").addClass("is-immediate-impact");
  cy.getElementById(packageID).removeClass("is-dimmed").addClass("is-focused");
}

export function focusPath(cy: Core, path?: string[]) {
  if (!path?.length) return focusPackage(cy);
  clearClasses(cy);
  cy.elements().addClass("is-dimmed");
  const nodes = cy.collection();
  for (const id of path) nodes.merge(cy.getElementById(id));
  const edges = cy.collection();
  for (let index = 0; index < path.length - 1; index++) {
    edges.merge(cy.edges().filter(edge => edge.source().id() === path[index] && edge.target().id() === path[index + 1]));
  }
  nodes.removeClass("is-dimmed").addClass("is-path");
  edges.removeClass("is-dimmed").addClass("is-path");
  cy.getElementById(path[path.length - 1]).removeClass("is-dimmed").addClass("is-focused");
}
