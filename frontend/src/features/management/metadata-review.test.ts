import {describe,it,expect} from 'vitest';
import {applyReviewedFields} from './MetadataReview';
import {defaultContentSettings,readContentSettings} from './ContentProviders';
const current={title:'Original',genre:'Aventura',description:'Texto original',cover:'/cover.png',background:'',logo:''};
describe('metadata approval',()=>{
 it('applies only approved fields and preserves artwork and rejected text',()=>{
  const result=applyReviewedFields(current,{...current,title:'Propuesta',description:'Nuevo texto',cover:'/other.png'},['description']);
  expect(result).toEqual({...current,description:'Nuevo texto'});
  expect(current.description).toBe('Texto original');
 });
 it('does not modify metadata without field approval',()=>{
  expect(applyReviewedFields(current,{...current,title:'Nuevo'},[])).toEqual(current);
 });
});
describe('content provider persistence',()=>{
 it('migrates old settings and rejects invalid priorities',()=>{
  expect(readContentSettings(undefined)).toEqual(defaultContentSettings());
  expect(readContentSettings({igdb:{enabled:true,priority:'99'}}).igdb).toEqual({enabled:false,priority:'1'});
 });
 it('restores preferences without retaining injected credentials',()=>{
  const result=readContentSettings({igdb:{enabled:true,priority:'2',secret:'private'},unknown:{enabled:true,priority:'1'}});
  expect(result.igdb).toEqual({enabled:true,priority:'2'});
  expect(JSON.stringify(result)).not.toContain('private');
  expect(Object.keys(result)).toHaveLength(4);
 });
});
