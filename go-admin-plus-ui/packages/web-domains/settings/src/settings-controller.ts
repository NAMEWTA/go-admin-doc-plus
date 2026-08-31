import { SettingsRequestError, emptyDictionary, emptyItem, emptySetting, settingsPermissions, validKey, validSearch, validText,
  type Dictionary, type DictionaryInput, type DictionaryItem, type DictionaryItemInput, type Setting, type SettingCategory, type SettingInput,
  type DictionaryOption, type SettingsClient, type SettingsFailure, type SettingsPermissionCode } from '@go-admin-plus/domain-settings'
import { createListController, type ListController } from '@go-admin-plus/ui'

export interface SearchFilters { readonly search:string }
export interface SettingsCapabilityPort { can(code:SettingsPermissionCode):boolean; scope():string|null }
export type MutationResult='completed'|'invalid'|'cancelled'|'empty'|'busy'|'failed'|'refresh-failed'
type Projection='settings'|'dictionaries'|'items'
export type SettingsRemovalKind='setting'|'dictionary'|'dictionary-item'
export interface SettingsController {
  readonly settings:ListController<SearchFilters,Setting,string>
  readonly dictionaries:ListController<SearchFilters,Dictionary,string>
  readonly items:ListController<SearchFilters,DictionaryItem,string>
  readonly busy:boolean;readonly pendingRepair:boolean
  failure():SettingsFailure|null;failureTraceId():string|null;can(code:SettingsPermissionCode):boolean;selectCategory(value:SettingCategory):Promise<void>;selectDictionary(value:Dictionary|null):Promise<void>
  category():SettingCategory;dictionary():Dictionary|null
  saveSetting(value:SettingInput&{id?:string;revision?:number}):Promise<MutationResult>;removeSetting(value:Setting):Promise<MutationResult>
  saveDictionary(value:DictionaryInput&{id?:string;revision?:number}):Promise<MutationResult>;removeDictionary(value:Dictionary):Promise<MutationResult>
  saveItem(value:DictionaryItemInput&{id?:string;revision?:number}):Promise<MutationResult>;removeItem(value:DictionaryItem):Promise<MutationResult>
  options(key:string):Promise<ReadonlyArray<DictionaryOption>>
  repairProjection():Promise<MutationResult>;emptySetting():SettingInput;emptyDictionary():DictionaryInput;emptyItem():DictionaryItemInput
}

export const createSettingsController=(client:SettingsClient,confirm:(kind:SettingsRemovalKind)=>Promise<boolean>,capability:SettingsCapabilityPort):SettingsController=>{
  let currentCategory:SettingCategory='business',requestedCategory:SettingCategory=currentCategory,currentDictionary:Dictionary|null=null,requestedDictionary:Dictionary|null=currentDictionary,failure:SettingsFailure|null=null,failureTraceId:string|null=null,busy=false,selectionBusy=false,pending:Projection|null=null,categorySelection=0,dictionarySelection=0
  const visible=(permission:SettingsPermissionCode)=>{
    if(capability.scope()!=='all'||!capability.can(permission))return false
    if(permission===settingsPermissions.valuesWrite||permission===settingsPermissions.valuesDelete)return capability.can(settingsPermissions.valuesRead)
    if(permission===settingsPermissions.dictionariesWrite||permission===settingsPermissions.dictionariesDelete||permission===settingsPermissions.optionsRead)return capability.can(settingsPermissions.dictionariesRead)
    return true
  }
  const clearFailure=()=>{failure=null;failureTraceId=null}
  const record=(error:unknown)=>{failure=error instanceof SettingsRequestError?error.category:'unavailable';failureTraceId=error instanceof SettingsRequestError?error.traceId??null:null}
  const makeList=<Row>(permission:SettingsPermissionCode,rowKey:(row:Row)=>string,load:(request:{filters:SearchFilters;page:number;pageSize:number})=>Promise<{rows:Row[];total:number}>):ListController<SearchFilters,Row,string>=>{
    const raw=createListController<SearchFilters,Row,string>({initialFilters:()=>({search:''}),normalizeFilters:value=>({search:value.search.trim()}),validate:request=>{if(!validSearch(request.filters.search)){failure='validation';failureTraceId=null;throw new SettingsRequestError('validation')}},rowKey,load:async request=>{clearFailure();if(!visible(permission)){failure='forbidden';throw new SettingsRequestError('forbidden')}try{return await load(request)}catch(error){record(error);throw error}}})
    return {snapshot(){const value=raw.snapshot();return visible(permission)&&failure!== 'relogin'&&failure!=='forbidden'&&failure!=='unavailable'?value:{...value,rows:[],total:0,selectedKeys:[]}},refresh:()=>raw.refresh(),search:value=>raw.search(value),reset:()=>raw.reset(),setPage:value=>raw.setPage(value),setPageSize:value=>raw.setPageSize(value),setSort:value=>raw.setSort(value),select:rows=>{if(visible(permission))raw.select(rows)},clearSelection:()=>raw.clearSelection()}
  }
  const settings=makeList<Setting>(settingsPermissions.valuesRead,row=>row.id,request=>client.listSettings(requestedCategory,{search:request.filters.search,page:request.page,pageSize:request.pageSize}))
  const dictionaries=makeList<Dictionary>(settingsPermissions.dictionariesRead,row=>row.id,request=>client.listDictionaries({search:request.filters.search,page:request.page,pageSize:request.pageSize}))
  const items=makeList<DictionaryItem>(settingsPermissions.dictionariesRead,row=>row.id,request=>requestedDictionary?client.listItems(requestedDictionary.id,{search:request.filters.search,page:request.page,pageSize:request.pageSize}):Promise.resolve({rows:[],total:0}))
  const projection=(value:Projection)=>value==='settings'?settings:value==='dictionaries'?dictionaries:items
  const refresh=async(value:Projection):Promise<MutationResult>=>{try{await projection(value).refresh();if(value==='dictionaries'&&currentDictionary){currentDictionary=dictionaries.snapshot().rows.find(row=>row.id===currentDictionary?.id)??null;if(!currentDictionary)items.clearSelection()}pending=null;clearFailure();return'completed'}catch(error){record(error);return'refresh-failed'}}
  const mutate=async(target:Projection,permission:SettingsPermissionCode,operation:()=>Promise<void>,confirmation?:SettingsRemovalKind):Promise<MutationResult>=>{if(busy||selectionBusy)return'busy';if(pending)return'refresh-failed';if(!visible(permission)){failure='forbidden';failureTraceId=null;return'failed'};busy=true;clearFailure();try{if(confirmation&&!await confirm(confirmation))return'cancelled';if(!visible(permission)){failure='forbidden';failureTraceId=null;return'failed'};try{await operation()}catch(error){record(error);return'failed'};pending=target;projection(target).clearSelection();return await refresh(target)}finally{busy=false}}
  const beginSelection=()=>{if(busy||selectionBusy){failure='unavailable';failureTraceId=null;throw new SettingsRequestError('unavailable')}selectionBusy=true}
  return {settings,dictionaries,items,get busy(){return busy||selectionBusy},get pendingRepair(){return pending!==null},failure:()=>failure,failureTraceId:()=>failureTraceId,can:visible,category:()=>currentCategory,dictionary:()=>currentDictionary,
    async selectCategory(value){beginSelection();const selection=++categorySelection,previous=currentCategory;requestedCategory=value;try{await settings.reset();if(selection===categorySelection)currentCategory=value}catch(error){if(selection===categorySelection){currentCategory=previous;requestedCategory=previous}throw error}finally{if(selection===categorySelection)selectionBusy=false}},
    async selectDictionary(value){beginSelection();const selection=++dictionarySelection,previous=currentDictionary;requestedDictionary=value;try{await items.reset();if(selection===dictionarySelection)currentDictionary=value}catch(error){if(selection===dictionarySelection){currentDictionary=previous;requestedDictionary=previous}throw error}finally{if(selection===dictionarySelection)selectionBusy=false}},
    saveSetting(value){if(!validKey(value.key)||!validText(value.label,1,120)||!validText(value.value,1,500)||!validText(value.description,0,500)||value.id!==undefined&&(!Number.isSafeInteger(value.revision)||(value.revision??0)<1))return Promise.resolve('invalid');const input={category:value.category,key:value.key.trim().toLowerCase(),label:value.label.trim(),value:value.value.trim(),description:value.description.trim(),enabled:value.enabled};return mutate('settings',settingsPermissions.valuesWrite,async()=>{if(value.id)await client.updateSetting(value.id,{...input,revision:value.revision!});else await client.createSetting(input)})},
    removeSetting(value){return mutate('settings',settingsPermissions.valuesDelete,()=>client.deleteSetting(value.id,value.revision),'setting')},
    saveDictionary(value){if(!validKey(value.key)||!validText(value.name,1,120)||!validText(value.description,0,500)||value.id!==undefined&&(!Number.isSafeInteger(value.revision)||(value.revision??0)<1))return Promise.resolve('invalid');const input={key:value.key.trim().toLowerCase(),name:value.name.trim(),description:value.description.trim(),enabled:value.enabled};return mutate('dictionaries',settingsPermissions.dictionariesWrite,async()=>{if(value.id)await client.updateDictionary(value.id,{...input,revision:value.revision!});else await client.createDictionary(input)})},
    removeDictionary(value){return mutate('dictionaries',settingsPermissions.dictionariesDelete,()=>client.deleteDictionary(value.id,value.revision),'dictionary')},
    saveItem(value){if(!currentDictionary||!validText(value.value,1,120)||!validText(value.label,1,120)||!Number.isSafeInteger(value.sortOrder)||value.sortOrder<0||value.sortOrder>100000||value.id!==undefined&&(!Number.isSafeInteger(value.revision)||(value.revision??0)<1))return Promise.resolve('invalid');const input={value:value.value.trim(),label:value.label.trim(),sortOrder:value.sortOrder,enabled:value.enabled};return mutate('items',settingsPermissions.dictionariesWrite,async()=>{if(value.id)await client.updateItem(value.id,{...input,revision:value.revision!});else await client.createItem(currentDictionary!.id,input)})},
    removeItem(value){return mutate('items',settingsPermissions.dictionariesDelete,()=>client.deleteItem(value.id,value.revision),'dictionary-item')},
    async options(key){if(!visible(settingsPermissions.optionsRead)){failure='forbidden';failureTraceId=null;return[]};try{const value=await client.options(key);clearFailure();return value}catch(error){record(error);return[]}},
    async repairProjection(){if(busy)return'busy';if(!pending)return'completed';busy=true;try{return await refresh(pending)}finally{busy=false}},
    emptySetting:()=>emptySetting(currentCategory),emptyDictionary,emptyItem,
  }
}
