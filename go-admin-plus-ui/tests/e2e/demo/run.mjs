import { spawn, spawnSync } from 'node:child_process'
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { request as httpsRequest } from 'node:https'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { activeChildren, assertChildHealthy, CDPClient, delay, spawnTracked, terminateChild, withTimeout } from '../iam/administration/runner-support.mjs'

const required = process.env.GO_ADMIN_REQUIRE_DEMO_E2E === '1'
if (!required) { console.log('DEMO_E2E_SKIP required opt-in is disabled'); process.exit(0) }
const chromium = process.env.GO_ADMIN_TEST_CHROMIUM_EXECUTABLE
const postgresKey = 'GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN'; const postgres = process.env[postgresKey]
if (!chromium || !postgres) { console.error('DEMO_E2E_RUN_FAIL|required environment is missing'); process.exit(1) }
const root = dirname(fileURLToPath(import.meta.url)); const uiRoot = resolve(root, '../../..'); const backend = resolve(uiRoot, '../go-admin-plus')
const temporary = mkdtempSync(join(tmpdir(), 'go-admin-demo-e2e-')); const staticRoot = join(temporary, 'static'); const deadline = Date.now()+5*60_000
const keys = ['APPDATA','CC','CGO_ENABLED','COMSPEC','COREPACK_HOME','CXX','GOCACHE','GOENV','GOMODCACHE','GONOPROXY','GONOSUMDB','GOPATH','GOPRIVATE','GOPROXY','GOROOT','GOSUMDB','GOTOOLCHAIN','HOME','LANG','LC_ALL','LOCALAPPDATA','NO_COLOR','PATH','PATHEXT','PNPM_HOME','SSL_CERT_DIR','SSL_CERT_FILE','SystemRoot','TEMP','TMP','TMPDIR','USERPROFILE','WINDIR','XDG_CACHE_HOME','XDG_CONFIG_HOME','XDG_RUNTIME_DIR']
const remaining = maximum => { const value = Math.min(maximum, deadline-Date.now()); if (value <= 0) throw new Error('overall deadline exceeded'); return value }
const environment = (extra={}, pg=false) => { const result={}; for (const key of keys) if (process.env[key] !== undefined) result[key]=process.env[key]; Object.assign(result, extra); delete result[postgresKey]; if (pg) result[postgresKey]=postgres; return result }
const checked = (command,args,options) => { const value=spawnSync(command,args,{...options,encoding:'utf8',timeout:remaining(120_000),killSignal:'SIGKILL'}); if(value.status!==0||value.error||value.signal) throw new Error(options.failure) }
const waitReady = async (path, child, profile) => { const until=Date.now()+remaining(60_000); while(Date.now()<until){ if(existsSync(path)) return readFileSync(path,'utf8').trim(); assertChildHealthy(child,`${profile} HTTPS host`); await delay(100) } throw new Error(`${profile} HTTPS host readiness timed out`) }
const devTools = child => withTimeout(new Promise((resolvePromise,reject)=>{ let buffered=''; child.stderr.setEncoding('utf8'); child.stderr.on('data',chunk=>{ buffered=(buffered+chunk).slice(-8192); const match=buffered.match(/DevTools listening on (ws:\/\/[^\s]+)/); if(match) resolvePromise(match[1]) }); child.once('error',()=>reject(new Error('Chromium could not start'))); child.once('exit',()=>reject(new Error('Chromium exited early'))) }),remaining(30_000),'Chromium readiness')
const websocket = url => { const socket=new WebSocket(url); return withTimeout(new Promise((resolvePromise,reject)=>{ socket.addEventListener('open',()=>resolvePromise(socket),{once:true}); socket.addEventListener('error',()=>reject(new Error('CDP connection failed')),{once:true}) }),remaining(15_000),'CDP connection',()=>socket.close()) }
const shutdown = baseURL => new Promise(resolvePromise=>{ if(!baseURL){resolvePromise();return} const request=httpsRequest(new URL('/__test/shutdown',baseURL),{method:'POST',rejectUnauthorized:false},response=>{response.resume();response.once('end',resolvePromise)}); request.once('error',resolvePromise); request.end() })
const runProfile = async profile => {
  const ready=join(temporary,profile,'ready'); let host,browser,socket,cdp,baseURL,operationError
  try {
    host=spawnTracked(spawn,'go',['test','./test/demo','-run','^TestDemoBrowserHarnessServer$','-count=1','-v'],{cwd:backend,env:environment({GO_ADMIN_DEMO_E2E_SERVE:'1',GO_ADMIN_DEMO_E2E_PROFILE:profile,GO_ADMIN_DEMO_E2E_READY_FILE:ready,GO_ADMIN_DEMO_E2E_STATIC_DIR:staticRoot},profile==='postgres'),stdio:['ignore','pipe','pipe'],drainStdout:true,drainStderr:true})
    baseURL=await waitReady(ready,host,profile); browser=spawnTracked(spawn,chromium,['--headless=new','--disable-gpu','--ignore-certificate-errors','--no-first-run','--remote-debugging-port=0',`--user-data-dir=${join(temporary,profile,'chromium')}`,baseURL],{env:environment(),stdio:['ignore','ignore','pipe']})
    socket=await websocket(await devTools(browser)); cdp=new CDPClient(socket,10_000); let target
    const targetDeadline=Date.now()+remaining(30_000); while(Date.now()<targetDeadline&&!target){ const values=await cdp.send('Target.getTargets'); target=values.targetInfos.find(value=>value.type==='page'&&value.url.startsWith(baseURL)); if(!target) await delay(100) }
    if(!target) throw new Error('browser page did not load'); const attached=await cdp.send('Target.attachToTarget',{targetId:target.targetId,flatten:true}); await cdp.send('Runtime.enable',{},attached.sessionId)
    const resultDeadline=Date.now()+remaining(120_000); while(Date.now()<resultDeadline){ const value=await cdp.send('Runtime.evaluate',{expression:"document.querySelector('#result')?.textContent ?? ''",returnByValue:true},attached.sessionId); const text=String(value.result.value??''); if(text.includes('DEMO_E2E_PASS')) return; if(text.includes('DEMO_E2E_FAIL|')) throw new Error(`${profile} browser scenario failed: ${text.split('|').slice(1).join('|').slice(0,180)}`); await delay(100) } throw new Error(`${profile} browser result timed out`)
  } catch(error){operationError=error} finally { const hostResult=await terminateChild(host,()=>shutdown(baseURL)); const browserResult=await terminateChild(browser,async()=>{if(cdp)await cdp.send('Browser.close');else socket?.close()}); try{socket?.close()}catch{} if(!hostResult.exited||!browserResult.exited)operationError??=new Error('test child cleanup failed') }
  if(operationError) throw operationError
}
let failure=''
try { checked('pnpm',['exec','vite','build','--config',join(root,'vite.config.ts')],{cwd:uiRoot,env:environment({GO_ADMIN_DEMO_E2E_OUT_DIR:staticRoot}),stdio:'pipe',failure:'browser fixture build failed'}); checked('go',['test','./test/demo','-run','^TestDemoBrowserHarnessServer$','-count=1'],{cwd:backend,env:environment({GO_ADMIN_DEMO_E2E_SERVE:'0'}),stdio:'pipe',failure:'HTTPS host compile self-check failed'}); await runProfile('sqlite'); await runProfile('postgres') } catch(error){ failure=error instanceof Error&&/^[a-zA-Z0-9 .,:|'-]{1,240}$/.test(error.message)?error.message:'demo runner failed' }
for(const child of activeChildren)await terminateChild(child); if(activeChildren.size===0)rmSync(temporary,{recursive:true,force:true})
if(failure){console.error(`DEMO_E2E_RUN_FAIL|${failure}`);process.exitCode=1}else console.log('DEMO_E2E_PASS profiles=sqlite,postgres')
