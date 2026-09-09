// Default: deterministic API fixtures for findings UI states.
// TEST_LIVE=1: real URL submission, worker, GitHub, OSV, databases and UI.
const assert=require('node:assert/strict');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fixture=require('../../fixtures/python/vulnerabilities/affected.json');
const api=process.env.TEST_API_URL || 'http://localhost:8080';
(async()=>{
 const browser=await chromium.launch({headless:true});
 try {
  const page=await browser.newPage({viewport:{width:1440,height:1000}});
  const errors=[];page.on('pageerror',e=>errors.push(e.message));
  const live=process.env.TEST_LIVE==='1';
  let state='COMPLETE';
  const finding={id:'fixture-finding',scan_id:'fixture',package_id:'a@1.0.0',name:'a',version:'1.0.0',ecosystem:'PyPI',purl:'pkg:pypi/a@1.0.0',direct:true,advisory_id:'CVE-2099-0001',aliases:['CVE-2099-0001','GHSA-fixture-0001','PYSEC-fixture-0001'],dependency_confidence:'HIGH',vulnerability_match_confidence:'HIGH',confidence:'HIGH',explanation:['Potential dependency exposure; test fixture only.'],advisories:fixture.vulns.map((v,i)=>({...v,source:'OSV',fixed_versions:i?[]:['2.0.0'],match_confidence:'HIGH',match_evidence:'Exact PyPI name and locked version explicitly appear in affected versions.'}))};
  if(!live)await page.route('**/api/v1/**',async route=>{
   const u=new URL(route.request().url());let data;
   if(u.pathname==='/api/v1/projects')data={id:'project'};
   else if(u.pathname.endsWith('/scans') || u.pathname.endsWith('/status'))data={id:'fixture',status:'COMPLETED',stage:'COMPLETED',progress:100,incomplete:true,limitations:['Unselected group conditions'],enrichment_status:state};
   else if(u.pathname.endsWith('/graph'))data={nodes:[{id:'app:fixture',name:'fixture',input_files:['pyproject.toml','poetry.lock'],source:'pyproject.toml'},{id:'a@1.0.0',name:'a',version:'1.0.0',source:'poetry.lock',confidence:'HIGH'}],edges:[{source:'app:fixture',target:'a@1.0.0',direct:true,constraint:'1.0.0',group:'main'}]};
   else if(u.pathname.endsWith('/findings'))data={status:state,provider:'OSV',graph_status:'COMPLETE',packages:[{package_id:'a@1.0.0',status:state==='COMPLETE'?'CHECKED':state==='PARTIAL'?'PARTIAL':'UNQUERIED',reason:state==='COMPLETE'?'':'provider rate limited'}],findings:state==='DEGRADED'?[]:[finding]};
   else if(u.pathname.includes('/findings/'))data=finding;
   else if(u.pathname.endsWith('/paths'))data={paths:[['app:fixture','a@1.0.0']],limit:20,truncated:false};
   else if(u.pathname.includes('/packages/'))data={package:{id:'a@1.0.0',name:'a',version:'1.0.0',source:'poetry.lock',confidence:'HIGH'},dependencies:[],dependents:[{id:'app:fixture',name:'fixture'}]};
   else throw new Error('Unhandled fixture route '+u.pathname);
   await route.fulfill({status:200,json:data});
  });
  async function submit(){await page.goto(process.env.TEST_FRONTEND_URL || 'http://localhost:3000');await page.getByLabel('Public GitHub repository').fill('https://github.com/Hasenpfote/python-poetry-example');const response=page.waitForResponse(r=>r.url().endsWith('/scans') && r.request().method()==='POST');await page.getByRole('button',{name:'Start scan',exact:true}).click();const scan=await(await response).json();await page.waitForFunction(()=>document.querySelector('[role="status"]')?.textContent.includes('COMPLETED'),{},{timeout:150000});await page.getByRole('heading',{name:'Vulnerability findings',exact:true}).waitFor();return scan.id;}
  const scanID=await submit();
  const panel=page.getByRole('region',{name:'Vulnerability findings',exact:true});
  await panel.getByRole('button').first().waitFor({timeout:30000});
  const before=await page.locator('.graph-canvas').evaluate(el=>{const cy=el._cyreg.cy;return {zoom:cy.zoom(),pan:cy.pan(),positions:cy.nodes().map(n=>({id:n.id(),position:n.position()}))};});
  await panel.getByRole('button').first().click();
  await page.getByRole('region',{name:'Finding details'}).waitFor();
  assert.match(await panel.innerText(),/Potential dependency exposure/);
  assert.match(await panel.innerText(),/Dependency evidence confidence/);
  assert.match(await panel.innerText(),/Vulnerability-match confidence/);
  if(!live){assert.match(await panel.innerText(),/Not supplied/);assert.match(await panel.innerText(),/2.0.0/);assert.equal(await panel.getByRole('link').count(),1);}
  await page.getByRole('button',{name:'Show dependency paths',exact:true}).click();
  await page.locator('.inspector .path').first().waitFor();
  await page.locator('.inspector .path').first().click();
  const after=await page.locator('.graph-canvas').evaluate(el=>{const cy=el._cyreg.cy;return {zoom:cy.zoom(),pan:cy.pan(),positions:cy.nodes().map(n=>({id:n.id(),position:n.position()}))};});
  assert.deepEqual(after,before,'finding inspection changed graph layout/viewport');
  await page.setViewportSize({width:390,height:900});await panel.scrollIntoViewIfNeeded();assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  if(process.env.TEST_SCREENSHOT)await page.screenshot({path:process.env.TEST_SCREENSHOT,fullPage:true});
  if(live){
   const scan=await(await page.request.get(api+'/api/v1/scans/'+scanID)).json();const report=await(await page.request.get(api+'/api/v1/scans/'+scanID+'/findings')).json();
   console.log(JSON.stringify({mode:'LIVE',scanID,commit:scan.commit_sha,enrichment:report.status,checked:report.packages.filter(p=>p.status==='CHECKED').length,total:report.packages.length,findings:report.findings.length,limitations:report.packages.filter(p=>p.status!=='CHECKED').map(p=>({id:p.package_id,status:p.status,reason:p.reason}))},null,2));
  }else{
   for(state of ['PARTIAL','DEGRADED']){await submit();await panel.getByText('Coverage is incomplete or unavailable.',{exact:false}).waitFor();assert.ok(!(await panel.innerText()).includes('No matching advisories returned'));if(state==='DEGRADED')assert.equal(await panel.getByRole('button').count(),0);}
   console.log('FIXTURE PASS: missing severity/fix, alias evidence, paths, stable layout, narrow UI, partial/degraded coverage');
  }
  assert.deepEqual(errors,[]);
 }finally{await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
