class CardPreviewButton extends HTMLElement{
    private card: HTMLImageElement | null;

    constructor(){
        super();
        this.card = null;
    }

    connectedCallback(){
        this.addEventListener("mouseenter", this.onMouseEnter);
        this.addEventListener("focus", this.onMouseEnter);
        this.addEventListener("mouseleave", this.onMouseLeave);
        this.addEventListener("blur", this.onMouseLeave);
    }

    private onMouseEnter = () => {
        if (this.card && this.card.isConnected){
            this.card.remove();
        }
        this.card = document.createElement("img");
        this.card.src = this.dataset.cardUrl || "";
        if (!this.card.src) return;
        const bounds = this.getBoundingClientRect();
        this.card.style.position = "fixed";
        let bottom = bounds.bottom;
        if (bottom + 488 > window.innerHeight){
            bottom = window.innerHeight - 488;
        }
        this.card.style.top = `${bottom}px`;
        this.card.style.left = `${bounds.left - 350}px`;
        this.card.style.width = "350px";
        this.card.style.boxShadow = "var(--shadow-black-lg)";
        this.card.style.borderRadius = "4%";
        this.card.style.zIndex = "1000";
        this.card.style.opacity = "0";
        this.card.style.transition = "opacity 150ms var(--ease-in-out)";
        this.card.style.pointerEvents = "none";
        this.card.addEventListener("load", () => {
            if (this.card){
                this.card.style.opacity = "1";
            }
        });
        document.body.appendChild(this.card);
    }

    private onMouseLeave = () => {
        if (this.card && this.card.isConnected){
            this.card.remove();
        }
    }
}
if (!customElements.get("card-preview")) customElements.define("card-preview", CardPreviewButton);
