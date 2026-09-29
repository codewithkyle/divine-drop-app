// Where images and symbols are served from. The layout writes it into a meta
// tag so this stays in step with the server's CDN_URL; the fallback only covers
// that tag going missing, and names the Cloudflare origin rather than the
// DigitalOcean host being decommissioned.
const CDN_URL: string =
    document.querySelector<HTMLMetaElement>('meta[name="cdn-url"]')?.content?.replace(/\/+$/, "") ||
    "https://cdn.divinedrop.app";

class CardText extends HTMLElement {
    constructor(){
        super();
    }

    connectedCallback(){
        this.parse();
    }

    parse() {
        let str = this.innerHTML;
        const segments = str.match(/\{.*?\}/g);
        if (segments == null) return;
        for (let i = 0; i < segments.length; i++){
            // Strip the separators too, not just the braces: hybrid and Phyrexian
            // mana are stored flattened, so {W/U} is WU.svg and {G/U/P} is GUP.svg.
            // Keeping the slash asked for symbols/W/U.svg, which has never existed.
            const symbol = segments[i].replace(/[{}\/]/g, "");
            const url = `${CDN_URL}/symbols/${symbol}.svg`;
            str = str.replace(segments[i], `<img src=\"${url}\">`);
        }
        this.innerHTML = str;
    }
}
if (!customElements.get("card-text")) customElements.define("card-text", CardText);
