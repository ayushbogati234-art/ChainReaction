// Read-only browser regression against a persisted scan. Only the two creation
// responses are replayed to open that scan; graph/inspector/path GETs are live.
// PLAYWRIGHT_MODULE may point to an external Playwright installation.
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const api = process.env.TEST_API_URL || 'http://localhost:8080';
const scanID = process.env.TEST_SCAN_ID || '5284dce3-a714-4f11-91e6-bb8ad93112a5';
(async () => {
 const browser = await chromium.launch({headless:true});
 try {
  const page = await browser.newPage({viewport:{width:1440,height:1000}});
  const errors=[]; page.on('pageerror',e=>errors.push(e.message));
  const get = async path => { const r=await page.request.get(api+path); assert.ok(r.ok(),path); return r.json(); };
  const scan=await get('/api/v1/scans/'+scanID);
  assert.equal(scan.status,'COMPLETED');
  const graph=await get('/api/v1/scans/'+scanID+'/graph');
  await page.route('**/api/v1/**',async route=>{
   if(route.request().method()!=='POST') return route.continue();
   const body=route.request().url().endsWith('/projects')?{id:scan.project_id}:{...scan,status:'QUEUED',stage:'QUEUED',commit_sha:''};
   await route.fulfill({status:200,json:body});
  });
  await page.goto(process.env.TEST_FRONTEND_URL || 'http://localhost:3000');
  await page.getByLabel('Public GitHub repository').fill('https://github.com/Hasenpfote/python-poetry-example');
  await page.getByRole('button',{name:'Start scan',exact:true}).click();
  await page.waitForFunction(()=>document.querySelector('.graph-canvas')?._cyreg?.cy?.nodes().length>0);
  const host=page.locator('.graph-canvas');
  const evidence=page.locator('details').filter({has:page.locator('summary',{hasText:'Supported inputs and scan provenance'})}).first();
  await evidence.locator('summary').first().click();
  const evidenceText=await evidence.innerText();
  assert.match(evidenceText,/tox.ini.*not parsed/);
  assert.match(evidenceText,/Direct dependency declarations and constraints come from pyproject.toml/);
  const application=graph.nodes.find(n=>n.id.startsWith('app:'));
  if(application.input_files?.length) {
   for(const file of application.input_files) assert.ok(evidenceText.includes(file+' (repository root)'));
   assert.ok(evidenceText.includes(scan.commit_sha));
  } else assert.match(evidenceText,/predates recorded file provenance/);
  await evidence.locator('summary').first().click();
  const state=()=>host.evaluate(el=>{const cy=el._cyreg.cy;return {zoom:cy.zoom(),pan:cy.pan(),positions:cy.nodes().map(n=>({id:n.id(),p:n.position()}))};});
  const elements=()=>host.evaluate(el=>{const cy=el._cyreg.cy;return {nodes:cy.nodes().map(n=>n.id()).sort(),edges:cy.edges().map(e=>[e.source().id(),e.target().id()].join(' -> ')).sort()};});
  const expected={nodes:graph.nodes.map(n=>n.id).sort(),edges:graph.edges.map(e=>[e.source,e.target].join(' -> ')).sort()};
  assert.deepEqual(await elements(),expected);
  const metrics=await host.evaluate(el=>{
   const cy=el._cyreg.cy, nodes=cy.nodes(); let overlaps=0;
   for(let i=0;i<nodes.length;i++) for(let j=i+1;j<nodes.length;j++) {
    const a=nodes[i].boundingBox(), b=nodes[j].boundingBox();
    if(a.x1<b.x2 && b.x1<a.x2 && a.y1<b.y2 && b.y1<a.y2) overlaps++;
   }
   const lengths=cy.edges().map(e=>{const a=e.source().position(),b=e.target().position();return Math.hypot(a.x-b.x,a.y-b.y);}).sort((a,b)=>a-b);
   return {overlaps,medianEdgeLength:lengths[Math.floor(lengths.length/2)],rootShape:cy.nodes('.application-root').style('shape')};
  });
  assert.equal(metrics.overlaps,0,'node/label bounding boxes overlap');
  assert.equal(metrics.rootShape,'round-rectangle');
  const initial=await state();
  await page.screenshot({path:(process.env.TEST_SCREENSHOT_PREFIX || '/tmp/graph')+'-initial.png'});
  await page.getByRole('button',{name:'Zoom in',exact:true}).click(); assert.ok((await state()).zoom>initial.zoom);
  await page.getByRole('button',{name:'Zoom out',exact:true}).click();
  await page.getByRole('button',{name:'Fit graph',exact:true}).click();
  assert.ok((await state()).zoom<initial.zoom);
  await page.screenshot({path:process.env.TEST_SCREENSHOT_PREFIX?process.env.TEST_SCREENSHOT_PREFIX+'-overview.png':'/tmp/graph-overview.png'});
  await page.getByRole('button',{name:'Reset layout',exact:true}).click();
  // Reset must undo a manually moved node without replacing graph elements.
  await host.evaluate(el=>el._cyreg.cy.nodes()[0].position({x:-9000,y:-9000}));
  await page.getByRole('button',{name:'Reset layout',exact:true}).click();
  assert.deepEqual(await state(),initial,'reset did not restore the deterministic layout');
  const before=await state();
  await page.locator('.package-choice').filter({hasText:/^pytest@/}).click();
  await page.locator('.inspector .path').first().waitFor();
  assert.match(await page.locator('.inspector').innerText(),/Package-version evidence: poetry.lock/);
  assert.match(await page.locator('.inspector').innerText(),/Direct declaration: pyproject.toml/);
  assert.deepEqual(await state(),before,'inspection changed positions/viewport');
  assert.ok(await host.evaluate(el=>{const cy=el._cyreg.cy,n=cy.nodes('.is-focused');return n.length===1 && n.closedNeighborhood().every(e=>!e.hasClass('is-dimmed')) && cy.elements('.is-dimmed').length>0;}));
  await page.locator('.inspector .path').last().click();
  assert.deepEqual(await state(),before,'path selection changed positions/viewport');
  const route=(await page.locator('.inspector .path[aria-pressed="true"]').innerText()).replace(/^Length \d+: /,'').split(' -> ');
  assert.deepEqual(await host.evaluate(el=>el._cyreg.cy.nodes('.is-path').map(n=>n.id()).sort()),[...new Set(route)].sort());
  assert.ok(await host.evaluate((el,route)=>route.slice(1).every((id,i)=>el._cyreg.cy.edges('.is-path').some(e=>e.source().id()===route[i] && e.target().id()===id)),route));
  // Choose a visible background point away from nodes and use a real drag.
  const point=await host.evaluate(el=>{const cy=el._cyreg.cy,r=el.getBoundingClientRect();for(let y=30;y<Math.min(r.height,innerHeight-r.y)-40;y+=30)for(let x=30;x<r.width-100;x+=30)if(cy.nodes().every(n=>{const p=n.renderedPosition();return Math.hypot(p.x-x,p.y-y)>55;}))return {x:r.x+x,y:r.y+y};});
  assert.ok(point); await page.mouse.move(point.x,point.y); await page.mouse.down(); await page.mouse.move(point.x+80,point.y+25,{steps:12}); await page.mouse.up();
  assert.notDeepEqual((await state()).pan,before.pan,'background drag did not pan');
  const wheelBefore=(await state()).zoom; await page.mouse.wheel(0,-150); await page.waitForTimeout(400); assert.notEqual((await state()).zoom,wheelBefore);
  await page.getByRole('button',{name:'Fit graph',exact:true}).click();
  assert.ok(await host.evaluate(el=>{const cy=el._cyreg.cy,n=cy.nodes('.is-focused');return Number(n.style('font-size').replace('px',''))*cy.zoom()>=11.9 && n.style('min-zoomed-font-size')==='0px';}),'selected label unreadable');
  await page.getByRole('button',{name:'Reset layout',exact:true}).click();
  await page.screenshot({path:(process.env.TEST_SCREENSHOT_PREFIX || '/tmp/graph')+'-detail.png'});
  // Actual canvas click, not just the accessible sidebar selection.
  const click=await host.evaluate(el=>{const cy=el._cyreg.cy,n=cy.nodes().filter(n=>n.id().startsWith('isort@'))[0];cy.center(n);const p=n.renderedPosition(),r=el.getBoundingClientRect();return {id:n.id(),x:r.x+p.x,y:r.y+p.y};});
  const beforeTap=await state();
  await page.mouse.click(click.x,click.y);
  await page.waitForFunction(id=>document.querySelector('.inspector h3')?.textContent===id,click.id);
  await page.locator('.inspector .path').first().waitFor();
  assert.deepEqual(await state(),beforeTap,'canvas selection changed positions/viewport');
  for(const width of [1000,760,390]) {
   await page.setViewportSize({width,height:900});
   await page.locator('.inspector').scrollIntoViewIfNeeded();
   assert.ok(await page.locator('.inspector').isVisible());
   assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),'horizontal page overflow at '+width);
   await page.locator('.inspector .path').first().click();
  }
  await page.screenshot({path:(process.env.TEST_SCREENSHOT_PREFIX || '/tmp/graph')+'-narrow.png',fullPage:true});
  assert.deepEqual(await elements(),expected); assert.deepEqual(errors,[]);
  console.log(JSON.stringify({result:'PASS',scanID,nodes:expected.nodes.length,edges:expected.edges.length,metrics,checks:'zoom, wheel, pan, fit, reset, selection stability, neighborhood, full directed path, selected label, responsive inspector, element preservation, no browser errors'},null,2));
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
