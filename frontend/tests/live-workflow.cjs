// Run with PLAYWRIGHT_MODULE pointing to an installed Playwright package.
// All API, database, worker, and GitHub requests are real; no routes are mocked.
const assert = require("node:assert/strict");
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");

(async () => {
 const browser = await chromium.launch({ headless:true });
 const page = await browser.newPage({ viewport:{width:1440,height:1000} });
 const errors = [];
 page.on("pageerror", e=>errors.push(e.message));
 const observed=[];
 page.on("response", async response => {
  if (response.url().includes("/api/v1/") && response.request().method()!=="OPTIONS") {
   try { const body=await response.json(); observed.push({url:response.url(),code:response.status(),body}); } catch {}
  }
 });
 try {
  await page.goto(process.env.TEST_FRONTEND_URL || "http://localhost:3000");
  const repo=process.env.TEST_GITHUB_URL || "https://github.com/Hasenpfote/python-poetry-example";
  await page.getByLabel("Public GitHub repository").fill(repo);
  await page.getByRole("button",{name:"Start scan",exact:true}).click();
  await page.waitForFunction(()=>/COMPLETED|FAILED/.test(document.querySelector('[role="status"]')?.textContent || ""),{},{timeout:150000});
  const status=await page.getByRole("status").innerText();
  console.log("SCAN STATUS:",status);
  assert.match(status,/COMPLETED/);
  await page.waitForFunction(()=>Number(document.querySelector('[aria-label="Dependency graph"]')?.getAttribute("data-nodes"))>0);
  const counts=await page.getByLabel("Dependency graph",{exact:true}).evaluate(el=>({nodes:Number(el.dataset.nodes),edges:Number(el.dataset.edges),canvases:el.querySelectorAll("canvas").length}));
  assert.ok(counts.nodes>1 && counts.edges>0 && counts.canvases>0,"graph canvas missing");
  const graphHost=page.locator(".graph-canvas");
  const beforeControls=await graphHost.evaluate(el=>({zoom:el._cyreg.cy.zoom(),pan:el._cyreg.cy.pan(),nodes:el._cyreg.cy.nodes().length,edges:el._cyreg.cy.edges().length}));
  await page.getByRole("button",{name:"Zoom in",exact:true}).click();
  const afterZoomIn=await graphHost.evaluate(el=>el._cyreg.cy.zoom());
  assert.ok(afterZoomIn>beforeControls.zoom,"zoom in did not increase zoom");
  await page.getByRole("button",{name:"Zoom out",exact:true}).click();
  await page.getByRole("button",{name:"Fit graph",exact:true}).click();
  const afterFit=await graphHost.evaluate(el=>({nodes:el._cyreg.cy.nodes().length,edges:el._cyreg.cy.edges().length}));
  assert.deepEqual(afterFit,{nodes:beforeControls.nodes,edges:beforeControls.edges},"graph controls removed elements");
  await page.getByRole("button",{name:"Reset layout",exact:true}).click();
  await page.waitForTimeout(1000);
  console.log("GRAPH CONTROLS: PASS");
  const choices=page.locator(".package-choice");
  const target=choices.filter({hasText:/pytest@/}).first();
  const pick=(await target.count())?target:choices.first();
  const name=await pick.innerText();
  const selectedID=name;
  const positionBefore=await graphHost.evaluate((el,id)=>el._cyreg.cy.getElementById(id).position(),selectedID);
  await pick.click();
  await page.locator(".inspector .path").first().waitFor();
  const focus=await graphHost.evaluate((el,id)=>({position:el._cyreg.cy.getElementById(id).position(),focused:el._cyreg.cy.getElementById(id).hasClass("is-focused"),related:el._cyreg.cy.elements(".is-related").length,dimmed:el._cyreg.cy.elements(".is-dimmed").length}),selectedID);
  assert.deepEqual(focus.position,positionBefore,"package inspection reran the layout");
  assert.ok(focus.focused && focus.related>0 && focus.dimmed>0,"package focus did not highlight the local neighborhood");
  console.log("PACKAGE FOCUS: PASS");
  const pathText=await page.locator(".inspector .path").allTextContents();
  assert.ok(pathText.some(p=>p.includes(name)),"selected package is missing from paths");
  await page.locator(".inspector .path").first().click();
  const pathFocus=await graphHost.evaluate(el=>({nodes:el._cyreg.cy.elements("node.is-path").length,edges:el._cyreg.cy.elements("edge.is-path").length,dimmed:el._cyreg.cy.elements(".is-dimmed").length}));
  assert.ok(pathFocus.nodes>1 && pathFocus.edges>0 && pathFocus.dimmed>0,"path highlight is incomplete");
  console.log("PATH FOCUS: PASS");
  // Exercise the Cytoscape node tap as well as the accessible package button.
  const node=await graphHost.evaluate(el=>{
   const cy=el._cyreg?.cy;
   if (!cy) throw new Error("Cytoscape instance missing");
   const chosen=cy.nodes().filter(n=>n.id().startsWith("isort@"))[0] || cy.nodes().filter(n=>!n.id().startsWith("app:"))[0];
   cy.center(chosen);
   const p=chosen.renderedPosition(); const rect=el.getBoundingClientRect();
   return {id:chosen.id(),x:rect.x+p.x,y:rect.y+p.y};
  });
  await page.mouse.click(node.x,node.y);
  await page.waitForFunction(id=>document.querySelector(".inspector")?.textContent.includes(id),node.id);
  assert.ok((await page.locator(".inspector").innerText()).includes(node.id));
  const box=await graphHost.boundingBox();
  const panBefore=await graphHost.evaluate(el=>el._cyreg.cy.pan());
  await page.mouse.move(box.x+18,box.y+box.height-18); await page.mouse.down(); await page.mouse.move(box.x+88,box.y+box.height-53); await page.mouse.up();
  const panAfter=await graphHost.evaluate(el=>el._cyreg.cy.pan());
  assert.notDeepEqual(panAfter,panBefore,"graph drag did not pan the viewport");
  console.log("PAN: PASS");
  await page.setViewportSize({width:760,height:900});
  await page.locator(".inspector").scrollIntoViewIfNeeded();
  assert.ok(await page.locator(".inspector").isVisible(),"inspector is hidden at narrower width");
  if(process.env.TEST_SCREENSHOT) await page.screenshot({path:process.env.TEST_SCREENSHOT,fullPage:true});
  const scanResponse=observed.find(r=>r.url.endsWith("/scans") && r.code===202);
  assert.ok(scanResponse,"scan was not enqueued");
  const scanID=scanResponse.body.id;
  const full=await (await page.request.get("http://localhost:8080/api/v1/scans/"+scanID)).json();
  assert.match(full.commit_sha,/^[a-f0-9]{40}$/);
  const incomplete=await page.getByRole("heading",{name:"Incomplete result"}).count();
  if(full.incomplete) assert.equal(incomplete,1,"incomplete warning missing");
  console.log(JSON.stringify({repo,scanID,commit:full.commit_sha,counts,incomplete:full.incomplete,limitations:full.limitations?.length,selected:name,paths:pathText,graphNodeClicked:node.id},null,2));
  // A metadata-only unsupported repository must fail visibly.
  await page.getByLabel("Public GitHub repository").fill("https://github.com/octocat/Hello-World");
  await page.getByRole("button",{name:"Start scan",exact:true}).click();
  await page.waitForFunction(()=>/FAILED/.test(document.querySelector('[role="status"]')?.textContent || ""),{},{timeout:150000});
  const failure=await page.getByRole("status").innerText();
  assert.match(failure,/UNSUPPORTED_INPUT/);
  console.log("UNSUPPORTED UI:",failure);
  assert.equal(await page.locator(".inspector .path").count(),0,"stale paths survived new scan");
  assert.deepEqual(errors,[]);
  console.log("PASS: live browser URL submission, queue, worker, completed scan, graph canvas, package selection, dependency paths, incomplete warning, unsupported failure, no browser errors");
 } finally { await browser.close(); }
})().catch(error=>{ console.error(error); process.exitCode=1; });
