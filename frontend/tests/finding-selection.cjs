// Regression: package selection owns the inspector and cancels stale finding details.
const assert=require('node:assert/strict');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async()=>{
 const browser=await chromium.launch({headless:true});
 try{
  const page=await browser.newPage({viewport:{width:1280,height:900}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
  const black={id:'black-finding',scan_id:'fixture',package_id:'black@23.3.0',name:'black',version:'23.3.0',ecosystem:'PyPI',purl:'pkg:pypi/black@23.3.0',direct:true,advisory_id:'CVE-2024-21503',aliases:['CVE-2024-21503'],dependency_confidence:'HIGH',vulnerability_match_confidence:'HIGH',confidence:'HIGH',explanation:['Potential dependency exposure; fixture.'],advisories:[{id:'CVE-2024-21503',aliases:[],source:'OSV',summary:'fixture',details:'',severity:[],references:[],affected:[],fixed_versions:[],match_confidence:'HIGH',match_evidence:'fixture'}]};
  const attrs={id:'attrs@22.2.0',name:'attrs',version:'22.2.0',source:'poetry.lock',confidence:'HIGH'};
  await page.route('**/api/v1/**',async route=>{
   const u=new URL(route.request().url());let data;
   if(u.pathname==='/api/v1/projects')data={id:'project'};
   else if(u.pathname.endsWith('/scans')||u.pathname.endsWith('/status'))data={id:'fixture',status:'COMPLETED',stage:'COMPLETED',progress:100,incomplete:false,limitations:[],enrichment_status:'COMPLETE'};
   else if(u.pathname.endsWith('/graph'))data={nodes:[{id:'app:fixture',name:'fixture',source:'pyproject.toml',input_files:['pyproject.toml','poetry.lock']},{id:'black@23.3.0',name:'black',version:'23.3.0',source:'poetry.lock',confidence:'HIGH'},attrs],edges:[{source:'app:fixture',target:'black@23.3.0',direct:true,constraint:'23.3.0',group:'main'},{source:'app:fixture',target:'attrs@22.2.0',direct:true,constraint:'22.2.0',group:'main'}]};
   else if(u.pathname.endsWith('/findings'))data={status:'COMPLETE',provider:'OSV',graph_status:'COMPLETE',packages:[{package_id:'black@23.3.0',status:'CHECKED'},{package_id:'attrs@22.2.0',status:'CHECKED'}],findings:[black]};
   else if(u.pathname.endsWith('/findings/black-finding')){await new Promise(resolve=>setTimeout(resolve,350));data=black;}
   else if(u.pathname.endsWith('/impact')){const id=decodeURIComponent(u.pathname.split('/').at(-2));data={package_id:id,downstream_packages:0,immediate_dependent_packages:[],transitive_dependent_packages:[],applications_exposed:1,application_ids:['app:fixture'],affected_node_ids:['app:fixture',id],reachability_exact:true,evidence_confidence:'HIGH',evidence_complete:true,limitations:[],paths:{paths:[['app:fixture',id]],path_lengths:[1],applications:['app:fixture'],limit:20,truncated:false,total_paths:1,count_exact:true,count_capped:false,count_limit:1000,maximum_depth:1,maximum_depth_exact:true,max_depth_limit:100,depth_limited:false,cycles_detected:false,evidence_confidence:'HIGH',evidence_complete:true,limitations:[]}};}
   else if(u.pathname.includes('/packages/')){const id=decodeURIComponent(u.pathname.split('/').pop());const pkg=id===attrs.id?attrs:{id:'black@23.3.0',name:'black',version:'23.3.0',source:'poetry.lock',confidence:'HIGH'};data={package:pkg,dependencies:[],dependents:[{id:'app:fixture',name:'fixture'}]};}
   else throw Error('unhandled '+u.pathname);
   await route.fulfill({status:200,json:data});
  });
  await page.goto(process.env.TEST_FRONTEND_URL||'http://localhost:3000');await page.getByLabel('Public GitHub repository').fill('https://github.com/example/fixture');await page.getByRole('button',{name:'Start scan',exact:true}).click();
  await page.getByRole('heading',{name:'Vulnerability findings',exact:true}).waitFor();
  const panel=page.getByRole('region',{name:'Vulnerability findings',exact:true});const blackButton=panel.getByRole('button',{name:/black@23\.3\.0.*CVE-2024-21503/});
  // A finding selects its own package and leaves enough time to begin its delayed detail request.
  await blackButton.click();await page.locator('.inspector').getByRole('heading',{name:'black@23.3.0',exact:true}).waitFor();
  await page.getByRole('button',{name:'attrs@22.2.0',exact:true}).click();
  await page.locator('.inspector').getByRole('heading',{name:'attrs@22.2.0',exact:true}).waitFor();await page.waitForTimeout(500);
  assert.equal(await page.getByRole('region',{name:'Finding details'}).count(),0,'black detail returned after attrs selection');
  assert.equal(await blackButton.getAttribute('aria-pressed'),'false','stale finding stays selected');
  assert.match(await panel.innerText(),/Selected package coverage: CHECKED/);
  // Rapid black -> attrs -> black selection cannot resurrect detail for the wrong package.
  await blackButton.click();await page.getByRole('button',{name:'attrs@22.2.0',exact:true}).click();await blackButton.click();
  await page.getByRole('region',{name:'Finding details'}).waitFor();assert.match(await page.getByRole('region',{name:'Finding details'}).innerText(),/CVE-2024-21503/);
  await page.getByRole('button',{name:'Show dependency paths',exact:true}).click();await page.locator('.inspector .path').first().waitFor();
  assert.match(await page.locator('.inspector .path').first().innerText(),/black@23\.3\.0/);
  assert.deepEqual(errors,[]);console.log('PASS: finding/package synchronization, delayed detail cancellation, rapid changes, package selection, paths');
 }finally{await browser.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
