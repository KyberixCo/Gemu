import {useLocalization} from '../../localization';
import {useEffect,useRef,type ReactNode} from 'react';
export function WizardFrame({title,steps,current,children,actions}:{title:string;steps:readonly string[];current:number;children:ReactNode;actions:ReactNode}){
 const {t}=useLocalization();
 const root=useRef<HTMLDivElement>(null);
 useEffect(()=>{root.current?.querySelector<HTMLElement>('[data-nav]:not(:disabled)')?.focus({preventScroll:true});},[current]);
 return <div className="configuration-wizard" ref={root}>
 <header className="wizard-heading"><p className="eyebrow">{t(title)} {t("· PASO")} {current+1} {t("DE")} {steps.length}</p><h2>{t(steps[current])}</h2><ol aria-label={t("Pasos de configuración")}>{steps.map((step,index)=><li key={step} aria-current={index===current?'step':undefined} data-complete={index<current}>{t(step)}</li>)}</ol></header>
 <div className="wizard-content" key={steps[current]}>{children}</div>
 <footer className="wizard-actions">{actions}</footer></div>;
}
