import type { components } from './generated/client'

export type Setting = components['schemas']['SettingValue']
export type SettingInput = components['schemas']['SettingValueInput']
export type SettingPage = components['schemas']['SettingValuePage']
export type SettingCategory = components['schemas']['SettingCategory']
export type Dictionary = components['schemas']['DictionaryType']
export type DictionaryInput = components['schemas']['DictionaryTypeInput']
export type DictionaryPage = components['schemas']['DictionaryTypePage']
export type DictionaryItem = components['schemas']['DictionaryItem']
export type DictionaryItemInput = components['schemas']['DictionaryItemInput']
export type DictionaryItemPage = components['schemas']['DictionaryItemPage']
export type DictionaryOption = components['schemas']['DictionaryOption']
export type SettingsFailure = 'relogin'|'forbidden'|'validation'|'conflict'|'not-found'|'unavailable'
export type SettingsPermissionCode = typeof settingsPermissions[keyof typeof settingsPermissions]
export const settingsPermissions = {
  valuesRead:'settings.values.read', valuesWrite:'settings.values.write', valuesDelete:'settings.values.delete',
  dictionariesRead:'settings.dictionaries.read', dictionariesWrite:'settings.dictionaries.write', dictionariesDelete:'settings.dictionaries.delete', optionsRead:'settings.options.read',
} as const
export interface PageQuery { readonly search:string; readonly page:number; readonly pageSize:number }
export interface SettingsClient {
  listSettings(category:SettingCategory,query:PageQuery):Promise<SettingPage>
  createSetting(input:SettingInput):Promise<Setting>
  updateSetting(id:string,input:SettingInput&{revision:number}):Promise<Setting>
  deleteSetting(id:string,revision:number):Promise<void>
  listDictionaries(query:PageQuery):Promise<DictionaryPage>
  createDictionary(input:DictionaryInput):Promise<Dictionary>
  updateDictionary(id:string,input:DictionaryInput&{revision:number}):Promise<Dictionary>
  deleteDictionary(id:string,revision:number):Promise<void>
  listItems(dictionaryId:string,query:PageQuery):Promise<DictionaryItemPage>
  createItem(dictionaryId:string,input:DictionaryItemInput):Promise<DictionaryItem>
  updateItem(id:string,input:DictionaryItemInput&{revision:number}):Promise<DictionaryItem>
  deleteItem(id:string,revision:number):Promise<void>
  options(dictionaryKey:string):Promise<ReadonlyArray<DictionaryOption>>
}
export class SettingsRequestError extends Error { readonly category:SettingsFailure;readonly traceId?:string;constructor(category:SettingsFailure,traceId:string|null=null){super(category);this.category=category;if(traceId!==null)this.traceId=traceId} }
export const codePointLength=(value:string)=>Array.from(value).length
export const validSearch=(value:string)=>codePointLength(value.trim())<=100
export const validKey=(value:string)=>/^[a-z0-9][a-z0-9_.-]{2,79}$/.test(value.trim().toLowerCase())
export const validText=(value:string,min:number,max:number)=>{const length=codePointLength(value.trim());return length>=min&&length<=max&&!/\p{Cc}/u.test(value)}
export const emptySetting=(category:SettingCategory='business'):SettingInput=>({category,key:'',label:'',value:'',description:'',enabled:true})
export const emptyDictionary=():DictionaryInput=>({key:'',name:'',description:'',enabled:true})
export const emptyItem=():DictionaryItemInput=>({value:'',label:'',sortOrder:0,enabled:true})
