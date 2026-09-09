// Read-only browser verification of the reproduced issue against a persisted scan.
const assert=require('node:assert/strict');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const api=process.env.TEST_API_URL || 'http://localhost:8080';
const scanID=process.env.TEST_SCAN_ID || '579acc06-c7c0-45c1-8bba-a5eb0002d2db';
(async()=>{
 const browser=await chromium.launch({headless:true});
 try{
  const page=await browser.newPage({viewport:{width:1440,height:1000}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
  const scan=await (await page.request.get(api+'/api/v1/scans/'+scanID)).json();assert.equal(scan.status,'COMPLETED');
  await page.route('**/api/v1/**',async route=>{
   if(route.request().method()!=='POST')return route.continue();
   if(route.request().url().endsWith('/projects'))return route.fulfill({status:201,json:{id:scan.project_id}});
   return route.fulfill({status:202,json:scan});
  });
  await page.goto(process.env.TEST_FRONTEND_URL||'http://localhost:3000');await page.getByLabel('Public GitHub repository').fill('https://github.com/Hasenpfote/python-poetry-example');await page.getByRole('button',{name:'Start scan',exact:true}).click();
  const panel=page.getByRole('region',{name:'Vulnerability findings',exact:true});const black=panel.getByRole('button',{name:/black@23\.3\.0.*CVE-2024-21503/});await black.waitFor();
  await black.click();await page.getByRole('region',{name:'Finding details'}).waitFor();assert.match(await page.getByRole('region',{name:'Finding details'}).innerText(),/CVE-2024-21503/);
  await page.getByRole('button',{name:'attrs@22.2.0',exact:true}).click();await page.locator('.inspector').getByRole('heading',{name:'attrs@22.2.0',exact:true}).waitFor();await page.waitForTimeout(250);
  assert.equal(await page.getByRole('region',{name:'Finding details'}).count(),0,'black finding remained after attrs selection');assert.equal(await black.getAttribute('aria-pressed'),'false');
  await black.click();await page.getByRole('region',{name:'Finding details'}).waitFor();await page.getByRole('button',{name:'Show dependency paths',exact:true}).click();await page.locator('.inspector .path').first().waitFor();
  assert.match(await page.locator('.inspector .path').first().innerText(),/black@23\.3\.0/);assert.deepEqual(errors,[]);
  console.log(JSON.stringify({result:'PASS',scanID,sequence:'black CVE-2024-21503 -> attrs -> black -> paths'},null,2));
 }finally{await browser.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
