// e2e for spec 013 (013-AC14): a page open in instance A, then in B. Drives headless Chrome over CDP.
// Args: <port A, running on <work>/docs> <port B, started here> <cdp port> <work dir>
const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

const [portA, portB, cdpPort, work] = process.argv.slice(2);
if (!work) {
  console.error("usage: instance-routing.js <port A> <port B> <cdp port> <work dir>");
  process.exit(2);
}
const docs = path.join(work, "docs");
const sidecar = path.join(docs, "route.annotations.json5");
const pageA = `http://127.0.0.1:${portA}/route.md`;
const pageB = `http://127.0.0.1:${portB}/docs/route.md`;
const cdp = `http://127.0.0.1:${cdpPort}`;

let ws, seq = 0;
const waiting = new Map();
function send(method, params = {}) {
  if (ws.readyState !== WebSocket.OPEN) return Promise.reject(new Error("CDP socket closed"));
  const id = ++seq;
  ws.send(JSON.stringify({ id, method, params }));
  return new Promise((res, rej) => {
    const t = setTimeout(() => {
      waiting.delete(id);
      rej(new Error("CDP " + method + " got no reply within 10 s"));
    }, 10000);
    waiting.set(id, { res: (v) => (clearTimeout(t), res(v)), rej: (e) => (clearTimeout(t), rej(e)) });
  });
}
async function ev(expr) {
  const r = await send("Runtime.evaluate", { expression: expr, awaitPromise: true, returnByValue: true });
  if (r.exceptionDetails) throw new Error(r.exceptionDetails.exception?.description || r.exceptionDetails.text);
  return r.result.value;
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function waitFor(expr, ms = 4000) {
  const t0 = Date.now();
  let lastErr;
  for (;;) {
    try {
      if (await ev(expr)) return Date.now() - t0;
    } catch (e) {
      lastErr = e;
    }
    if (Date.now() - t0 > ms) throw new Error("timeout: " + expr + (lastErr ? " (last error: " + lastErr.message + ")" : ""));
    await sleep(200);
  }
}
let failures = 0;
const consoleErrors = [];
let completed = false;
process.on("beforeExit", () => {
  if (completed) return;
  console.error("ERROR run ended before completion");
  process.exit(3);
});
function check(label, ok, extra) {
  if (!ok) failures++;
  console.log((ok ? "PASS " : "FAIL ") + label + (extra !== undefined ? "  " + JSON.stringify(extra) : ""));
}
const sidecarJSON = () => JSON.parse(fs.readFileSync(sidecar, "utf8").replace(/^\/\/.*$/gm, ""));
const banner = () =>
  ev(`(() => { const b = document.querySelector(".annotations-banner"), s = b && b.querySelector("select");
    return { shown: !!b && !b.hidden, text: b ? b.textContent : null, via: s ? s.value : null, marker: window.__marker || null } })()`);
const labels = () => ev(`[...document.querySelectorAll(".annotations-list .annot-msg-reader small")].map((s) => s.textContent)`);

// registryFiles finds the registry file of pid. macOS keeps the user cache under HOME; Linux
// under XDG_CACHE_HOME, which run.sh sets in <work>/cache.
function registryFiles(pid) {
  return ["home", "cache"]
    .filter((d) => fs.existsSync(path.join(work, d)))
    .flatMap((d) => fs.readdirSync(path.join(work, d), { recursive: true }))
    .filter((f) => f.endsWith(`instances${path.sep}${pid}.json`));
}

// startB runs B on the docs directory, with the HOME of A, and resolves on its start line.
function startB() {
  const b = spawn(path.join(work, "gld"), ["-b=false", "-H", "127.0.0.1", "-p", portB, docs], { env: Object.assign({}, process.env, { HOME: path.join(work, "home") }) });
  fs.writeFileSync(path.join(work, "b.pid"), String(b.pid));
  let out = "";
  return new Promise((res, rej) => {
    const t = setTimeout(() => rej(new Error("B printed no start line: " + out)), 10000);
    b.stdout.on("data", (d) => {
      out += d;
      const m = /\(pid (\d+)\)/.exec(out);
      if (m) {
        clearTimeout(t);
        res({ proc: b, pid: Number(m[1]) });
      }
    });
    b.stderr.on("data", (d) => (out += d));
    b.on("exit", (code) => rej(new Error("B exited " + code + ": " + out)));
  });
}

(async () => {
  const list = await (await fetch(`${cdp}/json/list`)).json();
  ws = new WebSocket(list.find((t) => t.type === "page").webSocketDebuggerUrl);
  ws.onmessage = (m) => {
    const d = JSON.parse(m.data);
    if (d.method === "Runtime.consoleAPICalled" && d.params.type === "error") consoleErrors.push("console: " + d.params.args.map((x) => x.description || x.value).join(" ").slice(0, 200));
    if (d.method === "Runtime.exceptionThrown") consoleErrors.push("exception: " + (d.params.exceptionDetails.exception?.description || d.params.exceptionDetails.text).slice(0, 200));
    if (d.id && waiting.has(d.id)) {
      const w = waiting.get(d.id);
      waiting.delete(d.id);
      if (d.error) w.rej(new Error("CDP " + d.error.message));
      else w.res(d.result);
    }
  };
  ws.onclose = () => {
    for (const w of waiting.values()) w.rej(new Error("CDP socket closed"));
    waiting.clear();
  };
  await new Promise((r) => (ws.onopen = r));
  await send("Page.enable");
  await send("Runtime.enable");
  await send("Page.navigate", { url: pageA });
  await waitFor(`document.readyState === "complete" && !!document.querySelector(".annotations-banner")`);
  // Runtime.enable replays what the page logged before: the 010 run's forced error.
  consoleErrors.length = 0;
  await ev(`window.__marker = "a1"`);
  check("page open in A only: no banner", !(await banner()).shown, await banner());

  // B starts and opens the page in a second tab.
  const b = await startB();
  check("B's registry file written", registryFiles(b.pid).length === 1, registryFiles(b.pid));
  const tab = await (await fetch(`${cdp}/json/new?${encodeURIComponent(pageB)}`, { method: "PUT" })).json();
  const t0 = Date.now();
  const shown = await waitFor(`(() => { const b = document.querySelector(".annotations-banner"); return !b.hidden })()`, 30000).catch(() => -1);
  const bn = await banner();
  check("banner in A within 30 s, no reload (" + (Date.now() - t0) + " ms)", shown >= 0 && bn.marker === "a1" && bn.text.includes("Also open in :" + portB), bn);
  check("selector defaults to A", bn.via === portA, bn);

  // Send to B.
  await ev(`(() => { const s = document.querySelector(".annotations-banner select"); s.value = "${portB}"; s.dispatchEvent(new Event("change")) })()`);
  await ev(`document.getElementById("annotation-document").click()`);
  await ev(`document.querySelector(".annot-box textarea").value = "for B"`);
  await ev(`document.querySelector(".annot-save").click()`);
  await waitFor(`!document.querySelector(".annot-box").matches(":popover-open")`);
  const msg = sidecarJSON().annotations[0].thread[0];
  check("message stamped via B on disk", msg.via === Number(portB), msg);
  await waitFor(`document.querySelectorAll(".annotations-list .annot-msg-reader small").length === 1`);
  check("message shows → :B", (await labels())[0].endsWith("→ :" + portB), await labels());

  await send("Page.reload");
  await sleep(500);
  await waitFor(`document.readyState === "complete" && !document.querySelector(".annotations-banner").hidden`, 10000);
  check("choice survives a reload", (await banner()).via === portB, await banner());

  // B stops, its tab still open: A learns it from the registry, not from a refetch on focus.
  const visibility = await ev(`document.visibilityState`);
  b.proc.removeAllListeners("exit");
  b.proc.kill("SIGTERM");
  const t1 = Date.now();
  const gone = await waitFor(`document.querySelector(".annotations-banner").hidden`, 45000).catch(() => -1);
  check("banner gone within 45 s of B stopping (" + (Date.now() - t1) + " ms, A " + visibility + ")", gone >= 0, await banner());
  await fetch(`${cdp}/json/close/${tab.id}`);
  check("B's registry file removed", registryFiles(b.pid).length === 0, registryFiles(b.pid));
  await waitFor(`[...document.querySelectorAll(".annotations-list .annot-msg-reader small")].some((s) => s.textContent.endsWith("(gone)"))`, 5000).catch(() => {});
  check("message shows (gone)", (await labels())[0].endsWith("→ :" + portB + " (gone)"), await labels());

  check("no page exceptions / console errors", consoleErrors.length === 0, consoleErrors);
  console.log(failures ? failures + " FAILURE(S)" : "ALL PASS");
  completed = true;
  ws.close();
  process.exit(failures ? 1 : 0);
})().catch((e) => {
  console.error("ERROR", e);
  process.exit(2);
});
