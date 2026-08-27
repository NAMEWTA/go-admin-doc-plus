import { describe,expect,it } from 'vitest'
import { codePointLength,validKey,validSearch,validText } from './settings'
describe('settings domain validation',()=>{
  it('uses Unicode code points at locked boundaries',()=>{expect(codePointLength('😀')).toBe(1);expect(validText('😀'.repeat(120),1,120)).toBe(true);expect(validText('😀'.repeat(121),1,120)).toBe(false);expect(validText('界'.repeat(500),1,500)).toBe(true);expect(validText('界'.repeat(501),1,500)).toBe(false);expect(validSearch('😀'.repeat(100))).toBe(true);expect(validSearch('😀'.repeat(101))).toBe(false)})
  it('locks stable keys and control-free display values',()=>{expect(validKey('order.status')).toBe(true);expect(validKey('Order status')).toBe(false);expect(validText('unsafe\ntext',1,120)).toBe(false)})
})
