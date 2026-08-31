import { describe,expect,it } from 'vitest'
import { SettingsRequestError } from '@go-admin-plus/domain-settings'
import { createWebSettingsClient } from './web-settings-client'
const json=(status:number,body:unknown,csrf?:string)=>new Response(JSON.stringify(body),{status,headers:{'Content-Type':'application/json',...(csrf?{'X-CSRF-Token':csrf}:{})}})
describe('web settings client',()=>{
  it('rejects malformed replacement CSRF without exposing it',async()=>{const client=createWebSettingsClient(async()=>json(200,{rows:[],total:0},'malformed'),'https://example.test/api');await expect(client.listDictionaries({search:'',page:1,pageSize:20})).rejects.toEqual(expect.objectContaining({category:'relogin'}))})
  it('keeps ordinary authorization distinct from relogin',async()=>{const client=createWebSettingsClient(async()=>json(403,{category:'authorization',code:'PERMISSION_DENIED'}),'https://example.test/api');await expect(client.listDictionaries({search:'',page:1,pageSize:20})).rejects.toEqual(expect.objectContaining({category:'forbidden'}));const csrf=createWebSettingsClient(async()=>json(403,{category:'authorization',code:'CSRF_REJECTED'}),'https://example.test/api');await expect(csrf.listDictionaries({search:'',page:1,pageSize:20})).rejects.toBeInstanceOf(SettingsRequestError)})
  it('preserves only a safe problem trace reference',async()=>{const client=createWebSettingsClient(async()=>json(409,{category:'conflict',traceId:'trace-settings-01234567',detail:'private row'}),'https://example.test/api');await expect(client.deleteSetting('setting-1',1)).rejects.toMatchObject({category:'conflict',traceId:'trace-settings-01234567'})})
})
