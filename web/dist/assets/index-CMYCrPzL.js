(function(){const t=document.createElement("link").relList;if(t&&t.supports&&t.supports("modulepreload"))return;for(const i of document.querySelectorAll('link[rel="modulepreload"]'))n(i);new MutationObserver(i=>{for(const s of i)if(s.type==="childList")for(const o of s.addedNodes)o.tagName==="LINK"&&o.rel==="modulepreload"&&n(o)}).observe(document,{childList:!0,subtree:!0});function r(i){const s={};return i.integrity&&(s.integrity=i.integrity),i.referrerPolicy&&(s.referrerPolicy=i.referrerPolicy),i.crossOrigin==="use-credentials"?s.credentials="include":i.crossOrigin==="anonymous"?s.credentials="omit":s.credentials="same-origin",s}function n(i){if(i.ep)return;i.ep=!0;const s=r(i);fetch(i.href,s)}})();const g={vnode:null,unmount:null,diffed:null,event:null,debounceRendering:null,requestAnimationFrame:null,_diff:null,_render:null,_commit:null,_hook:null,_catchError:null,_root:null,_hydrationMismatch:null};let He=1;function z(e,t,...r){if(e==null&&(typeof t=="string"||typeof t=="number"||typeof t=="bigint"))return pe(null,t,null,null);const n={};let i=null,s=null;if(t!=null&&typeof t=="object")for(const o in t)o==="key"?i=t.key:o==="ref"?s=t.ref:n[o]=t[o];return r.length===1?n.children=r[0]:r.length>1&&(n.children=r),pe(e,n,i,s)}const _=z;function pe(e,t,r,n){const i={type:e,props:t,key:r,ref:n,_children:null,_parent:null,_depth:0,_dom:null,_component:null,constructor:void 0,_original:He++,_index:0,_flags:0};return g.vnode&&g.vnode(i),i}function re(e){return e.children}const U=1,Je=4,M=1,ne=2;let q=[],ee=!1,B,Se;function $e(e){B=e.patch,Se=e.applyRef}class Ue{constructor(t,r){this.props=t,this.context=r,this.state={},this._nextState=null,this._vnode=null,this._parentDom=null,this._bits=0,this._renderCallbacks=[],this._stateCallbacks=[],this.__hooks=null,this._depth=0}setState(t,r){const n=this._nextState==null?this._nextState={...this.state}:this._nextState,i=typeof t=="function"?t(n,this.props):t;i&&Object.assign(n,i),r&&this._stateCallbacks.push(r),this._vnode&&W(this)}forceUpdate(t){this._bits|=ne,t&&this._renderCallbacks.push(t),this._vnode&&W(this)}update(t,r){this.setState(t,r)}render(){return null}}function qe(e,t){return this.constructor(e,t)}function Be(e,t,r){if(e.prototype&&typeof e.prototype.render=="function")return new e(t,r);const n=new Ue(t,r);return n.constructor=e,n.render=qe,n}function W(e){e._bits&M||(e._bits|=M,q.push(e),ee||(ee=!0,(g.debounceRendering||queueMicrotask)(We)))}function We(){ee=!1,q.sort((t,r)=>t._depth-r._depth);const e=q;q=[];for(let t=0;t<e.length;t++){const r=e[t];r._bits&M&&r._vnode!=null&&(r._bits&=~M,Ke(r))}}function Ke(e){const t=e._vnode,r=e._parentDom,n=z(t.type,{...t.props,key:t.key,ref:t.ref});n._original=t._original;const i=[],s=[];B(r,n,t,e.context||{},Ve(r),null,i,null,!1,s),Ge(t,n,r),Ae(n,i,s)}function Ge(e,t,r){const n=e._parent;if(n&&n._children){const i=n._children.indexOf(e);i!==-1&&(n._children[i]=t),t._parent=n}r&&r.__root&&Ce(r.__root.vnode,e,t)}function Ce(e,t,r){if(e===t)return r;if(!e||!e._children)return e;for(let n=0;n<e._children.length;n++){const i=e._children[n];i===t?e._children[n]=r:Ce(i,t,r)}return e}function Ve(e){return!!(e&&e.ownerSVGElement!==void 0)}function Ae(e,t,r){for(let n=0;n<r.length;n++)Se(r[n],r[++n],r[++n]);g._commit&&g._commit(e,t);for(let n=0;n<t.length;n++){const i=t[n],s=i._renderCallbacks;i._renderCallbacks=[];for(let c=0;c<s.length;c++)s[c].call(i);const o=i._stateCallbacks;i._stateCallbacks=[];for(let c=0;c<o.length;c++)o[c].call(i)}}function he(e,t){return Re(e,t,!1)}function Ye(e,t){return Re(e,t,!0)}function Re(e,t,r){if(!B)throw new Error("Strike patch not bound");const n=t.__root?t.__root.vnode:null,i=r||!n?Array.from(t.childNodes):null,s=z(re,null,e);r&&(s._flags|=Je);const o=[],c=[];if(B(t,s,n,{},!1,i,o,null,r,c),i)for(let p=0;p<i.length;p++){const a=i[p];a.parentNode===t&&a.remove()}return t.__root={vnode:s,hydrating:!1},Ae(s,o,c),s._dom}const Qe="http://www.w3.org/2000/svg",Xe="http://www.w3.org/1998/Math/MathML",Ze={};function Te(e,t,r,n,i,s,o,c,p,a){if(t==null||t.constructor!==void 0)return null;g._diff&&g._diff(t);try{typeof t.type=="function"?tt(e,t,r,n,i,s,o,p,a):t.type==null?et(t,r,s,p):nt(e,t,r,n,i,s,o,p,a)}catch(h){oe(h,t,r)}return g.diffed&&g.diffed(t),t._dom}function et(e,t,r,n){const i=""+e.props;let s=t&&t.type===null?t._dom:null;if(s==null||s.nodeType!==3){if(r){for(let o=0;o<r.length;o++)if(r[o].nodeType===3){s=r.splice(o,1)[0];break}}s==null&&(s=document.createTextNode(i))}s.data!==i&&(s.data=i),e._dom=s}function tt(e,t,r,n,i,s,o,c,p){if(t.type===re||t.props&&t.props._parentDom)return rt(e,t,r,n,i,s,o,c,p);let a=r&&r.type===t.type?r._component:null;const h=!!a;if(t.type.contextType!=null){const x=n[t.type.contextType._id];t._contextValue=x?x.props.value:t.type.contextType._defaultValue}h?(t._component=a,a._vnode=t,a.props=t.props):(r&&se(r,null,!0),a=Be(t.type,t.props,n),t._component=a,a._vnode=t,a._parentDom=e,a._depth=(t._depth||0)+1),a.context=n;let b=a._nextState!=null?a._nextState:a.state;if(a._nextState=null,!(a._bits&ne)&&h&&a.shouldComponentUpdate&&!a.shouldComponentUpdate(t.props,b,n)){a.state=b,t._dom=r._dom,t._children=r._children,a._bits&=~M,st(t);return}a.state=b,a._bits&=-4,g._render&&g._render(t);let v,k=0;do a._bits&=~M,g._render&&g._render(t),v=a.render(a.props,t.type.contextType!=null?t._contextValue:n);while(a._bits&M&&++k<25);a.getChildContext&&(n=Object.assign({},n,a.getChildContext()));const y=t.props&&t.props._parentDom;if(ie(y||e,t,r,v,n,i,s,o,c,p),t._dom=y?null:K(t._children),o.push(a),!h)a.componentDidMount&&a._renderCallbacks.push(()=>a.componentDidMount());else if(a.componentDidUpdate){const x=r._component&&r._component.state;a._renderCallbacks.push(()=>a.componentDidUpdate(r.props,x))}t.ref&&p.push(t.ref,a,t)}function rt(e,t,r,n,i,s,o,c,p){const a=t.props&&t.props._parentDom||e;ie(a,t,r,t.props.children,n,i,s,o,c,p),t._dom=t.props&&t.props._parentDom?null:K(t._children)}function nt(e,t,r,n,i,s,o,c,p){const a=t.type;a==="svg"&&(i=!0),a==="foreignObject"&&(i=!1);let h=r&&r.type===a?r._dom:null;const b=r&&r.type===a?r.props:Ze,v=t.props;if(h==null&&(s&&(h=it(s,a)),h==null)){const k=i?Qe:a==="math"?Xe:null;h=k?document.createElementNS(k,a):document.createElement(a),c=!1,s=null}if(t._dom=h,v.dangerouslySetInnerHTML!=null){const k=v.dangerouslySetInnerHTML.__html,y=b.dangerouslySetInnerHTML&&b.dangerouslySetInnerHTML.__html;!c&&k!==y&&(h.innerHTML=k||"")}else{const k=s!=null||c?Array.from(h.childNodes):null;ie(h,t,r&&r.type===a?r:null,v.children,n,i,k,o,c,p)}dt(h,v,b,i,c),t.ref&&p.push(t.ref,h,t)}function it(e,t){for(let r=0;r<e.length;r++){const n=e[r],i=n.localName||n.nodeName;if(typeof t=="string"&&i&&i.toLowerCase()===t.toLowerCase()||t==null&&n.nodeType===3)return e.splice(r,1)[0]}return g._hydrationMismatch&&g._hydrationMismatch(t,e),null}function st(e){const t=e._children;if(t)for(let r=0;r<t.length;r++)t[r]&&(t[r]._parent=e)}function ie(e,t,r,n,i,s,o,c,p,a){const h=r&&r._children||[],b=t._children=[],v=[];Oe(n,v);for(let u=0;u<v.length;u++){let f=v[u];if(f==null||typeof f=="boolean"||typeof f=="function"){b[u]=null;continue}typeof f=="string"||typeof f=="number"||typeof f=="bigint"?f=z(null,f):Array.isArray(f)?f=z(re,null,f):(f._parent!=null||f._component!=null)&&(f=z(f.type,{...f.props,key:f.key,ref:f.ref})),f._parent=t,f._depth=(t._depth||0)+1,b[u]=f}const k=new Array(h.length),y=new Array(b.length);for(let u=0;u<b.length;u++){const f=b[u];if(f==null){y[u]=-1;continue}const m=ot(f,h,k,u);y[u]=m,m!==-1&&(k[m]=!0,f._index=m)}const T=[];for(let u=0;u<b.length;u++)b[u]!=null&&y[u]!==-1&&T.push(y[u]);const x=new Set(lt(T));for(let u=0;u<b.length;u++){const f=b[u];if(f==null)continue;const m=y[u];(m===-1||!x.has(m))&&(f._flags|=U)}for(let u=0;u<h.length;u++)!k[u]&&h[u]&&se(h[u],e,!1);for(let u=0;u<b.length;u++){const f=b[u];if(f==null)continue;const m=y[u]!==-1?h[y[u]]:null;Te(e,f,m,i,s,m?null:o,c,null,p&&!m,a)}for(let u=0;u<b.length;u++){const f=b[u];!f||!f._dom||(f._flags&U&&ct(e,f._dom,at(b,u+1)),f._flags&=~U)}}function Oe(e,t){if(!(e==null||typeof e=="boolean"))if(Array.isArray(e))for(let r=0;r<e.length;r++)Oe(e[r],t);else t.push(e)}function ot(e,t,r,n){const i=t[n];if(i&&!r[n]&&i.key==null&&e.key==null&&i.type===e.type)return n;for(let s=0;s<t.length;s++){const o=t[s];if(o&&!r[s]&&o.key===e.key&&o.type===e.type)return s}return-1}function lt(e){if(!e.length)return[];const t=[],r=new Array(e.length).fill(-1);for(let i=0;i<e.length;i++){let s=0,o=t.length;for(;s<o;){const c=s+o>>1;e[t[c]]<e[i]?s=c+1:o=c}s>0&&(r[i]=t[s-1]),t[s]=i}const n=[];for(let i=t[t.length-1];i!==-1;i=r[i])n.push(e[i]);return n.reverse()}function K(e){if(!e)return null;for(let t=0;t<e.length;t++){const r=e[t];if(!r||r.props&&r.props._parentDom)continue;if(r._dom)return r._dom;const n=K(r._children);if(n)return n}return null}function at(e,t){for(let r=t;r<e.length;r++){const n=e[r];if(!n||n._flags&U||n.props&&n.props._parentDom)continue;if(n._dom)return n._dom;const i=K(n._children);if(i)return i}return null}function ct(e,t,r){r==null?e.lastChild!==t&&e.appendChild(t):t.nextSibling!==r&&e.insertBefore(t,r)}function dt(e,t,r,n,i){for(const s in r)ke[s]||s in t||me(e,s,null,r[s],n);for(const s in t){if(ke[s]||s==="dangerouslySetInnerHTML")continue;const o=t[s];i&&(s==="value"||s==="checked")||r[s]!==o&&me(e,s,o,r[s],n)}}const ke={children:1,key:1,ref:1};function be(e,t,r){r==null||r===!1?e.removeAttribute(t):r===!0?e.setAttribute(t,""):e.setAttribute(t,r)}function me(e,t,r,n,i){if(t==="class"||t==="className"){e.setAttribute("class",r||"");return}if(t==="style"){pt(e,r,n);return}if(t[0]==="o"&&t[1]==="n"&&t.length>2){kt(e,t,r);return}if(i){const s=ut[t]||t;if(t==="xlinkHref"||s==="href"){r==null||r===!1?(e.removeAttributeNS("http://www.w3.org/1999/xlink","href"),e.removeAttribute("href")):e.setAttribute("href",r);return}be(e,s,r);return}if(t!=="href"&&t!=="list"&&t!=="form"&&t!=="tabIndex"&&t!=="download"&&t in e)try{e[t]=r??"";return}catch{}typeof r!="function"&&be(e,t,r)}const ut={strokeWidth:"stroke-width",strokeLinecap:"stroke-linecap",strokeLinejoin:"stroke-linejoin",fillOpacity:"fill-opacity",clipPath:"clip-path",fontSize:"font-size",stopColor:"stop-color",stopOpacity:"stop-opacity",xlinkHref:"href"},ft={animationIterationCount:1,columnCount:1,flexGrow:1,flexShrink:1,fontWeight:1,lineHeight:1,opacity:1,order:1,orphans:1,widows:1,zIndex:1,zoom:1};function pt(e,t,r){if(t==null){e.removeAttribute("style");return}if(typeof t=="string"){e.style.cssText=t;return}const n=e.style;if(r&&typeof r=="object")for(const i in r)t&&i in t||n.setProperty(ht(i),"");for(const i in t){const s=t[i],o=typeof s=="number"&&!ft[i]?s+"px":s==null?"":""+s;i.startsWith("--")?n.setProperty(i,o):n[i]=o}}function ht(e){return e.replace(/[A-Z]/g,t=>"-"+t.toLowerCase())}function kt(e,t,r){const n=e._listeners||(e._listeners={});let i=t.slice(2),s=!1;i.endsWith("Capture")&&(s=!0,i=i.slice(0,-7));const o=i.toLowerCase()+(s?"c":""),c=n[o];!c&&r?e.addEventListener(i.toLowerCase(),_e,s):c&&!r&&e.removeEventListener(i.toLowerCase(),_e,s),n[o]=r}function _e(e){this._listeners;const t=e.eventPhase===1,r=e.type+(t?"c":"");let n=this._listeners&&this._listeners[r];if(!n&&this._listeners&&(n=this._listeners[e.type]),n)return g.event&&(e=g.event(e)),n.call(this,e)}function se(e,t,r){if(!e)return;g.unmount&&g.unmount(e),e.ref&&Le(e.ref,null,e);const n=e._component;if(n){if(n.componentWillUnmount)try{n.componentWillUnmount()}catch(o){oe(o,e)}n._parentDom=null,n._vnode=null}const i=e._children;if(i){const o=r||typeof e.type=="string"||e.type==null;for(let c=0;c<i.length;c++)i[c]&&se(i[c],typeof e.type=="string"||e.type==null?e._dom:t,o)}const s=e._dom;e._dom=null,e._component=null,!r&&s&&s.parentNode&&(typeof e.type=="string"||e.type==null)&&s.remove()}function Le(e,t,r){try{if(typeof e=="function"){const n=e._cleanup;typeof n=="function"&&n();const i=e(t);e._cleanup=typeof i=="function"?i:null}else e&&(e.current=t)}catch(n){oe(n,r)}}function oe(e,t,r){for(let n=t;n;n=n._parent){const i=n._component;if(!i)continue;const s=i.constructor;let o=!1;if(s&&typeof s.getDerivedStateFromError=="function"){const c=s.getDerivedStateFromError(e);i._nextState=Object.assign({},i._nextState||i.state,c||{}),i._bits|=ne,i._bits&=~M,W(i),o=!0}if(typeof i.componentDidCatch=="function"&&(i.componentDidCatch(e),o=!0),o)return}if(g._catchError)g._catchError(e,t,r);else throw e}$e({patch:Te,applyRef:Le});const bt=new Map;function mt(e,t){{const r=document.querySelectorAll(e),n=[];for(let i=0;i<r.length;i++)n.push(t(r[i]));return n}}function _t(e,t,r,n){if(!document.querySelectorAll(e).length)throw new Error("Strike mount: host not found");return mt(e,i=>gt(i,t))}function gt(e,t,r,n){if(!e)throw new Error("Strike mount: host not found");let i=t;if(typeof i=="string"&&(i=bt.get(i),!i))throw new Error('Strike mount: unknown component "'+t+'"');const s=yt(e),o=z(i,s||{}),c=e.__root,p=e.getAttribute&&e.getAttribute("data-hydrate")!=null;return c?he(o,e):p?(Ye(o,e),e.removeAttribute&&e.removeAttribute("data-hydrate")):(e.textContent="",he(o,e)),e}function yt(e){if(!e.getAttribute)return null;const t=e.getAttribute("data-props");return t==null||t===""?null:JSON.parse(t)}function l(e,t,r){return r!==void 0&&(t=t==null?{key:r}:{...t,key:r}),_(e,t)}const S=l;let F=null,te=0;const ge=g._render;g._render=e=>{ge&&ge(e),F=e._component,te=0,F&&!F.__hooks&&(F.__hooks={_list:[]})};const ye=g._commit;g._commit=(e,t)=>{ye&&ye(e,t);for(let n=0;n<t.length;n++)ve(t[n],"_pendingLayout");const r=Ne(e);(g.requestAnimationFrame||xt)(()=>{for(const n of r)ve(n,"_pendingEffect")})};const xe=g.unmount;g.unmount=e=>{const t=e._component;if(t&&t.__hooks){const r=t.__hooks._list;for(let n=0;n<r.length;n++){const i=r[n];i._cleanup&&(i._cleanup(),i._cleanup=null)}}xe&&xe(e)};function Ne(e,t=[]){if(!e)return t;e._component&&e._component.__hooks&&t.push(e._component);const r=e._children;if(r)for(let n=0;n<r.length;n++)Ne(r[n],t);return t}function ve(e,t){const r=e.__hooks;if(r)for(let n=0;n<r._list.length;n++){const i=r._list[n],s=i[t];if(!s)continue;i._cleanup&&i._cleanup();const o=s();i._cleanup=typeof o=="function"?o:null,i[t]=null}}function xt(e){typeof requestAnimationFrame=="function"?requestAnimationFrame(e):setTimeout(e,0)}function le(e){const t=F.__hooks._list;return te>=t.length&&t.push(typeof e=="function"?e():e),t[te++]}function A(e){return wt(vt,e)}function vt(e,t){return typeof t=="function"?t(e):t}function wt(e,t,r){const n=le(()=>({_value:typeof t=="function"?t():t,_reducer:e}));n._reducer=e;const i=F,s=o=>{const c=n._reducer(n._value,o);c!==n._value&&(n._value=c,W(i))};return[n._value,s]}function St(e,t){if(e==null||t==null||e.length!==t.length)return!0;for(let r=0;r<e.length;r++)if(e[r]!==t[r])return!0;return!1}function Ct(e){return le(()=>({current:e}))}function At(e,t){const r=le(()=>({_deps:void 0,_value:void 0}));return St(r._deps,t)&&(r._deps=t,r._value=e()),r._value}function Rt(e,t){return At(()=>e,t)}const $=new WeakMap,Y=new Map;function w(e,...t){let r;if(typeof e=="string")r=e;else{r=e[0]||"";for(let i=0;i<t.length;i++)r+=t[i]+(e[i+1]||"");if($.has(e))return $.get(e)}if(Y.has(r)){const i=Y.get(r);return typeof e!="string"&&$.set(e,i),i}const n=document.createElement("style");return n.setAttribute("data-strike",""),n.textContent=r,document.head.appendChild(n),Y.set(r,n),typeof e!="string"&&$.set(e,n),n}function G(...e){let t="";for(let r=0;r<e.length;r++){const n=e[r];n==null||n===!1||n===""||(t=t?t+" "+n:""+n)}return t}w`
.strike-field__label,
.strike-select__label,
.strike-number__label,
.strike-autocomplete__label,
.strike-radio-group__label {
  font-size: 0.85rem;
  color: var(--strike-muted, #5c5c5c);
}
.strike-field__input,
.strike-select__control,
.strike-autocomplete__input,
.strike-number__input {
  font: inherit;
  padding: 0.45rem 0.65rem;
  border: 1px solid var(--strike-line, #ccc);
  border-radius: var(--strike-radius, 6px);
  background: #fff;
}
.strike-field__input:focus-visible,
.strike-select__control:focus-visible,
.strike-autocomplete__input:focus-visible,
.strike-number__input:focus-visible {
  outline: 2px solid var(--strike-accent, #0b6e4f);
  outline-offset: 1px;
}
.strike-field[data-state="invalid"] .strike-field__input,
.strike-select[data-state="invalid"] .strike-select__control,
.strike-autocomplete[data-state="invalid"] .strike-autocomplete__input,
.strike-number[data-state="invalid"] .strike-number__input {
  border-color: var(--strike-danger, #9b2226);
}
.strike-btn-group > .strike-btn,
.strike-toggle-group--joined > .strike-btn {
  border-radius: 0;
  margin-left: -1px;
}
.strike-btn-group > .strike-btn:first-child,
.strike-toggle-group--joined > .strike-btn:first-child {
  margin-left: 0;
  border-radius: var(--strike-radius, 6px) 0 0 var(--strike-radius, 6px);
}
.strike-btn-group > .strike-btn:last-child,
.strike-toggle-group--joined > .strike-btn:last-child {
  border-radius: 0 var(--strike-radius, 6px) var(--strike-radius, 6px) 0;
}
.strike-btn-group > .strike-btn:only-child,
.strike-toggle-group--joined > .strike-btn:only-child {
  border-radius: var(--strike-radius, 6px);
  margin-left: 0;
}
.strike-btn[data-state="busy"],
.strike-check[data-state="busy"],
.strike-switch[data-state="busy"],
.strike-btn-group[data-state="busy"],
.strike-toggle-group[data-state="busy"],
.strike-number[data-state="busy"],
.strike-autocomplete[data-state="busy"],
.strike-radio-group[data-state="busy"] {
  opacity: 0.6;
  pointer-events: none;
}
.strike-form[data-state="busy"] {
  opacity: 0.7;
  pointer-events: none;
}
`;w`
.strike-btn {
  font: inherit;
  cursor: pointer;
  border: 1px solid var(--strike-line, #ccc);
  background: #fff;
  color: inherit;
  border-radius: var(--strike-radius, 6px);
  padding: 0.4rem 0.85rem;
}
.strike-btn:hover { border-color: var(--strike-muted, #666); }
.strike-btn:focus-visible { outline: 2px solid var(--strike-accent, #0b6e4f); outline-offset: 2px; }
.strike-btn:active { transform: translateY(1px); }
.strike-btn--primary {
  background: var(--strike-accent, #0b6e4f);
  border-color: transparent;
  color: #fff;
}
.strike-btn--ghost { background: transparent; }
.strike-btn--default { background: #fff; }
.strike-btn--outline {
  background: transparent;
  border-color: var(--strike-accent, #0b6e4f);
  color: var(--strike-accent, #0b6e4f);
}
`;function O({variant:e="primary",state:t="rest",class:r,type:n="button",children:i,...s}){return _("button",{...s,type:n,class:G("strike-btn","strike-btn--"+e,r),"data-state":t},i)}w`
.strike-field { display: flex; flex-direction: column; gap: 0.35rem; font: inherit; }
`;function Q({label:e,id:t,class:r,state:n="rest",...i}){return _("label",{class:G("strike-field",r),"data-state":n},e&&_("span",{class:"strike-field__label"},e),_("input",{id:t,class:"strike-field__input",...i}))}w`
.strike-stack { display: flex; flex-direction: column; }
.strike-stack--row { flex-direction: row; align-items: center; }
`;function E({gap:e=8,row:t,class:r,children:n,style:i,...s}){return _("div",{...s,class:G("strike-stack",t&&"strike-stack--row",r),style:{gap:e,...i||{}}},n)}w`
.strike-text { margin: 0; font: inherit; color: inherit; }
.strike-text--muted { color: var(--strike-muted, #5c5c5c); }
.strike-text--danger { color: var(--strike-danger, #9b2226); }
.strike-text--title { font-size: 1.25rem; font-weight: 600; }
`;const Tt={p:1,span:1,h1:1,h2:1,h3:1,label:1,strong:1,em:1};function H({as:e="p",tone:t,class:r,children:n,...i}){const s=Tt[e]?e:"p";return _(s,{...i,class:G("strike-text",t&&"strike-text--"+t,r)},n)}w`
.strike-check {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font: inherit;
  cursor: pointer;
}
.strike-check__input {
  width: 1rem;
  height: 1rem;
  accent-color: var(--strike-accent, #0b6e4f);
}
`;w`
.strike-select { display: flex; flex-direction: column; gap: 0.35rem; font: inherit; }
`;w`
.strike-image {
  display: block;
  max-width: 100%;
  height: auto;
}
.strike-image--round { border-radius: var(--strike-radius, 6px); }
.strike-image--circle { border-radius: 50%; }
`;w`
.strike-form {
  display: flex;
  flex-direction: column;
  gap: var(--strike-gap, 0.75rem);
  font: inherit;
}
`;w`
.strike-switch {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font: inherit;
  cursor: pointer;
  position: relative;
}
.strike-switch__track {
  position: relative;
  width: 2.25rem;
  height: 1.25rem;
  border-radius: 999px;
  background: var(--strike-line, #ccc);
  transition: background 0.15s ease;
  flex-shrink: 0;
}
.strike-switch__track::after {
  content: "";
  position: absolute;
  top: 0.15rem;
  left: 0.15rem;
  width: 0.95rem;
  height: 0.95rem;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 2px rgba(0,0,0,0.2);
  transition: transform 0.15s ease;
}
.strike-switch__input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.strike-switch__input:focus-visible + .strike-switch__track {
  outline: 2px solid var(--strike-accent, #0b6e4f);
  outline-offset: 2px;
}
.strike-switch__input:checked + .strike-switch__track {
  background: var(--strike-accent, #0b6e4f);
}
.strike-switch__input:checked + .strike-switch__track::after {
  transform: translateX(1rem);
}
`;w`
.strike-dialog-root {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.4);
}
.strike-dialog {
  max-width: 24rem;
  width: 100%;
  max-height: min(90vh, 32rem);
  overflow: auto;
  background: #fff;
  color: var(--strike-ink, #1a1a1a);
  border-radius: var(--strike-radius, 6px);
  border: 1px solid var(--strike-line, #ccc);
  padding: 1rem 1.1rem;
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.18);
}
.strike-dialog__title {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  font-weight: 600;
}
`;w`
.strike-radio-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  font: inherit;
}
.strike-radio-group__options {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.strike-radio-group__options--row {
  flex-direction: row;
  flex-wrap: wrap;
  gap: 0.75rem;
}
.strike-radio {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font: inherit;
  cursor: pointer;
}
.strike-radio__input {
  width: 1rem;
  height: 1rem;
  accent-color: var(--strike-accent, #0b6e4f);
}
`;w`
.strike-number {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font: inherit;
}
.strike-number__row {
  display: inline-flex;
  align-items: stretch;
  gap: 0;
}
.strike-number__input {
  border-radius: 0;
  width: 4.5rem;
  text-align: center;
  -moz-appearance: textfield;
}
.strike-number__input::-webkit-outer-spin-button,
.strike-number__input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.strike-number__input:focus-visible { z-index: 1; }
.strike-number__row .strike-btn {
  border-radius: 0;
  padding: 0.4rem 0.65rem;
  min-width: 2.25rem;
}
.strike-number__row .strike-btn:first-child {
  border-radius: var(--strike-radius, 6px) 0 0 var(--strike-radius, 6px);
}
.strike-number__row .strike-btn:last-child {
  border-radius: 0 var(--strike-radius, 6px) var(--strike-radius, 6px) 0;
}
.strike-number__row .strike-btn + .strike-number__input { margin-left: -1px; }
.strike-number__row .strike-number__input + .strike-btn { margin-left: -1px; }
`;w`
.strike-btn-group {
  display: inline-flex;
  flex-direction: row;
  align-items: stretch;
  font: inherit;
}
`;w`
.strike-toggle-group {
  display: inline-flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 0.35rem;
  font: inherit;
}
.strike-toggle-group--joined { gap: 0; }
`;w`
.strike-autocomplete {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font: inherit;
  position: relative;
}
.strike-autocomplete__input {
  width: 100%;
  box-sizing: border-box;
}
.strike-autocomplete__list {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  z-index: 20;
  margin: 0.2rem 0 0;
  padding: 0.25rem 0;
  list-style: none;
  background: #fff;
  border: 1px solid var(--strike-line, #ccc);
  border-radius: var(--strike-radius, 6px);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.12);
  max-height: 12rem;
  overflow: auto;
}
.strike-autocomplete__option {
  padding: 0.4rem 0.65rem;
  cursor: pointer;
}
.strike-autocomplete__option[aria-selected="true"],
.strike-autocomplete__option:hover {
  background: color-mix(in srgb, var(--strike-accent, #0b6e4f) 12%, #fff);
}
`;function P(...e){let t="";for(let r=0;r<e.length;r++){const n=e[r];n==null||n===!1||n===""||(t=t?t+" "+n:""+n)}return t}w`
.strike-appbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: var(--strike-appbar-height, 3rem);
  padding: 0 1rem;
  background: #fff;
  color: var(--strike-ink, #1a1a1a);
  border-bottom: 1px solid var(--strike-line, #d4d4d4);
  box-sizing: border-box;
}
.strike-appbar--elevated {
  box-shadow: var(--strike-elev-1, 0 1px 2px rgba(0,0,0,.08));
  border-bottom-color: transparent;
}
.strike-appbar--sticky { position: sticky; top: 0; z-index: 30; }
.strike-appbar--fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
}
.strike-appbar--dense { min-height: 2.5rem; padding: 0 0.75rem; }
.strike-appbar--inverse {
  background: var(--strike-ink, #1a1a1a);
  color: #fff;
  border-bottom-color: transparent;
}
.strike-appbar__start,
.strike-appbar__end { display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
.strike-appbar__main { flex: 1; min-width: 0; display: flex; align-items: center; gap: 0.75rem; }
`;function Ot({position:e="static",elevated:t,start:r,end:n,dense:i,tone:s="default",class:o,children:c,...p}){return _("header",{...p,class:P("strike-appbar",t&&"strike-appbar--elevated",e==="sticky"&&"strike-appbar--sticky",e==="fixed"&&"strike-appbar--fixed",i&&"strike-appbar--dense",s==="inverse"&&"strike-appbar--inverse",o)},r&&_("div",{class:"strike-appbar__start"},r),_("div",{class:"strike-appbar__main"},c),n&&_("div",{class:"strike-appbar__end"},n))}function Lt(e,t,r){const n=e!==void 0,[i,s]=A(n?e:t),o=n?e:i;function c(p){const a=typeof p=="function"?p(o):p;n||s(a),r&&r(a)}return[o,c]}w`
.strike-tabs { display: flex; flex-direction: column; gap: 0.75rem; }
.strike-tabs--vertical { flex-direction: row; align-items: flex-start; }
.strike-tabs__list {
  display: flex;
  gap: 0.25rem;
  border-bottom: 1px solid var(--strike-line, #d4d4d4);
}
.strike-tabs--vertical .strike-tabs__list {
  flex-direction: column;
  border-bottom: 0;
  border-right: 1px solid var(--strike-line, #d4d4d4);
}
.strike-tabs__tab {
  font: inherit;
  padding: 0.5rem 0.85rem;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  cursor: pointer;
  color: var(--strike-muted, #5c5c5c);
  margin-bottom: -1px;
}
.strike-tabs--vertical .strike-tabs__tab {
  border-bottom: 0;
  border-right: 2px solid transparent;
  margin-bottom: 0;
  margin-right: -1px;
}
.strike-tabs__tab[aria-selected="true"] {
  color: var(--strike-accent, #0b6e4f);
  border-color: var(--strike-accent, #0b6e4f);
  font-weight: 600;
}
.strike-tabs__tab:focus-visible {
  outline: 2px solid var(--strike-accent, #0b6e4f);
  outline-offset: 2px;
}
.strike-tabs__panel { min-width: 0; flex: 1; }
`;function Nt({value:e,defaultValue:t,onChange:r,items:n=[],orientation:i="horizontal",class:s,...o}){const c=n[0]&&n[0].id,[p,a]=Lt(e,t??c,r),h=Ct(null);function b(k){const y=n.map(function(m){return m.id}),T=y.indexOf(p);if(T<0)return;const x=i==="vertical";let u=T;if(!x&&k.key==="ArrowRight"||x&&k.key==="ArrowDown")u=(T+1)%y.length,k.preventDefault();else if(!x&&k.key==="ArrowLeft"||x&&k.key==="ArrowUp")u=(T-1+y.length)%y.length,k.preventDefault();else if(k.key==="Home")u=0,k.preventDefault();else if(k.key==="End")u=y.length-1,k.preventDefault();else return;a(y[u]);const f=h.current&&h.current.querySelectorAll('[role="tab"]');f&&f[u]&&f[u].focus()}const v=n.find(function(k){return k.id===p})||n[0];return _("div",{...o,class:P("strike-tabs",i==="vertical"&&"strike-tabs--vertical",s)},_("div",{ref:h,role:"tablist","aria-orientation":i,class:"strike-tabs__list",onKeyDown:b},n.map(function(k){const y=k.id===(v&&v.id);return _("button",{key:k.id,type:"button",role:"tab",id:"strike-tab-"+k.id,class:"strike-tabs__tab","aria-selected":y,"aria-controls":"strike-tabpanel-"+k.id,tabIndex:y?0:-1,onClick:function(){a(k.id)}},k.label)})),v&&_("div",{role:"tabpanel",id:"strike-tabpanel-"+v.id,"aria-labelledby":"strike-tab-"+v.id,class:"strike-tabs__panel",tabIndex:0},v.panel))}function Dt(e,t){return!t||t==="default"?null:e+"--"+t}w`
.strike-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--strike-icon-size, 1.25rem);
  height: var(--strike-icon-size, 1.25rem);
  flex-shrink: 0;
  line-height: 0;
  color: inherit;
}
.strike-icon svg {
  width: 100%;
  height: 100%;
  fill: currentColor;
}
.strike-icon--sm { --strike-icon-size: 1rem; }
.strike-icon--md { --strike-icon-size: 1.25rem; }
.strike-icon--lg { --strike-icon-size: 1.75rem; }
`;function Et({size:e="md",label:t,class:r,children:n,style:i,...s}){const o=typeof e=="number"?null:e==="sm"||e==="lg"?"strike-icon--"+e:"strike-icon--md",c={...i||{}};return typeof e=="number"&&(c.width=e+"px",c.height=e+"px"),_("span",{...s,...t?{role:"img","aria-label":t}:{"aria-hidden":"true"},class:P("strike-icon",o,r),style:c},n)}w`
.strike-alert {
  padding: 0.75rem 0.9rem;
  border-radius: var(--strike-radius, 6px);
  border: 1px solid var(--strike-line, #d4d4d4);
  background: var(--strike-fill, #f6f6f4);
}
.strike-alert--accent {
  border-color: var(--strike-accent, #0b6e4f);
  background: color-mix(in srgb, var(--strike-accent, #0b6e4f) 10%, #fff);
}
.strike-alert--danger {
  border-color: var(--strike-danger, #9b2226);
  background: color-mix(in srgb, var(--strike-danger, #9b2226) 10%, #fff);
}
.strike-alert--ok {
  border-color: var(--strike-ok, #2d6a4f);
  background: color-mix(in srgb, var(--strike-ok, #2d6a4f) 10%, #fff);
}
.strike-alert--warn {
  border-color: var(--strike-warn, #b08900);
  background: color-mix(in srgb, var(--strike-warn, #b08900) 12%, #fff);
}
.strike-alert--info {
  border-color: var(--strike-info, #1d4e89);
  background: color-mix(in srgb, var(--strike-info, #1d4e89) 10%, #fff);
}
.strike-alert--muted {
  border-color: var(--strike-line, #d4d4d4);
  color: var(--strike-muted, #5c5c5c);
}
.strike-alert__body { flex: 1; min-width: 0; }
.strike-alert__title { margin: 0 0 0.2rem; font-weight: 600; }
`;function It(e){return _(Et,{size:"sm",label:void 0},e==="danger"?"!":e==="ok"?"ok":e==="warn"?"!":"i")}function jt({tone:e="default",title:t,children:r,icon:n,onClose:i,action:s,assertive:o,class:c,...p}){const a=n!==!1,h=n===!1?null:n??It(e);return _("div",{...p,role:o?"alert":"status",class:P("strike-alert",Dt("strike-alert",e),c)},_(E,{row:!0,gap:10,class:"strike-alert__row"},a&&h,_("div",{class:"strike-alert__body"},t&&_(H,{as:"p",class:"strike-alert__title"},t),r&&_(H,{as:"p"},r)),s,i&&_(O,{variant:"ghost",type:"button","aria-label":"Close",onClick:i},"x")))}w`
.strike-box { box-sizing: border-box; }
.strike-box--flex { display: flex; }
.strike-box--grid { display: grid; }
.strike-box--inline-flex { display: inline-flex; }
.strike-box--block { display: block; }
.strike-box--none { display: none; }
`;const Mt={div:1,section:1,main:1,span:1,article:1,aside:1,header:1,footer:1,nav:1};function N(e){if(e!=null)return typeof e=="number"?"calc("+e+" * var(--strike-space, 0.75rem))":e}const Pt={flex:"strike-box--flex",grid:"strike-box--grid","inline-flex":"strike-box--inline-flex",block:"strike-box--block",none:"strike-box--none"};function zt({as:e="div",p:t,px:r,py:n,m:i,mx:s,my:o,gap:c,display:p,direction:a,align:h,justify:b,width:v,height:k,class:y,style:T,children:x,...u}){const f=Mt[e]?e:"div",m=Object.assign({},T||{});return t!=null&&(m.padding=N(t)),r!=null&&(m.paddingLeft=N(r),m.paddingRight=N(r)),n!=null&&(m.paddingTop=N(n),m.paddingBottom=N(n)),i!=null&&(m.margin=N(i)),s!=null&&(m.marginLeft=N(s),m.marginRight=N(s)),o!=null&&(m.marginTop=N(o),m.marginBottom=N(o)),c!=null&&(m.gap=N(c)),a&&(m.flexDirection=a),h&&(m.alignItems=h),b&&(m.justifyContent=b),v!=null&&(m.width=v),k!=null&&(m.height=k),_(f,{...u,class:P("strike-box",Pt[p]||null,y),style:m},x)}w`
.strike-container {
  width: 100%;
  margin-left: auto;
  margin-right: auto;
  padding-left: var(--strike-space, 0.75rem);
  padding-right: var(--strike-space, 0.75rem);
  box-sizing: border-box;
}
.strike-container--sm { max-width: 40rem; }
.strike-container--md { max-width: 48rem; }
.strike-container--lg { max-width: 64rem; }
.strike-container--xl { max-width: 80rem; }
`;function Ft({size:e="md",class:t,children:r,...n}){return _(zt,{...n,class:P("strike-container","strike-container--"+e,t)},r)}w`
.strike-paper {
  background: #fff;
  color: var(--strike-ink, #1a1a1a);
  border-radius: var(--strike-radius, 6px);
}
.strike-paper--elevated { box-shadow: var(--strike-elev-1, 0 1px 2px rgba(0,0,0,.08)); }
.strike-paper--elev-2 { box-shadow: var(--strike-elev-2, 0 4px 16px rgba(0,0,0,.12)); }
.strike-paper--elev-3 {
  box-shadow: var(--strike-elev-2, 0 4px 16px rgba(0,0,0,.12)), 0 8px 28px rgba(0,0,0,.1);
}
.strike-paper--outlined {
  border: 1px solid var(--strike-line, #d4d4d4);
  box-shadow: none;
}
.strike-paper--filled {
  background: var(--strike-fill, #f6f6f4);
  box-shadow: none;
}
.strike-paper--square { border-radius: 0; }
`;function Ht({elevation:e=0,variant:t="elevated",square:r,class:n,children:i,...s}){const o=t==="elevated"&&e>0?e>=3?"strike-paper--elev-3":e===2?"strike-paper--elev-2":"strike-paper--elevated":t==="elevated"&&e===0?"strike-paper--elevated":null;return _("div",{...s,class:P("strike-paper",t!=="elevated"&&"strike-paper--"+t,o,r&&"strike-paper--square",n)},i)}w`
.strike-card { overflow: hidden; }
.strike-card__media { display: block; width: 100%; }
.strike-card__media > * { display: block; width: 100%; }
.strike-card__body { padding: 0.85rem 1rem; }
.strike-card__actions {
  display: flex;
  gap: 0.5rem;
  padding: 0 1rem 0.85rem;
  flex-wrap: wrap;
}
.strike-card__subtitle {
  margin: 0.15rem 0 0;
  color: var(--strike-muted, #5c5c5c);
  font-size: 0.9rem;
}
`;function X({elevated:e,outlined:t,filled:r,title:n,subtitle:i,media:s,actions:o,class:c,children:p,...a}){const h=t?"outlined":r?"filled":"elevated";let b=0;return h==="elevated"&&(b=e===!1?0:typeof e=="number"?e:1),_(Ht,{...a,variant:h,elevation:b,class:P("strike-card",c)},s&&_("div",{class:"strike-card__media"},s),_("div",{class:"strike-card__body"},(n||i)&&_(E,{gap:4},n&&_(H,{as:"h3",tone:"title"},n),i&&_(H,{as:"p",class:"strike-card__subtitle"},i)),p),o&&_("div",{class:"strike-card__actions"},o))}async function D(e){const t=await e.text();let r=null;if(t)try{r=JSON.parse(t)}catch{r=t}if(!e.ok){const n=new Error(typeof r=="object"&&r&&r.error?String(r.error):`HTTP ${e.status}`);throw n.status=e.status,n.body=r,n}return r}function Jt(e){if(!e||typeof e!="object")return[];const t=e.participants;return t?Array.isArray(t)?t.map((r,n)=>{if(r==null||typeof r!="object")return{id:String(r),role:"",meta:""};const i=r.participant_id||r.id||`row-${n}`,s=r.role!=null?String(r.role):"",{participant_id:o,id:c,role:p,...a}=r,h=Object.keys(a).length?JSON.stringify(a):"";return{id:String(i),role:s,meta:h}}):typeof t=="object"?Object.keys(t).sort().map(r=>{const n=t[r];let i="",s="";if(n!=null&&typeof n=="object"){n.role!=null&&(i=String(n.role));const{role:o,...c}=n;Object.keys(c).length&&(s=JSON.stringify(c))}else n!=null&&(s=String(n));return{id:r,role:i,meta:s}}):[]:[]}async function $t(){return D(await fetch("/v1/sessions"))}async function Ut(e){return D(await fetch(`/v1/sessions/${encodeURIComponent(e)}/presence`))}async function qt(e,t){return D(await fetch(`/v1/sessions/${encodeURIComponent(e)}/moderate`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t)}))}async function Bt(e,t){return D(await fetch(`/v1/sessions/${encodeURIComponent(e)}/record/start`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t??{})}))}async function Wt(e,t){return D(await fetch(`/v1/sessions/${encodeURIComponent(e)}/record/stop`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t??{})}))}async function Kt(e){return D(await fetch(`/v1/sessions/${encodeURIComponent(e)}/record`))}async function Gt(e=50){return D(await fetch(`/v1/history/sessions?limit=${e}`))}async function Vt(e=!1){return D(await fetch(`/v1/fleet/nodes${e?"?refresh=1":""}`))}async function Yt(e=!1){return D(await fetch(`/v1/fleet/summary${e?"?refresh=1":""}`))}async function we(e,t){return D(await fetch("/v1/fleet/accepting",{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({node:e,accepting:t})}))}const Qt=[{id:"live",label:"Live"},{id:"history",label:"History"},{id:"fleet",label:"Fleet"}],Xt=`{
  "action": "kick",
  "participant_id": "p2"
}`,Zt=`{
  "mode": "audio"
}`;function er(e){if(e==null)return"(empty)";if(typeof e=="string")return e;try{return JSON.stringify(e,null,2)}catch{return String(e)}}function Z(e){if(e==null)return"-";try{return new Date(Number(e)*1e3).toISOString()}catch{return String(e)}}function tr(){const[e,t]=A("live"),[r,n]=A("Idle"),[i,s]=A("muted"),[o,c]=A("room-1"),[p,a]=A([]),[h,b]=A(Xt),[v,k]=A(Zt),[y,T]=A(""),[x,u]=A(!1),[f,m]=A([]),[ae,ce]=A([]),[de,De]=A([]),[I,Ee]=A(null),V=Rt((d,C="info")=>{n(d),s(C)},[]);function ue(d,C){const R=`[${new Date().toISOString()}] ${d}: ${C}`;T(j=>j?j+`

`+R:R)}async function L(d,C){u(!0);try{const R=await C();return ue(d,er(R)),V(`${d} ok`,"ok"),R}catch(R){const j=String(R&&R.message?R.message:R);return ue(`${d} error`,j),V(j,"danger"),null}finally{u(!1)}}async function Ie(){const d=await L("sessions",()=>$t()),C=d&&d.sessions||[];a(Array.isArray(C)?C:[])}async function je(){const d=await L("presence",()=>Ut(o.trim()));ce(Jt(d))}async function Me(){const d=await L("history",()=>Gt()),C=d&&d.sessions||[];m(Array.isArray(C)?C:[])}async function J(d){const C=await L("fleet nodes",()=>Vt(!!d)),R=C&&C.nodes||[];De(Array.isArray(R)?R:[]);const j=await L("fleet summary",()=>Yt(!1));Ee(j&&j.summary||null)}async function Pe(d){await L("fleet drain",()=>we(d,!1)),await J(!0)}async function ze(d){await L("fleet undrain",()=>we(d,!0)),await J(!0)}function fe(d){c(d),t("live"),ce([]),V(`Selected ${d}`,"info")}function Fe(d){b(JSON.stringify({action:"kick",participant_id:d},null,2))}return S("div",{class:"admin-shell",children:[l(Ot,{elevated:!0,position:"sticky",start:l("span",{class:"admin-brand-mark","aria-hidden":"true"}),children:S("div",{children:[l("p",{class:"admin-brand",children:"Xylon Admin"}),l("p",{class:"admin-tagline",children:"Demo console | Live | History | Fleet"})]})}),S(Ft,{class:"admin-main",size:"xl",children:[l(H,{as:"p",class:"admin-hint",children:"Browser talks only to this admin process. Live routes proxy through the platform interface; History and Fleet read local SQLite (Fleet poller refreshes node snapshots)."}),l(Nt,{value:e,onChange:t,items:Qt.map(d=>({id:d.id,label:d.label,panel:null}))}),S("div",{class:"admin-tab-body",children:[l("div",{hidden:e!=="live",children:l(X,{outlined:!0,title:"Live room",subtitle:"Presence, moderate, record",children:S(E,{gap:12,children:[l("div",{class:"admin-panel-grid admin-panel-grid--2",children:l(Q,{label:"Session id",value:o,onInput:d=>c(d.target.value)})}),l(E,{row:!0,gap:8,children:l(O,{type:"button",disabled:x,onClick:Ie,children:"Refresh live list"})}),p.length>0?l("div",{class:"admin-live-list",children:p.map(d=>{const C=d.session_id||d.id||String(d);return l("button",{type:"button",onClick:()=>fe(C),children:C},C)})}):null,S(E,{row:!0,gap:8,children:[l(O,{type:"button",disabled:x||!o.trim(),onClick:je,children:"Refresh presence"}),l(O,{type:"button",disabled:x||!o.trim(),onClick:()=>L("record status",()=>Kt(o.trim())),children:"Record status"})]}),l("div",{class:"admin-roster",children:S("table",{class:"admin-history-table",children:[l("thead",{children:S("tr",{children:[l("th",{children:"Participant"}),l("th",{children:"Role"}),l("th",{children:"Meta"}),l("th",{})]})}),l("tbody",{children:ae.length===0?l("tr",{children:l("td",{colspan:"4",children:"No roster yet. Refresh presence for this session."})}):ae.map(d=>S("tr",{children:[l("td",{children:d.id}),l("td",{children:d.role||"-"}),l("td",{children:d.meta||"-"}),l("td",{children:l("button",{type:"button",onClick:()=>Fe(d.id),children:"Kick"})})]},d.id))})]})}),l(Q,{label:"Moderate JSON",as:"textarea",rows:3,value:h,onInput:d=>b(d.target.value)}),l(O,{type:"button",disabled:x||!o.trim(),onClick:()=>L("moderate",()=>qt(o.trim(),JSON.parse(h))),children:"Moderate"}),l(Q,{label:"Record start JSON",as:"textarea",rows:3,value:v,onInput:d=>k(d.target.value)}),S(E,{row:!0,gap:8,children:[l(O,{type:"button",disabled:x||!o.trim(),onClick:()=>L("record start",()=>Bt(o.trim(),JSON.parse(v))),children:"Record start"}),l(O,{type:"button",disabled:x||!o.trim(),onClick:()=>L("record stop",()=>Wt(o.trim(),{})),children:"Record stop"})]}),l("pre",{class:"admin-log",children:y||"Results appear here."})]})})}),l("div",{hidden:e!=="history",children:l(X,{outlined:!0,title:"Past sessions",subtitle:"SQLite on this admin process",children:S(E,{gap:12,children:[l(O,{type:"button",disabled:x,onClick:Me,children:"Refresh history"}),l("div",{class:"admin-history",children:S("table",{class:"admin-history-table",children:[l("thead",{children:S("tr",{children:[l("th",{children:"Session"}),l("th",{children:"Created"}),l("th",{children:"Closed"}),l("th",{children:"Region"})]})}),l("tbody",{children:f.length===0?l("tr",{children:l("td",{colspan:"4",children:"No rows yet. Refresh after webhooks arrive."})}):f.map(d=>S("tr",{children:[l("td",{children:l("button",{type:"button",onClick:()=>fe(d.session_id),children:d.session_id})}),l("td",{children:Z(d.created_at)}),l("td",{children:Z(d.closed_at)}),l("td",{children:d.last_region||"-"})]},d.session_id))})]})})]})})}),l("div",{hidden:e!=="fleet",children:l(X,{outlined:!0,title:"Fleet",subtitle:"Node snapshots from SQLite (polled via interface)",children:S(E,{gap:12,children:[S(E,{row:!0,gap:8,children:[l(O,{type:"button",disabled:x,onClick:()=>J(!1),children:"Refresh cache"}),l(O,{type:"button",disabled:x,onClick:()=>J(!0),children:"Refresh now"})]}),I?S(H,{as:"p",class:"admin-hint",children:["nodes=",I.node_count??0," healthy=",I.healthy_count??0," unhealthy=",I.unhealthy_count??0," sessions=",I.session_count??0," participants=",I.participant_count??0," regions=",(I.regions||[]).join(",")||"-"," hub=",I.hub_node||"-"]}):null,l("div",{class:"admin-history",children:S("table",{class:"admin-history-table",children:[l("thead",{children:S("tr",{children:[l("th",{children:"Node"}),l("th",{children:"Region"}),l("th",{children:"Sessions"}),l("th",{children:"Participants"}),l("th",{children:"Hub"}),l("th",{children:"Cohost hubs"}),l("th",{children:"Reachable"}),l("th",{children:"Accepting"}),l("th",{children:"Polled"}),l("th",{})]})}),l("tbody",{children:de.length===0?l("tr",{children:l("td",{colspan:"10",children:"No snapshots yet. Refresh now or wait for the poller."})}):de.map(d=>S("tr",{children:[l("td",{children:d.node}),l("td",{children:d.region||"-"}),l("td",{children:d.session_count??0}),l("td",{children:d.participant_count??0}),l("td",{children:d.hub?"yes":"no"}),l("td",{children:d.cohost_hub_count??0}),l("td",{children:d.reachable===!1?"no":"yes"}),l("td",{children:d.accepting===!1?"no":"yes"}),l("td",{children:Z(d.polled_at)}),l("td",{children:d.accepting===!1?l(O,{size:"sm",disabled:x,onClick:()=>ze(d.node),children:"Undrain"}):l(O,{size:"sm",tone:"danger",disabled:x,onClick:()=>Pe(d.node),children:"Drain"})})]},d.node))})]})})]})})})]}),l(jt,{class:"admin-status",tone:i,title:"Status",children:r})]})]})}_t("#app",tr);
