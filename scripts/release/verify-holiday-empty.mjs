const p=(await taskSpace(Number(process.env.EGO_TASK_SPACE))).page('p1');
if (!(await p.url()).startsWith('http://127.0.0.1:')) throw new Error('Local preview required');
await p.waitForFunction(()=>document.styleSheets.length>0);
const result=await p.evaluate(()=>{
 document.documentElement.dataset.themePreset='iceberg';
 const fixture=document.createElement('div');fixture.dataset.slot='empty-header';fixture.style.cssText='display:flex;flex-direction:column;align-items:center;padding:32px;background:var(--background)';
 fixture.innerHTML='<span class="holo holo-empty" aria-hidden="true"></span><p>暂无数据</p>';
 document.body.replaceChildren(fixture);
 const slot=fixture.firstElementChild;
 const rules=[...document.styleSheets].flatMap(s=>{try{return [...s.cssRules]}catch{return []}});
 const fix=rules.find(r=>r.selectorText?.includes('body[data-holiday]')&&r.selectorText?.includes('empty-header')&&r.selectorText?.endsWith('::before'));
 if(!fix)throw new Error('Missing regression fix');
 const original=fix.style.cssText;
 document.body.dataset.holiday='mid-autumn';
 fix.style.cssText='';
 const before=getComputedStyle(fixture,'::before').content!=='none'&&getComputedStyle(slot).display!=='none';
 fix.style.cssText=original;
 const checks=[];
 for(const holiday of ['mid-autumn','national-day','new-year','spring-festival','dragon-boat','']){
  if(holiday)document.body.dataset.holiday=holiday;else delete document.body.dataset.holiday;
  for(const dark of [false,true]){
   document.documentElement.classList.toggle('dark',dark);
   checks.push({holiday,dark,oldVisible:getComputedStyle(fixture,'::before').content!=='none',newVisible:getComputedStyle(slot).display!=='none'});
  }
 }
 document.body.dataset.holiday='mid-autumn';document.documentElement.classList.remove('dark');
 return {reproducedDuplicateWithoutFix:before,checks};
});
if(!result.reproducedDuplicateWithoutFix||result.checks.some(x=>x.holiday?x.oldVisible||!x.newVisible:!x.oldVisible||x.newVisible))throw new Error(JSON.stringify(result));
console.log(result);
