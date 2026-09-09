import type { Core } from "cytoscape";

const clearClasses = (cy: Core) => cy.elements().removeClass("is-focused is-related is-path is-dimmed");

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
