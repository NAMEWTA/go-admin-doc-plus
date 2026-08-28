import { createContractClient, SettingsRequestError, type SettingsClient, type SettingsFailure } from '@go-admin-plus/domain-settings'

interface Problem { category?:string; code?:string }
const csrfPattern=/^[A-Za-z0-9_-]{43}$/
export const createWebSettingsClient=(fetcher:typeof fetch=fetch,baseUrl='/api'):SettingsClient=>{
  let csrf='';let classified:SettingsFailure|null=null;let tail=Promise.resolve()
  const serialized=<T>(operation:()=>Promise<T>):Promise<T>=>{const result=tail.then(operation,operation);tail=result.then(()=>undefined,()=>undefined);return result}
  const contract=createContractClient({baseUrl,fetch:async input=>{const headers=new Headers(input.headers);if(csrf&&input.method!=='GET')headers.set('X-CSRF-Token',csrf);const response=await fetcher(new Request(input,{credentials:'include',headers}));const next=response.headers.get('X-CSRF-Token');if(next!==null&&!csrfPattern.test(next)){csrf='';classified='relogin';throw new SettingsRequestError('relogin')}const body=response.status>=400?await response.clone().json().catch(()=>null) as Problem|null:null;classified=classify(response.status,body);if(next)csrf=next;else if(classified==='relogin')csrf='';return response}})
  const failure=(error:unknown):never=>{const category=classified??problemCategory(error);classified=null;throw new SettingsRequestError(category)}
  const required=<T>(data:T|undefined,error:unknown):T=>error===undefined&&data!==undefined?data:failure(error)
  const completed=(error:unknown)=>{if(error!==undefined)failure(error)}
  return {
    listSettings:(category,query)=>serialized(async()=>{const result=await contract.GET('/settings/values',{params:{query:{category,...query}}});return required(result.data,result.error)}),
    createSetting:body=>serialized(async()=>{const result=await contract.POST('/settings/values',{body});return required(result.data,result.error)}),
    updateSetting:(settingId,body)=>serialized(async()=>{const result=await contract.PATCH('/settings/values/{settingId}',{params:{path:{settingId}},body});return required(result.data,result.error)}),
    deleteSetting:(settingId,revision)=>serialized(async()=>{const result=await contract.DELETE('/settings/values/{settingId}',{params:{path:{settingId},query:{revision}}});completed(result.error)}),
    listDictionaries:query=>serialized(async()=>{const result=await contract.GET('/settings/dictionaries',{params:{query}});return required(result.data,result.error)}),
    createDictionary:body=>serialized(async()=>{const result=await contract.POST('/settings/dictionaries',{body});return required(result.data,result.error)}),
    updateDictionary:(dictionaryId,body)=>serialized(async()=>{const result=await contract.PATCH('/settings/dictionaries/{dictionaryId}',{params:{path:{dictionaryId}},body});return required(result.data,result.error)}),
    deleteDictionary:(dictionaryId,revision)=>serialized(async()=>{const result=await contract.DELETE('/settings/dictionaries/{dictionaryId}',{params:{path:{dictionaryId},query:{revision}}});completed(result.error)}),
    listItems:(dictionaryId,query)=>serialized(async()=>{const result=await contract.GET('/settings/dictionaries/{dictionaryId}/items',{params:{path:{dictionaryId},query}});return required(result.data,result.error)}),
    createItem:(dictionaryId,body)=>serialized(async()=>{const result=await contract.POST('/settings/dictionaries/{dictionaryId}/items',{params:{path:{dictionaryId}},body});return required(result.data,result.error)}),
    updateItem:(itemId,body)=>serialized(async()=>{const result=await contract.PATCH('/settings/dictionary-items/{itemId}',{params:{path:{itemId}},body});return required(result.data,result.error)}),
    deleteItem:(itemId,revision)=>serialized(async()=>{const result=await contract.DELETE('/settings/dictionary-items/{itemId}',{params:{path:{itemId},query:{revision}}});completed(result.error)}),
    options:dictionaryKey=>serialized(async()=>{const result=await contract.GET('/settings/options/{dictionaryKey}',{params:{path:{dictionaryKey}}});return required(result.data,result.error)}),
  }
}
const classify=(status:number,value:Problem|null):SettingsFailure|null=>status===401||value?.code==='CSRF_REJECTED'?'relogin':status===403?'forbidden':status===400||status===422?'validation':status===404?'not-found':status===409?'conflict':status>=500?'unavailable':null
const problemCategory=(value:unknown):SettingsFailure=>typeof value==='object'&&value!==null&&'category'in value?({authentication:'relogin',authorization:'forbidden',validation:'validation',not_found:'not-found',conflict:'conflict'} as const)[String((value as Problem).category) as 'authentication'|'authorization'|'validation'|'not_found'|'conflict']??'unavailable':'unavailable'
