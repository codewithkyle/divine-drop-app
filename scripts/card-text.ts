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
            const symbol = segments[i].replace(/\{|\}/g, "");
            const url = `https://divinedrop.nyc3.cdn.digitaloceanspaces.com/symbols/${symbol}.svg`;
            str = str.replace(segments[i], `<img src=\"${url}\">`);
        }
        this.innerHTML = str;
    }
}
if (!customElements.get("card-text")) customElements.define("card-text", CardText);
