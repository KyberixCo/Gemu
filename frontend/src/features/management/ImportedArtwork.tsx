import {useLocalization} from '../../localization';
import {useState,type ReactNode} from 'react';
import type {Entry,LibraryArtwork} from '../../library';

export function ImportedArtwork({entry,children}:{entry?:Entry;children:ReactNode}){
 const {t}=useLocalization();
 const [page,setPage]=useState('Galería');
 const artworks:LibraryArtwork[]=[...(entry?.metadata.artworks??[])];
 for(const kind of ['cover','background'] as const){
  const url=entry?.metadata[kind];
  if(url&&!artworks.some(art=>art.url===url))artworks.push({id:kind,kind,url,provider:entry?.metadata.origin??'Steam',width:0,height:0});
 }
 const labels={background:'Fondo',cover:'Portada',logo:'Logotipo',icon:'Icono'};
 return <><div className="tv-subnav">{['Galería','Rutas propias'].map(name=><button data-nav key={name} aria-pressed={page===name} onClick={()=>setPage(name)}>{t(name)}</button>)}</div>
 {page==='Rutas propias'?children:<><p className="manager-hint">{t("El primer fondo válido de la prioridad más alta es el principal. La portada elegida también participa en la rotación si es horizontal y tiene al menos 1280 × 720 píxeles. Los logotipos e iconos se guardan por separado.")}</p>
 <div className="imported-artwork" tabIndex={0} data-nav aria-label={t("Arte guardado")}>{artworks.map(art=><figure key={art.id}>
  <img src={art.url} alt={t(labels[art.kind])} loading="lazy"/>
  <figcaption><strong>{t(labels[art.kind])}</strong><span>{art.provider}{art.width>0?` · ${art.width} × ${art.height}`:''}</span>{art.author&&<span>{art.author}</span>}{art.page&&<a data-nav href={art.page} target="_blank" rel="noreferrer">{t("Ver procedencia")}</a>}</figcaption>
 </figure>)}</div>{!artworks.length&&<p>{t("No hay arte importado. Configura un proveedor en Información y arte y actualiza la biblioteca.")}</p>}</>}
 </>;
}
