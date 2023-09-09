class DeckTray extends HTMLElement {
    private mainEl: HTMLElement | null;

    constructor() {
        super();
        this.mainEl = document.querySelector("main");
    }

    connectedCallback() {
        this.mainEl?.addEventListener("scroll", this.handleScroll.bind(this), { passive: true });
        this.setHeight();
    }

    private handleScroll = () => {
       this.setHeight();
    }

    private setHeight() {
        // Header: 128px
        if (!this.mainEl) return;
        const height = window.innerHeight - (16 * 4) - 128 + this.mainEl.scrollTop;
        const maxHeight = window.innerHeight - (16 * 4);
        this.style.height = `${height <= maxHeight ? height : maxHeight}px`;
    }
}
if (!customElements.get("deck-tray")) customElements.define("deck-tray", DeckTray);
