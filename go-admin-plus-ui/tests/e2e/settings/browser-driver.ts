import { createApp,h,type App,type Component } from 'vue'
import { createSettingsController,createWebSettingsClient,SettingsPage } from '@go-admin/web-domain-settings'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'
import { SettingsRequestError,settingsPermissions,type SettingsPermissionCode } from '@go-admin/domain-settings'

const result=document.querySelector<HTMLElement>('#result')!
const session=createSessionController(createWebSessionClient(fetch,'/api'))
const capabilities=createCapabilityController(createWebAdministrationClient(fetch,'/api'))
const client=createWebSettingsClient(fetch,'/api')
const controller=createSettingsController(client,async()=>true,{can:(code:SettingsPermissionCode)=>capabilities.can(code),scope:()=>capabilities.state().manifest?.dataScope??null})
let app:App<Element>|null=null
const mount=()=>{app?.unmount();document.querySelector('#app')!.textContent='';app=createApp({render:()=>h(SettingsPage as Component,{controller})});app.mount('#app')}
const delay=(ms:number)=>new Promise(resolve=>setTimeout(resolve,ms))
const wait=async(condition:()=>boolean,message:string)=>{const until=Date.now()+15000;while(Date.now()<until){if(condition())return;await delay(25)}throw new Error(message)}
const control=async(path:string,body?:unknown)=>{const response=await fetch(path,{method:'POST',headers:body===undefined?undefined:{'Content-Type':'application/json'},body:body===undefined?undefined:JSON.stringify(body)});if(response.status!==204)throw new Error('test control failed')}
const element=<T extends Element>(selector:string)=>{const value=document.querySelector<T>(selector);if(!value)throw new Error('expected interface element missing');return value}
const input=(name:string,value:string)=>{const field=element<HTMLInputElement|HTMLTextAreaElement>(`[name="${name}"]`);field.value=value;field.dispatchEvent(new Event('input',{bubbles:true}))}
const click=(testId:string)=>element<HTMLButtonElement>(`[data-testid="${testId}"]`).click()
const createSetting=async(key:string,label:string,value:string)=>{input('setting-key',key);input('setting-label',label);input('setting-value',value);click('save-setting');await wait(()=>controller.settings.snapshot().rows.some(row=>row.key===key),`setting create failed`)}

try{
  await session.login({username:'admin',password:'administrator password'})
  if(session.state().status!=='authenticated'||document.cookie.includes('__Host-go-admin-session'))throw new Error('login contract failed')
  await capabilities.refresh();for(const code of Object.values(settingsPermissions))if(!capabilities.can(code))throw new Error('settings capability missing')
  mount();await wait(()=>!controller.settings.snapshot().loading&&document.querySelector('[data-testid="setting-form"]')!==null,'settings page did not load')
  await createSetting('shop.title','商店标题','Public title')
  click('edit-setting-shop.title');await wait(()=>element<HTMLInputElement>('[name="setting-key"]').value==='shop.title','setting edit did not open');input('setting-value','Updated title');click('save-setting');await wait(()=>controller.settings.snapshot().rows.some(row=>row.key==='shop.title'&&row.value==='Updated title'),'setting update failed')
  const current=controller.settings.snapshot().rows.find(row=>row.key==='shop.title')!;try{await client.updateSetting(current.id,{category:current.category,key:current.key,label:current.label,value:'stale',description:current.description,enabled:current.enabled,revision:current.revision-1});throw new Error('stale revision succeeded')}catch(error){if(!(error instanceof SettingsRequestError)||error.category!=='conflict')throw error}
  for(const fixture of [{key:'literal.percent',label:'% literal'},{key:'literal.under',label:'_ literal'},{key:'literal.unicode',label:'ä Unicode'}])await client.createSetting({category:'business',key:fixture.key,label:fixture.label,value:'visible',description:'',enabled:true})
  await controller.settings.refresh();for(const search of ['%','_','ä']){input('setting-search',search);element<HTMLFormElement>('[name="setting-search"]').closest('form')!.requestSubmit();await wait(()=>controller.settings.snapshot().total===1&&controller.settings.snapshot().rows[0]?.label.includes(search)===true,'literal UI search failed')}
  click('tab-ui');await wait(()=>controller.category()==='ui','UI tab failed');await createSetting('ui.banner','Interface banner','Welcome')
  click('tab-business');await wait(()=>controller.category()==='business','business tab failed');await control('/__test/restart');await controller.settings.refresh();if(!controller.settings.snapshot().rows.some(row=>row.key==='shop.title'))throw new Error('restart persistence failed')
  click('tab-dictionaries');await wait(()=>document.querySelector('[data-testid="dictionary-form"]')!==null,'dictionary tab failed');input('dictionary-key','order.status');input('dictionary-name','订单状态');click('save-dictionary');await wait(()=>controller.dictionaries.snapshot().rows.some(row=>row.key==='order.status'),'dictionary create failed');click('select-dictionary-order.status');await wait(()=>document.querySelector('[data-testid="item-form"]')!==null,'dictionary selection failed')
  input('item-value','paid');input('item-label','Paid');input('item-order','20');click('save-item');await wait(()=>controller.items.snapshot().rows.some(row=>row.value==='paid'),'item create failed');input('item-value','draft');input('item-label','Draft');input('item-order','10');click('save-item');await wait(()=>controller.items.snapshot().rows.length===2,'item ordering create failed');if(controller.items.snapshot().rows.map(row=>row.value).join(',')!=='draft,paid')throw new Error('item deterministic order failed')
  click('edit-item-paid');await wait(()=>element<HTMLInputElement>('[name="item-value"]').value==='paid','item edit failed');input('item-label','Paid updated');click('save-item');await wait(()=>controller.items.snapshot().rows.some(row=>row.label==='Paid updated'),'item update failed')
  click('delete-dictionary-order.status');await wait(()=>controller.failure()==='conflict','dictionary reference conflict missing')
  const sensitive='Bearer abcdefghijklmnop';const sensitiveResponse=await fetch('/api/settings/values',{method:'POST',headers:{'Content-Type':'application/json','X-CSRF-Token':'invalid'},body:JSON.stringify({category:'business',key:'private-key',label:'Secret',value:sensitive,description:'',enabled:true})});if(sensitiveResponse.status!==403)throw new Error('csrf precedence failed');if((await sensitiveResponse.text()).includes(sensitive))throw new Error('sensitive response echoed material')
  await control('/__test/scope',{scope:'self'});await capabilities.refresh();if(controller.can(settingsPermissions.valuesWrite)||controller.settings.snapshot().rows.length)throw new Error('self scope did not fail closed');await control('/__test/scope',{scope:'all'});await control('/__test/permissions',{enabled:false});await capabilities.refresh();if(controller.can(settingsPermissions.valuesWrite))throw new Error('permission revoke not immediate');try{await client.createSetting({category:'business',key:'denied.key',label:'Denied',value:'Denied',description:'',enabled:true});throw new Error('revoked direct write succeeded')}catch(error){if(!(error instanceof SettingsRequestError)||error.category!=='forbidden')throw error}
  await control('/__test/permissions',{enabled:true});await capabilities.refresh();const baseline=(await client.listSettings('business',{search:'',page:1,pageSize:20})).total;const csrf=await fetch('/api/settings/values',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({category:'business',key:'csrf.key',label:'CSRF',value:'Rejected',description:'',enabled:true})});const problem=await csrf.json() as{code?:string};if(csrf.status!==403||problem.code!=='CSRF_REJECTED'||(await client.listSettings('business',{search:'',page:1,pageSize:20})).total!==baseline)throw new Error('csrf rejection contract failed')
  await control('/__test/revoke-session');try{await client.listSettings('business',{search:'',page:1,pageSize:20});throw new Error('revoked session remained active')}catch(error){if(!(error instanceof SettingsRequestError)||error.category!=='relogin')throw error}
  result.textContent='SETTINGS_E2E_PASS'
}catch(error){const message=error instanceof Error&&/^[a-zA-Z0-9 .,:|'-]{1,160}$/.test(error.message)?error.message:'browser assertion failed';result.textContent=`SETTINGS_E2E_FAIL|${message}`}
