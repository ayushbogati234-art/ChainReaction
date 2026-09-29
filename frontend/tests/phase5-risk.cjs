// Independent UI fixtures for available/unknown risk and systemic findings.
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = process.env.TEST_BROWSER_CDP ? await chromium.connectOverCDP(process.env.TEST_BROWSER_CDP) : await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  try {
    const page = await context.newPage(), errors = [];
    page.on('pageerror', error => errors.push(error.message));
    let mode = 'systemic';
    const nodes = [{ id: 'app:fixture', name: 'fixture', source: 'pyproject.toml' }, { id: 'x@1', name: 'x', version: '1', source: 'poetry.lock', confidence: 'HIGH' }];
    const paths = { paths: [['app:fixture', 'x@1']], path_lengths: [1], applications: ['app:fixture'], total_paths: 1, count_exact: true, count_capped: false, maximum_depth: 1, maximum_depth_exact: true, evidence_complete: true, limitations: [] };
    const component = (value, weight, reason_codes = []) => ({ value, weight, contribution: value === null ? null : value * weight, reason_codes, evidence: ['Saved scan evidence'] });
    const systemic = () => ({ package_id: 'x@1', status: mode === 'systemic' ? 'MATCH' : mode === 'partial' ? 'UNAVAILABLE' : 'NOT_MATCHED', threshold: 80, criticality_score: 100, provider: 'OSV', checked_at: '2026-09-25T00:00:00Z', reason_codes: mode === 'partial' ? ['INCOMPLETE_ENRICHMENT_COVERAGE'] : [], explanation: ['No matching advisory is not evidence of safety.'] });
    await page.route('**/api/v1/**', async route => {
      const pathname = new URL(route.request().url()).pathname;
      let data;
      if (pathname.endsWith('/systemic-dependencies')) data = { status: mode === 'partial' ? 'PARTIAL' : 'COMPLETE', total_packages: 1, evaluated_packages: mode === 'partial' ? 0 : 1, findings: mode === 'systemic' ? [systemic()] : [], reason_codes: mode === 'partial' ? ['INCOMPLETE_ENRICHMENT_COVERAGE'] : [] };
      else if (pathname === '/api/v1/projects') data = { id: 'p' };
      else if (pathname.endsWith('/scans') || pathname.endsWith('/status')) data = { id: 'scan', status: 'COMPLETED', stage: 'COMPLETED', progress: 100, incomplete: false, limitations: [], enrichment_status: mode === 'partial' ? 'PARTIAL' : 'COMPLETE' };
      else if (pathname.endsWith('/graph')) data = { nodes, edges: [{ source: 'app:fixture', target: 'x@1' }] };
      else if (pathname.endsWith('/findings')) data = { status: 'COMPLETE', provider: 'OSV', graph_status: 'COMPLETE', packages: [{ package_id: 'x@1', status: 'CHECKED' }], findings: [] };
      else if (pathname.endsWith('/impact')) data = {
        package_id: 'x@1', downstream_packages: 0, immediate_dependent_packages: [], transitive_dependent_packages: [], applications_exposed: 1, application_ids: ['app:fixture'], affected_node_ids: ['app:fixture', 'x@1'], reachability_exact: true, evidence_confidence: 'HIGH', evidence_complete: true, limitations: [], paths,
        systemic_dependency: systemic(),
        risk: { model: 'phase5-v1', score: mode === 'scored' ? 90 : null, reason_codes: mode === 'scored' ? [] : ['EXPLOITABILITY_EVIDENCE_UNAVAILABLE'], explanation: ['Unknown inputs withhold the total.'], components: { vulnerability_severity: component(mode === 'systemic' ? null : 98, .35), blast_radius: component(70, .30), structural_criticality: component(100, .20), exploitability_evidence: component(mode === 'scored' ? 100 : null, .15, mode === 'scored' ? [] : ['EXPLOITABILITY_EVIDENCE_UNAVAILABLE']) } },
      };
      else if (pathname.includes('/packages/')) data = { package: nodes[1], dependencies: [], dependents: [nodes[0]] };
      else throw Error('Unexpected route: ' + pathname);
      await route.fulfill({ status: 200, json: data });
    });
    async function start() {
      await page.goto(process.env.TEST_FRONTEND_URL || 'http://localhost:3100');
      await page.getByLabel('Public GitHub repository').fill('https://github.com/example/phase5');
      await page.getByRole('button', { name: 'Start scan', exact: true }).click();
      await page.locator('.graph-canvas').waitFor();
      await page.getByRole('region', { name: 'Systemic Dependencies', exact: true }).getByText(/packages assessed/).waitFor();
    }
    await start();
    const host = page.locator('.graph-canvas');
    const state = () => host.evaluate(el => { const cy = el._cyreg.cy; return { zoom: cy.zoom(), pan: cy.pan(), positions: cy.nodes().map(n => [n.id(), n.position()]) }; });
    const before = await state();
    await page.getByRole('button', { name: 'x@1 — criticality 100/100', exact: true }).click();
    const risk = page.getByRole('region', { name: 'Overall risk evidence', exact: true });
    await risk.getByText('Overall risk unavailable', { exact: true }).waitFor();
    assert.match(await risk.innerText(), /Unknown · weight 15% · contribution unavailable/);
    assert.match(await risk.innerText(), /70.0\/100 · weight 30% · 21.00 points/);
    await page.getByRole('region', { name: 'Systemic Dependency evidence', exact: true }).getByText('No known advisory returned at scan time.', { exact: true }).waitFor();
    assert.deepEqual(await state(), before, 'systemic selection moved the graph');
    await page.setViewportSize({ width: 390, height: 900 });
    await risk.scrollIntoViewIfNeeded();
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), 'narrow UI overflow');
    if (process.env.TEST_SCREENSHOT) await page.screenshot({ path: process.env.TEST_SCREENSHOT, fullPage: true });
    await page.setViewportSize({ width: 1280, height: 900 });
    for (mode of ['scored', 'unknown', 'partial']) {
      await start();
      await page.getByRole('button', { name: 'x@1', exact: true }).click();
      await risk.waitFor();
      if (mode === 'scored') {
        assert.match(await risk.innerText(), /90\/100 — prototype prioritization score/);
        assert.match(await risk.innerText(), /100.0\/100 · weight 15% · 15.00 points/);
      } else {
        assert.match(await risk.innerText(), /Overall risk unavailable/);
        assert.match(await risk.innerText(), /Unknown · weight 15% · contribution unavailable/);
      }
      if (mode === 'partial') {
        const panel = page.getByRole('region', { name: 'Systemic Dependencies', exact: true });
        assert.match(await panel.innerText(), /Assessment incomplete/);
        assert.ok(!(await panel.innerText()).includes('No packages meet'));
        assert.match(await page.getByRole('region', { name: 'Systemic Dependency evidence', exact: true }).innerText(), /Classification unavailable/);
      }
    }
    assert.deepEqual(errors, []);
    console.log('PASS: Phase 5 numeric/unknown risk, component weights, systemic selection, incomplete coverage, stable graph, narrow UI, no browser errors');
  } finally {
    await context.close();
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
